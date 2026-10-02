// Package inline implements mdp's inline phase: text becomes
// emphasis/strong/code/link nodes with sanitized URLs. Parsing never fails;
// malformed input degrades to literal text and reports a diagnostic.
package inline

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/ozguragain/mdp/ast"
	"github.com/ozguragain/mdp/diag"
)

// Parse splits raw inline text into nodes and diagnostics. Spans are
// relative to text; inline.Process remaps them onto the document.
func Parse(text string) ([]ast.InlineNode, []diag.Diagnostic) {
	p := &inlineParser{}
	nodes := p.parse(text, mapperFor(text, ast.Pos{Offset: 0, Line: 1, Column: 1}), false)
	return nodes, p.diags
}

// Process parses each block's raw text in place (call exactly once, before
// codegen) and returns its diagnostics in source order.
func Process(doc *ast.DocumentNode) []diag.Diagnostic {
	var diags []diag.Diagnostic
	processBlocks(doc.Blocks, &diags)
	return diags
}

func processBlocks(blocks []ast.BlockNode, diags *[]diag.Diagnostic) {
	for _, block := range blocks {
		switch n := block.(type) {
		case *ast.HeadingNode:
			n.Inlines = processInlines(n.Inlines, diags)
		case *ast.ParagraphNode:
			n.Inlines = processInlines(n.Inlines, diags)
		case *ast.BlockquoteNode:
			processBlocks(n.Blocks, diags)
		case *ast.ListNode:
			for _, item := range n.Items {
				processBlocks(item.Blocks, diags)
			}
		case *ast.ListItemNode:
			processBlocks(n.Blocks, diags)
		}
	}
}

func processInlines(inlines []ast.InlineNode, diags *[]diag.Diagnostic) []ast.InlineNode {
	if len(inlines) == 0 {
		return nil
	}
	parts := make([]string, 0, len(inlines))
	bases := make([]ast.Pos, 0, len(inlines))
	for _, in := range inlines {
		t, ok := in.(*ast.TextNode)
		if !ok {
			return inlines
		}
		parts = append(parts, t.Value)
		bases = append(bases, t.Span.Start)
	}
	nodes, ds := Parse(strings.Join(parts, "\n"))
	ast.RemapInlines(nodes, bases)
	for _, d := range ds {
		d.Span = ast.RemapSpan(d.Span, bases)
		*diags = append(*diags, d)
	}
	return nodes
}

// spanMapper maps fragment-local byte offsets to positions in the text the
// fragment came from; sub-mappers keep recursive parses in the same space.
type spanMapper struct {
	base   ast.Pos // position of the fragment's first byte
	starts []int   // byte offsets of the fragment's line starts
}

func mapperFor(text string, base ast.Pos) spanMapper {
	starts := []int{0}
	for i := 0; i < len(text); i++ {
		if text[i] == '\n' {
			starts = append(starts, i+1)
		}
	}
	return spanMapper{base: base, starts: starts}
}

func (m spanMapper) pos(local int) ast.Pos {
	li := len(m.starts) - 1
	for li > 0 && m.starts[li] > local {
		li--
	}
	if li == 0 {
		return ast.Pos{Offset: m.base.Offset + local, Line: m.base.Line, Column: m.base.Column + local}
	}
	return ast.Pos{
		Offset: m.base.Offset + local,
		Line:   m.base.Line + li,
		Column: local - m.starts[li] + 1,
	}
}

// span maps a local byte range to a half-open span.
func (m spanMapper) span(start, end int) ast.Span {
	return ast.Span{Start: m.pos(start), End: m.pos(end)}
}

// sub maps a verbatim sub-fragment of this fragment.
func (m spanMapper) sub(localStart int) spanMapper {
	out := spanMapper{base: m.pos(localStart), starts: []int{0}}
	for _, s := range m.starts {
		if s > localStart {
			out.starts = append(out.starts, s-localStart)
		}
	}
	return out
}

