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
	// Frontmatter is metadata, not content: consume and drop it. Container
	// sub-parsers never see this token.
	if p.currentToken().Type == lexer.TokenFrontmatter {
		p.nextToken()
	}
	return &ast.DocumentNode{Blocks: p.parseBlocks()}
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
		case lexer.TokenListItem, lexer.TokenOrderedListItem:
			blocks = append(blocks, p.parseList())
		case lexer.TokenThematicBreak:
			tok := p.currentToken()
			p.nextToken()
			blocks = append(blocks, &ast.ThematicBreakNode{Span: tok.Span})
		case lexer.TokenTableHeader:
			blocks = append(blocks, p.parseTable())
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

// parseList collects items of one kind. A bullet item and an ordered item
// never share a list: the marker change closes the list and opens another.
func (p *Parser) parseList() *ast.ListNode {
	ordered := p.currentToken().Type == lexer.TokenOrderedListItem
	want := lexer.TokenListItem
	if ordered {
		want = lexer.TokenOrderedListItem
	}

	list := &ast.ListNode{Ordered: ordered, Start: 1}
	for p.currentToken().Type == want {
		tok := p.currentToken()
		if len(list.Items) == 0 && ordered {
			list.Start = markerNumber(tok.Meta)
		}
		list.Items = append(list.Items, p.parseListItem())
	}
	list.Span = ast.Span{
		Start: list.Items[0].Span.Start,
		End:   list.Items[len(list.Items)-1].Span.End,
	}
	return list
}

// markerNumber reads the number off an ordered list marker ("1." -> 1).
func markerNumber(marker string) int {
	n := 0
	for i := 0; i < len(marker) && marker[i] >= '0' && marker[i] <= '9'; i++ {
		n = n*10 + int(marker[i]-'0')
	}
	if n == 0 {
		return 1
	}
	return n
}

// parseListItem absorbs following lines indented past the marker as item
// content — wrapped text, nested blocks, and open fences alike.
func (p *Parser) parseListItem() *ast.ListItemNode {
	tok := p.currentToken()
	p.nextToken()

	indent := leadingSpaces(tok.Literal)
	// A bullet marker is 2 bytes ("- "); an ordered marker is wider.
	markerWidth := 2
	if tok.Type == lexer.TokenOrderedListItem {
		markerWidth = len(tok.Meta) + 1 // marker plus the space after it
	}
	content, start := itemContent(tok.Literal, markerWidth)

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
		dedented, removed := dedent(next.Literal, indent+markerWidth)
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

// parseTable builds a GFM pipe table: a header row, a delimiter row that is
// consumed and turned into column alignments, then any body rows. The lexer
// has already proved the delimiter row is real.
func (p *Parser) parseTable() *ast.TableNode {
	headerTok := p.currentToken()
	p.nextToken()

	delimTok := p.currentToken()
	p.nextToken() // the delimiter row carries no content

	aligns := parseAlignments(delimTok.Meta)
	table := &ast.TableNode{Alignments: aligns}
	table.Header = padCells(p.rowCells(headerTok), aligns)

	end := headerTok.Span.End
	for p.currentToken().Type == lexer.TokenTableRow {
		tok := p.currentToken()
		p.nextToken()
		table.Rows = append(table.Rows, padCells(p.rowCells(tok), aligns))
		end = tok.Span.End
	}
	table.Span = ast.Span{Start: headerTok.Span.Start, End: end}
	return table
}

// parseAlignments reads the delimiter row's column alignments.
func parseAlignments(line string) []ast.Align {
	var aligns []ast.Align
	for _, cell := range splitRowContent(line) {
		c := cell.text
		left := strings.HasPrefix(c, ":")
		right := strings.HasSuffix(c, ":")
		switch {
		case left && right:
			aligns = append(aligns, ast.AlignCenter)
		case left:
			aligns = append(aligns, ast.AlignLeft)
		case right:
			aligns = append(aligns, ast.AlignRight)
		default:
			aligns = append(aligns, ast.AlignNone)
		}
	}
	return aligns
}

// padCells normalises a row to the delimiter's column count: short rows gain
// empty cells, long rows lose their tail.
func padCells(cells []ast.TableCellNode, aligns []ast.Align) []ast.TableCellNode {
	for len(cells) < len(aligns) {
		cells = append(cells, ast.TableCellNode{})
	}
	if len(cells) > len(aligns) {
		cells = cells[:len(aligns)]
	}
	return cells
}

// rowCells splits one table row into cells, each span covering exactly the
// cell's content with its padding spaces left out.
func (p *Parser) rowCells(tok lexer.Token) []ast.TableCellNode {
	var cells []ast.TableCellNode
	for _, cp := range splitRowContent(tok.Literal) {
		start := advance(tok.Span.Start, cp.start)
		span := ast.Span{Start: start, End: advance(start, len(cp.text))}
		cells = append(cells, ast.TableCellNode{
			Inlines: textInlines(cp.text, span),
			Span:    span,
		})
	}
	return cells
}

// rowCell is one table cell's text and where that text starts in its line.
type rowCell struct {
	text  string
	start int
}

// splitRowContent splits a raw table line on its pipes and reports each
// cell's content with the offset of that content in the line. Whitespace
// padding is trimmed, so the reported offset points at the content itself.
func splitRowContent(line string) []rowCell {
	lead := len(line) - len(strings.TrimLeftFunc(line, unicode.IsSpace))
	s := line[lead:]
	s = strings.TrimRightFunc(s, unicode.IsSpace)

	i := 0
	if strings.HasPrefix(s, "|") {
		i = 1
	}

	var out []rowCell
	for {
		j := i
		for j < len(s) && s[j] != '|' {
			j++
		}
		raw := s[i:j]
		begin := 0
		for begin < len(raw) && raw[begin] == ' ' {
			begin++
		}
		stop := len(raw)
		for stop > begin && raw[stop-1] == ' ' {
			stop--
		}
		out = append(out, rowCell{text: raw[begin:stop], start: lead + i + begin})

		if j >= len(s) {
			return out
		}
		i = j + 1
		if i >= len(s) {
			return out // the row ends with a pipe: no cell after it
		}
	}
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

// itemContent drops the list marker and reports the bytes dropped before the
// content, so the content's span lands on its real column.
func itemContent(line string, markerWidth int) (content string, removed int) {
	lead := len(line) - len(strings.TrimLeftFunc(line, unicode.IsSpace))
	after := strings.TrimSpace(line)[markerWidth:] // drop the marker
	content = strings.TrimSpace(after)
	extra := len(after) - len(strings.TrimLeftFunc(after, unicode.IsSpace))
	return content, lead + markerWidth + extra
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
