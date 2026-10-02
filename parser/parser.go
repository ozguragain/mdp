// Package parser builds a markdown AST from the lexer token stream.
package parser

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/ozguragain/mdp/ast"
	"github.com/ozguragain/mdp/diag"
	"github.com/ozguragain/mdp/lexer"
)

// Parser turns the token stream into a block tree. Parsing never fails:
// malformed input degrades to literal text and is reported through
// Diagnostics.
type Parser struct {
	tokens   []lexer.Token
	position int
	diags    []diag.Diagnostic
}

func NewParser(tokens []lexer.Token) *Parser {
	return &Parser{
		tokens:   tokens,
		position: 0,
	}
}

func (p *Parser) Parse() *ast.DocumentNode {
	blocks := p.parseBlocks()
	doc := &ast.DocumentNode{Blocks: blocks}
	if len(blocks) > 0 {
		doc.Span = ast.Span{
			Start: blocks[0].SourceSpan().Start,
			End:   blocks[len(blocks)-1].SourceSpan().End,
		}
	}
	return doc
}

// Diagnostics reports what Parse recovered; parsing still succeeds.
func (p *Parser) Diagnostics() []diag.Diagnostic {
	return p.diags
}

func (p *Parser) currentToken() lexer.Token {
	if p.position >= len(p.tokens) {
		return lexer.Token{Type: lexer.TokenEOF}
	}
	return p.tokens[p.position]
}

func (p *Parser) nextToken() {
	if p.position < len(p.tokens) {
		p.position++
	}
}

func (p *Parser) parseBlocks() []ast.BlockNode {
	var blocks []ast.BlockNode
	for {
		switch p.currentToken().Type {
		case lexer.TokenEOF:
			return blocks
		case lexer.TokenBlankLine:
			p.nextToken()
		case lexer.TokenHeading:
			blocks = append(blocks, p.parseHeading())
		case lexer.TokenTextLine:
			blocks = append(blocks, p.parseParagraph())
		case lexer.TokenCodeFence:
			blocks = append(blocks, p.parseFencedCode())
		case lexer.TokenBlockquote:
			blocks = append(blocks, p.parseBlockquote())
		case lexer.TokenListItem:
			blocks = append(blocks, p.parseList())
		default:
			// Unreachable with the current lexer; skip so the loop always
			// makes progress.
			p.nextToken()
		}
	}
}

// parseHeading tracks the trims applied to the source line so the content
// span (and later inline nodes) land on the right columns.
func (p *Parser) parseHeading() *ast.HeadingNode {
	tok := p.currentToken()
	p.nextToken()

	level, _ := strconv.Atoi(tok.Meta)

	lit := tok.Literal
	trimmed := strings.TrimSpace(lit) // drop indentation
	afterHash := trimmed[level:]
	content := strings.TrimSpace(afterHash)
	content = trimClosingHashes(content)

	lead := len(lit) - len(strings.TrimLeftFunc(lit, unicode.IsSpace))
	extra := len(afterHash) - len(strings.TrimLeftFunc(afterHash, unicode.IsSpace))
	start := advance(tok.Span.Start, lead+level+extra)
	textSpan := ast.Span{Start: start, End: advance(start, len(content))}

	return &ast.HeadingNode{
		Level:   level,
		Inlines: textInlines(content, textSpan),
		Span:    tok.Span,
	}
}

// parseParagraph keeps one TextNode per source line (spec: no merging).
func (p *Parser) parseParagraph() *ast.ParagraphNode {
	var inlines []ast.InlineNode
	var span ast.Span
	for p.currentToken().Type == lexer.TokenTextLine {
		tok := p.currentToken()
		if len(inlines) == 0 {
			span.Start = tok.Span.Start
		}
		span.End = tok.Span.End
		inlines = append(inlines, &ast.TextNode{Value: tok.Literal, Span: tok.Span})
		p.nextToken()
	}
	return &ast.ParagraphNode{Inlines: inlines, Span: span}
}

// parseFencedCode builds a FencedCodeBlockNode; an unclosed block runs to
// EOF and is reported as a diagnostic.
func (p *Parser) parseFencedCode() *ast.FencedCodeBlockNode {
	opening := p.currentToken()
	p.nextToken()

	start := opening.Span.Start
	end := opening.Span.End

	var lines []string
	for p.currentToken().Type == lexer.TokenCodeLine {
		tok := p.currentToken()
		lines = append(lines, tok.Literal)
		end = tok.Span.End
		p.nextToken()
	}
	if p.currentToken().Type == lexer.TokenCodeFence {
		end = p.currentToken().Span.End
		p.nextToken() // closing fence
	} else {
		p.diags = append(p.diags, diag.Diagnostic{
			Kind:    diag.UnclosedFence,
			Message: "fenced code block is never closed; the block runs to end of input",
			Span:    ast.Span{Start: start, End: end},
		})
	}

	return &ast.FencedCodeBlockNode{
		Info:  opening.Meta,
		Lines: lines,
		Span:  ast.Span{Start: start, End: end},
	}
}