type inlineParser struct {
	diags []diag.Diagnostic
}

// parse scans s left to right — first matching construct at each position
// wins, with recursive content. noLinks suppresses link syntax in labels.
func (p *inlineParser) parse(s string, m spanMapper, noLinks bool) []ast.InlineNode {
	var nodes []ast.InlineNode
	var buf strings.Builder
	textStart := 0

	flush := func(end int) {
		if buf.Len() > 0 {
			nodes = append(nodes, &ast.TextNode{Value: buf.String(), Span: m.span(textStart, end)})
			buf.Reset()
		}
	}
	appendText := func(start, end int) {
		if buf.Len() == 0 {
			textStart = start
		}
		buf.WriteString(s[start:end])
	}

	i := 0
	for i < len(s) {
		switch c := s[i]; c {
		case '\\':
			if buf.Len() == 0 {
				textStart = i
			}
			if i+1 < len(s) && isASCIIPunct(s[i+1]) {
				buf.WriteByte(s[i+1])
				i += 2
				continue
			}
			buf.WriteByte('\\')
			i++
		case '`':
			n := backtickRunAt(s, i)
			if close := findBacktickClose(s, i+n, n); close >= 0 {
				flush(i)
				nodes = append(nodes, &ast.CodeNode{
					Value: stripCodeSpaces(s[i+n : close]),
					Span:  m.span(i, close+n),
				})
				i = close + n
				continue
			}
			p.diags = append(p.diags, diag.Diagnostic{
				Kind:    diag.UnclosedCodeSpan,
				Message: fmt.Sprintf("code span opened with %q is never closed; kept as literal text", strings.Repeat("`", n)),
				Span:    m.span(i, i+n),
			})
			appendText(i, i+n)
			i += n
		case '[':
			if !noLinks {
				lm := matchLink(s, i)
				if lm.ok {
					children := p.parse(s[lm.labelStart:lm.labelEnd], m.sub(lm.labelStart), true)
					if isSafeURL(lm.dest) {
						flush(i)
						nodes = append(nodes, &ast.LinkNode{
							Destination: lm.dest,
							Title:       lm.title,
							Children:    children,
							Span:        m.span(i, lm.next),
						})
					} else {
						// Unsafe URL: no <a> at all, only the label content.
						p.diags = append(p.diags, diag.Diagnostic{
							Kind:    diag.UnsafeURL,
							Message: fmt.Sprintf("dropped href with unsafe URL %q; link label is kept", lm.dest),
							Span:    m.span(i, lm.next),
						})
						flush(i)
						nodes = append(nodes, children...)
					}
					i = lm.next
					continue
				}
				if lm.bracketed {
					p.diags = append(p.diags, diag.Diagnostic{
						Kind:    diag.UnresolvedLink,
						Message: "could not parse link destination; kept as literal text",
						Span:    m.span(i, lm.attemptEnd),
					})
				}
			}
			buf.WriteByte('[')
			if buf.Len() == 1 {
				textStart = i
			}
			i++
		case '*', '_':
			n := runLenAt(s, i, c)
			if n < 4 && canOpen(s, i, n, c) {
				if close := findEmphClose(s, i, c, n, noLinks); close >= 0 {
					flush(i)
					children := p.parse(s[i+n:close], m.sub(i+n), noLinks)
					nodes = append(nodes, wrapEmphasis(n, children, m.span(i, close+n), m.span(i+1, close+n-1)))
					i = close + n
					continue
				}
				p.diags = append(p.diags, diag.Diagnostic{
					Kind:    diag.UnmatchedDelimiter,
					Message: fmt.Sprintf("unmatched %q emphasis delimiter; kept as literal text", strings.Repeat(string(c), n)),
					Span:    m.span(i, i+n),
				})
			}
			appendText(i, i+n)
			i += n
		default:
			appendText(i, i+1)
			i++
		}
	}
	flush(len(s))
	return nodes
}

