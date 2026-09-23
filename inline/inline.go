// Package inline implements mdp's inline phase: it splits raw text into
// emphasis, strong, code, and link nodes and sanitizes link URLs. Parsing
// never fails; every malformed or unmatched construct degrades to literal
// text, and links with unsafe URLs degrade to their plain label content.
package inline

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/ozguragain/mdp/ast"
)

// Parse splits raw inline text into inline nodes. It never fails: any
// malformed or unmatched construct degrades to literal text.
func Parse(text string) []ast.InlineNode {
	return parse(text, false)
}

// Process walks the document and replaces each block's inline slice with
// Parse(...) of its raw text. TextNode values of one block are joined with
// "\n" first (soft breaks). It mutates doc in place and must be called
// exactly once, before codegen. Fenced code blocks are left untouched.
func Process(doc *ast.DocumentNode) {
	processBlocks(doc.Blocks)
}

func processBlocks(blocks []ast.BlockNode) {
	for _, block := range blocks {
		switch n := block.(type) {
		case *ast.HeadingNode:
			n.Inlines = parseRawInlines(n.Inlines)
		case *ast.ParagraphNode:
			n.Inlines = parseRawInlines(n.Inlines)
		case *ast.BlockquoteNode:
			processBlocks(n.Blocks)
		case *ast.ListNode:
			for _, item := range n.Items {
				processBlocks(item.Blocks)
			}
		case *ast.ListItemNode:
			processBlocks(n.Blocks)
		}
	}
}

// parseRawInlines joins the raw TextNodes of one block (one per source
// line) with "\n" and parses the result. Blocks already carrying non-text
// inlines are left untouched.
func parseRawInlines(inlines []ast.InlineNode) []ast.InlineNode {
	if len(inlines) == 0 {
		return nil
	}
	parts := make([]string, 0, len(inlines))
	for _, in := range inlines {
		t, ok := in.(*ast.TextNode)
		if !ok {
			return inlines
		}
		parts = append(parts, t.Value)
	}
	return Parse(strings.Join(parts, "\n"))
}

// parse splits s into inline nodes in a single left-to-right scan; at each
// position the first construct that matches wins, and construct content is
// parsed recursively. When noLinks is set (inside a link label) link syntax
// is not recognized and "[" stays literal text.
func parse(s string, noLinks bool) []ast.InlineNode {
	var nodes []ast.InlineNode
	var buf strings.Builder
	flush := func() {
		if buf.Len() > 0 {
			nodes = append(nodes, &ast.TextNode{Value: buf.String()})
			buf.Reset()
		}
	}

	i := 0
	for i < len(s) {
		switch c := s[i]; c {
		case '\\':
			// Backslash escapes: `\*` yields a literal "*" that is never
			// re-interpreted. Before anything else the backslash is kept.
			if i+1 < len(s) && isASCIIPunct(s[i+1]) {
				buf.WriteByte(s[i+1])
				i += 2
				continue
			}
			buf.WriteByte('\\')
			i++
		case '`':
			// Code spans take precedence over every other construct: their
			// content is literal and never re-parsed.
			n := backtickRunAt(s, i)
			if close := findBacktickClose(s, i+n, n); close >= 0 {
				flush()
				nodes = append(nodes, &ast.CodeNode{Value: stripCodeSpaces(s[i+n : close])})
				i = close + n
				continue
			}
			buf.WriteString(s[i : i+n]) // unmatched run stays literal
			i += n
		case '[':
			if !noLinks {
				if linkNodes, next, ok := parseLinkNodes(s, i); ok {
					flush()
					nodes = append(nodes, linkNodes...)
					i = next
					continue
				}
			}
			buf.WriteByte('[')
			i++
		case '*', '_':
			n := runLenAt(s, i, c)
			if n < 4 && canOpen(s, i, n, c) {
				if close := findEmphClose(s, i, c, n, noLinks); close >= 0 {
					flush()
					children := parse(s[i+n:close], noLinks)
					nodes = append(nodes, wrapEmphasis(n, children))
					i = close + n
					continue
				}
			}
			buf.WriteString(s[i : i+n]) // unmatched or too-long run stays literal
			i += n
		default:
			buf.WriteByte(c)
			i++
		}
	}
	flush()
	return nodes
}

// wrapEmphasis wraps parsed children according to the delimiter run length:
// 1 → em, 2 → strong, 3 → em around strong (CommonMark nesting order).
func wrapEmphasis(n int, children []ast.InlineNode) ast.InlineNode {
	switch n {
	case 1:
		return &ast.EmphasisNode{Children: children}
	case 2:
		return &ast.StrongNode{Children: children}
	default: // n == 3
		return &ast.EmphasisNode{Children: []ast.InlineNode{&ast.StrongNode{Children: children}}}
	}
}