// parseBlockquote strips one ">" marker per line and re-parses the content;
// per-line bases map spans back past the markers to real columns.
func (p *Parser) parseBlockquote() *ast.BlockquoteNode {
	var lines []string
	var bases []ast.Pos
	var span ast.Span
	for p.currentToken().Type == lexer.TokenBlockquote {
		tok := p.currentToken()
		if len(lines) == 0 {
			span.Start = tok.Span.Start
		}
		span.End = tok.Span.End
		content, start := stripQuoteMarker(tok.Literal)
		lines = append(lines, content)
		bases = append(bases, advance(tok.Span.Start, start))
		p.nextToken()
	}
	blocks := parseSource(strings.Join(lines, "\n"), bases, p)
	return &ast.BlockquoteNode{Blocks: blocks, Span: span}
}

func (p *Parser) parseList() *ast.ListNode {
	list := &ast.ListNode{}
	for p.currentToken().Type == lexer.TokenListItem {
		list.Items = append(list.Items, p.parseListItem())
	}
	list.Span = ast.Span{
		Start: list.Items[0].Span.Start,
		End:   list.Items[len(list.Items)-1].Span.End,
	}
	return list
}

// parseListItem absorbs following lines indented past the marker as item
// content — wrapped text, nested blocks, and open fences alike.
func (p *Parser) parseListItem() *ast.ListItemNode {
	tok := p.currentToken()
	p.nextToken()

	indent := leadingSpaces(tok.Literal)
	content, start := itemContent(tok.Literal)

	lines := []string{content}
	bases := []ast.Pos{advance(tok.Span.Start, start)}
	end := tok.Span.End

	inFence := false
	for {
		next := p.currentToken()
		if next.Type == lexer.TokenEOF {
			break
		}
		if !inFence && next.Type == lexer.TokenBlankLine {
			break // a blank line ends the item
		}
		if !inFence && leadingSpaces(next.Literal) <= indent {
			break // sibling item or top-level block
		}
		if next.Type == lexer.TokenCodeFence {
			inFence = !inFence // open the item's fence, then close it
		}
		dedented, removed := dedent(next.Literal, indent+2)
		lines = append(lines, dedented)
		bases = append(bases, advance(next.Span.Start, removed))
		end = next.Span.End
		p.nextToken()
	}

	blocks := parseSource(strings.Join(lines, "\n"), bases, p)
	return &ast.ListItemNode{
		Blocks: blocks,
		Span:   ast.Span{Start: tok.Span.Start, End: end},
	}
}

// parseSource re-lexes and re-parses container content; bases maps each
// chunk line back to its parent position so spans stay real.
func parseSource(src string, bases []ast.Pos, parent *Parser) []ast.BlockNode {
	tokens := lexer.NewLexer(src).Tokenize()
	sub := NewParser(tokens)
	blocks := sub.Parse().Blocks
	ast.RemapBlocks(blocks, bases)
	for _, d := range sub.Diagnostics() {
		d.Span = ast.RemapSpan(d.Span, bases)
		parent.diags = append(parent.diags, d)
	}
	return blocks
}

// textInlines wraps raw text as one TextNode; package inline splits it
// further via inline.Process.
func textInlines(text string, span ast.Span) []ast.InlineNode {
	if text == "" {
		return nil
	}
	return []ast.InlineNode{&ast.TextNode{Value: text, Span: span}}
}

// advance moves p forward n bytes along its line (byte columns).
func advance(p ast.Pos, n int) ast.Pos {
	return ast.Pos{Offset: p.Offset + n, Line: p.Line, Column: p.Column + n}
}

// trimClosingHashes removes an optional ATX closing sequence
// ("## Heading ##" -> "Heading").
func trimClosingHashes(s string) string {
	trimmed := strings.TrimRight(s, "#")
	if trimmed == s {
		return s
	}
	if trimmed == "" || strings.HasSuffix(trimmed, " ") {
		return strings.TrimRight(trimmed, " ")
	}
	return s
}

// stripQuoteMarker removes one ">" (plus one following space) from a line
// and reports the bytes dropped, so content maps back to its column.
func stripQuoteMarker(line string) (content string, removed int) {
	lead := len(line) - len(strings.TrimLeftFunc(line, unicode.IsSpace))
	after := strings.TrimPrefix(strings.TrimSpace(line), ">")
	content = strings.TrimPrefix(after, " ")
	return content, lead + 1 + (len(after) - len(content))
}

// itemContent drops the "- " / "* " / "+ " marker and reports the bytes
// dropped before the content.
func itemContent(line string) (content string, removed int) {
	lead := len(line) - len(strings.TrimLeftFunc(line, unicode.IsSpace))
	after := strings.TrimSpace(line)[2:] // drop the marker
	content = strings.TrimSpace(after)
	extra := len(after) - len(strings.TrimLeftFunc(after, unicode.IsSpace))
	return content, lead + 2 + extra
}

func leadingSpaces(s string) int {
	n := 0
	for n < len(s) && s[n] == ' ' {
		n++
	}
	return n
}

// dedent removes up to n leading spaces and reports the bytes removed.
func dedent(s string, n int) (string, int) {
	i := 0
	for i < n && i < len(s) && s[i] == ' ' {
		i++
	}
	return s[i:], i
}