// wrapEmphasis maps run length 1/2/3 to em/strong/em-around-strong; inner
// is the triple run minus one delimiter byte on each side.
func wrapEmphasis(n int, children []ast.InlineNode, whole, inner ast.Span) ast.InlineNode {
	switch n {
	case 1:
		return &ast.EmphasisNode{Children: children, Span: whole}
	case 2:
		return &ast.StrongNode{Children: children, Span: whole}
	default: // n == 3
		strong := &ast.StrongNode{Children: children, Span: inner}
		return &ast.EmphasisNode{Children: []ast.InlineNode{strong}, Span: whole}
	}
}

type linkMatch struct {
	labelStart int
	labelEnd   int // the "]"
	dest       string
	title      string
	next       int
	attemptEnd int  // how far the scan got when matching fails
	ok         bool // complete link syntax with a valid destination
	bracketed  bool // "](" was present: a link was attempted
}

// matchLink checks whether s[i] starts an inline link. It is pure syntax —
// no nodes or diagnostics — so the emphasis closer search can use it.
func matchLink(s string, i int) linkMatch {
	var lm linkMatch
	end := findLabelEnd(s, i+1)
	if end < 0 || end+1 >= len(s) || s[end+1] != '(' {
		return lm
	}
	lm.labelStart, lm.labelEnd = i+1, end
	lm.bracketed = true
	lm.dest, lm.title, lm.next, lm.attemptEnd, lm.ok = parseDest(s, end+2)
	return lm
}

// findEmphClose finds the closing run for the opener at start (exactly n c
// characters), or -1. Nested pairs are skipped (stack semantics), as are
// escapes, code spans, and links, whose content is protected.
func findEmphClose(s string, start int, c byte, n int, noLinks bool) int {
	i := start + n
	for i < len(s) {
		switch {
		case s[i] == '\\':
			if i+1 < len(s) && isASCIIPunct(s[i+1]) {
				i += 2
			} else {
				i++
			}
		case s[i] == '`':
			bt := backtickRunAt(s, i)
			if close := findBacktickClose(s, i+bt, bt); close >= 0 {
				i = close + bt
			} else {
				i += bt
			}
		case s[i] == '[' && !noLinks:
			if lm := matchLink(s, i); lm.ok {
				i = lm.next
			} else {
				i++
			}
		case s[i] == c:
			run := runLenAt(s, i, c)
			if run == n {
				if canClose(s, i, run, c) {
					return i
				}
				if canOpen(s, i, run, c) {
					if j := findEmphClose(s, i, c, n, noLinks); j >= 0 {
						i = j + n
						continue
					}
				}
			}
			i += run
		default:
			i++
		}
	}
	return -1
}

func findBacktickClose(s string, start, n int) int {
	i := start
	for i < len(s) {
		if s[i] != '`' {
			i++
			continue
		}
		run := backtickRunAt(s, i)
		if run == n {
			return i
		}
		i += run
	}
	return -1
}

// stripCodeSpaces applies the CommonMark code span rule: one space is
// dropped from each end when both ends have one and not all is space.
func stripCodeSpaces(content string) string {
	if len(content) > 2 && content[0] == ' ' && content[len(content)-1] == ' ' &&
		strings.Trim(content, " ") != "" {
		return content[1 : len(content)-1]
	}
	return content
}

// findLabelEnd returns the ']' matching the '[' before i, skipping escapes
// and code spans; -1 when there is none.
func findLabelEnd(s string, i int) int {
	depth := 1
	for i < len(s) {
		switch s[i] {
		case '\\':
			if i+1 < len(s) && isASCIIPunct(s[i+1]) {
				i += 2
			} else {
				i++
			}
			continue
		case '`':
			bt := backtickRunAt(s, i)
			if close := findBacktickClose(s, i+bt, bt); close >= 0 {
				i = close + bt
			} else {
				i += bt
			}
			continue
		case '[':
			depth++
		case ']':
			depth--
			if depth == 0 {
				return i
			}
		}
		i++
	}
	return -1
}