// findEmphClose returns the start index of the closing delimiter run for the
// opener at start (a run of exactly n c characters), or -1. Complete nested
// pairs of the same delimiter and length are skipped (stack semantics), and
// so are escaped characters, complete code spans, and complete links, whose
// content is protected.
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
			if _, next, ok := parseLinkNodes(s, i); ok {
				i = next
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
					// Skip a complete nested pair and keep looking.
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

// findBacktickClose returns the index of the backtick run of exactly n
// characters that closes the run starting at start, or -1.
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

// stripCodeSpaces applies the CommonMark code span rule: when the content
// both begins and ends with a space and is not all spaces, one space is
// removed from each end.
func stripCodeSpaces(content string) string {
	if len(content) > 2 && content[0] == ' ' && content[len(content)-1] == ' ' &&
		strings.Trim(content, " ") != "" {
		return content[1 : len(content)-1]
	}
	return content
}

// parseLinkNodes parses an inline link starting at s[i] == '['. It returns
// the nodes to splice into the output (the link itself, or just its parsed
// label when the URL is unsafe), the index just past the link, and whether
// a link was recognized at all.
func parseLinkNodes(s string, i int) ([]ast.InlineNode, int, bool) {
	labelEnd := findLabelEnd(s, i+1)
	if labelEnd < 0 || labelEnd+1 >= len(s) || s[labelEnd+1] != '(' {
		return nil, 0, false
	}
	dest, title, next, ok := parseDest(s, labelEnd+2)
	if !ok {
		return nil, 0, false
	}
	children := parse(s[i+1:labelEnd], true) // nested links are suppressed
	if !isSafeURL(dest) {
		return children, next, true
	}
	node := &ast.LinkNode{Destination: dest, Title: title, Children: children}
	return []ast.InlineNode{node}, next, true
}

// findLabelEnd returns the index of the ']' matching the '[' before i, or
// -1. Bracket depth is tracked; escaped characters and complete code spans
// are skipped as opaque.
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

// parseDest parses the tail of an inline link: destination, an optional
// "title", and the closing ')'. The destination is a run of non-space,
// non-control characters with balanced parentheses (possibly empty); the
// title is optional, double-quoted, and separated from the destination by a
// single space. Anything else is not a link.
func parseDest(s string, i int) (dest, title string, next int, ok bool) {
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

	if depth != 0 {
		return "", "", 0, false
	}
	dest = s[start:i]

	j := i
	if j < len(s) && s[j] == ' ' && j+1 < len(s) && s[j+1] == '"' {
		end := strings.IndexByte(s[j+2:], '"')
		if end < 0 {
			return "", "", 0, false
		}
		title = s[j+2 : j+2+end]
		j = j + 2 + end + 1
	}
	if j >= len(s) || s[j] != ')' {
		return "", "", 0, false
	}
	return dest, title, j + 1, true
}

// canOpen reports whether the delimiter run of n c characters at i can open
// emphasis/strong, per CommonMark 0.31.2 rules 1, 2, 5 and 6.
func canOpen(s string, i, n int, c byte) bool {
	left, right := flanking(s, i, n)
	if c == '*' {
		return left
	}
	prev, _ := runeBefore(s, i)
	return left && (!right || isUnicodePunct(prev))
}

// canClose reports whether the delimiter run of n c characters at i can
// close emphasis/strong, per CommonMark 0.31.2 rules 3, 4, 7 and 8.
func canClose(s string, i, n int, c byte) bool {
	left, right := flanking(s, i, n)
	if c == '*' {
		return right
	}
	next, _ := runeAfter(s, i+n)
	return right && (!left || isUnicodePunct(next))
}

// flanking implements the CommonMark left/right-flanking delimiter run
// definitions. The beginning and the end of s count as whitespace.
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

// runeBefore returns the rune before index i and whether i is at the start
// of s. runeAfter returns the rune at index i and whether i is past the end.
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

// isUnicodePunct reports whether r is a Unicode punctuation or symbol
// character (the CommonMark punctuation definition).
func isUnicodePunct(r rune) bool {
	return unicode.IsPunct(r) || unicode.IsSymbol(r)
}

// isASCIIPunct reports whether c is an ASCII punctuation character (the
// backslash-escapable set).
func isASCIIPunct(c byte) bool {
	return strings.IndexByte(asciiPunct, c) >= 0
}

const asciiPunct = "!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"

// backtickRunAt returns the number of consecutive backticks starting at i.
func backtickRunAt(s string, i int) int {
	n := 0
	for i+n < len(s) && s[i+n] == '`' {
		n++
	}
	return n
}

// runLenAt returns the number of consecutive c bytes starting at i.
func runLenAt(s string, i int, c byte) int {
	n := 0
	for i+n < len(s) && s[i+n] == c {
		n++
	}
	return n
}