// parseDest parses dest + optional "title" + ')': dest is a non-space run
// with balanced parens, the title is double-quoted after a single space.
func parseDest(s string, i int) (dest, title string, next, attemptEnd int, ok bool) {
	start := i
	depth := 0
scan:
	for i < len(s) {
		c := s[i]
		switch {
		case c == '\\' && i+1 < len(s) && isASCIIPunct(s[i+1]):
			i += 2
			continue
		case c == '(':
			depth++
		case c == ')':
			if depth == 0 {
				break scan
			}
			depth--
		case c <= ' ': // space or control ends the destination run
			break scan
		}
		i++
	}
	attemptEnd = i

	if depth != 0 {
		return "", "", 0, attemptEnd, false
	}
	dest = s[start:i]

	j := i
	if j < len(s) && s[j] == ' ' && j+1 < len(s) && s[j+1] == '"' {
		end := strings.IndexByte(s[j+2:], '"')
		if end < 0 {
			return "", "", 0, attemptEnd, false
		}
		title = s[j+2 : j+2+end]
		j = j + 2 + end + 1
	}
	if j >= len(s) || s[j] != ')' {
		return "", "", 0, attemptEnd, false
	}
	return dest, title, j + 1, attemptEnd, true
}

// canOpen reports whether the run of n c at i can open emphasis/strong
// (CommonMark 0.31.2 rules 1, 2, 5, 6).
func canOpen(s string, i, n int, c byte) bool {
	left, right := flanking(s, i, n)
	if c == '*' {
		return left
	}
	prev, _ := runeBefore(s, i)
	return left && (!right || isUnicodePunct(prev))
}

// canClose reports whether the run of n c at i can close emphasis/strong
// (CommonMark 0.31.2 rules 3, 4, 7, 8).
func canClose(s string, i, n int, c byte) bool {
	left, right := flanking(s, i, n)
	if c == '*' {
		return right
	}
	next, _ := runeAfter(s, i+n)
	return right && (!left || isUnicodePunct(next))
}

// flanking implements the CommonMark left/right-flanking run rules; the
// beginning and end of s count as whitespace.
func flanking(s string, i, n int) (left, right bool) {
	prev, prevStart := runeBefore(s, i)
	next, nextEnd := runeAfter(s, i+n)

	prevWS := prevStart || unicode.IsSpace(prev)
	nextWS := nextEnd || unicode.IsSpace(next)
	prevPunct := !prevStart && isUnicodePunct(prev)
	nextPunct := !nextEnd && isUnicodePunct(next)

	left = !nextWS && (!nextPunct || prevWS || prevPunct)
	right = !prevWS && (!prevPunct || nextWS || nextPunct)
	return left, right
}

func runeBefore(s string, i int) (r rune, start bool) {
	if i <= 0 {
		return 0, true
	}
	r, _ = utf8.DecodeLastRuneInString(s[:i])
	return r, false
}

func runeAfter(s string, i int) (r rune, end bool) {
	if i >= len(s) {
		return 0, true
	}
	r, _ = utf8.DecodeRuneInString(s[i:])
	return r, false
}

// isUnicodePunct matches the CommonMark punctuation definition (Unicode
// punctuation and symbol categories).
func isUnicodePunct(r rune) bool {
	return unicode.IsPunct(r) || unicode.IsSymbol(r)
}

func isASCIIPunct(c byte) bool {
	return strings.IndexByte(asciiPunct, c) >= 0
}

const asciiPunct = "!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"

func backtickRunAt(s string, i int) int {
	n := 0
	for i+n < len(s) && s[i+n] == '`' {
		n++
	}
	return n
}

func runLenAt(s string, i int, c byte) int {
	n := 0
	for i+n < len(s) && s[i+n] == c {
		n++
	}
	return n
}
