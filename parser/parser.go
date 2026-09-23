// Package parser builds a markdown AST from the lexer token stream.
package parser

import (
	"strconv"
	"strings"

	"github.com/ozguragain/mdp/ast"
	"github.com/ozguragain/mdp/lexer"
)

// Parser turns the flat, line-oriented token stream into a block tree.
// Container blocks (blockquote, list item) reach the parser as raw source
// lines; their content is re-lexed and re-parsed recursively via parseSource.
type Parser struct {
	tokens   []lexer.Token
	position int
}

func NewParser(tokens []lexer.Token) *Parser {
	return &Parser{
		tokens:   tokens,
		position: 0,
	}
}

// Parse consumes the token stream and returns the document tree.
func (p *Parser) Parse() *ast.DocumentNode {
	return &ast.DocumentNode{Blocks: p.parseBlocks()}
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

// parseBlocks builds blocks until EOF. Blank lines separate blocks and are
// simply consumed.
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
			// Unreachable with the current lexer: CODE_LINE only appears
			// between fences, which parseFencedCode consumes. Skip unknown
			// tokens so the loop always makes progress.
			p.nextToken()
		}
	}
}

// parseHeading builds a HeadingNode from a single HEADING token.
func (p *Parser) parseHeading() *ast.HeadingNode {
	tok := p.currentToken()
	p.nextToken()

	// The lexer guarantees Meta is a level in "1".."6" and that the trimmed
	// literal starts with exactly that many hashes followed by a space.
	level, _ := strconv.Atoi(tok.Meta)

	content := strings.TrimSpace(tok.Literal) // drop indentation
	content = strings.TrimSpace(content[level:])
	content = trimClosingHashes(content)

	return &ast.HeadingNode{Level: level, Inlines: textInlines(content)}
}

// parseParagraph collects consecutive TEXT_LINE tokens into one paragraph.
// Each source line becomes its own TextNode (spec: no merging; the inline
// phase will later split these further).
func (p *Parser) parseParagraph() *ast.ParagraphNode {
	var inlines []ast.InlineNode
	for p.currentToken().Type == lexer.TokenTextLine {
		inlines = append(inlines, &ast.TextNode{Value: p.currentToken().Literal})
		p.nextToken()
	}
	return &ast.ParagraphNode{Inlines: inlines}
}

// parseFencedCode builds a FencedCodeBlockNode from an opening fence, the
// code lines, and the closing fence. An unclosed block runs to EOF.
func (p *Parser) parseFencedCode() *ast.FencedCodeBlockNode {
	opening := p.currentToken()
	p.nextToken()

	var lines []string
	for p.currentToken().Type == lexer.TokenCodeLine {
		lines = append(lines, p.currentToken().Literal)
		p.nextToken()
	}
	if p.currentToken().Type == lexer.TokenCodeFence {
		p.nextToken() // closing fence
	}

	return &ast.FencedCodeBlockNode{Info: opening.Meta, Lines: lines}
}

// parseBlockquote collects consecutive BLOCKQUOTE tokens, strips one ">"
// marker from each line, and parses the remaining source recursively.
func (p *Parser) parseBlockquote() *ast.BlockquoteNode {
	var lines []string
	for p.currentToken().Type == lexer.TokenBlockquote {
		lines = append(lines, stripQuoteMarker(p.currentToken().Literal))
		p.nextToken()
	}
	return &ast.BlockquoteNode{Blocks: parseSource(strings.Join(lines, "\n"))}
}

// parseList collects consecutive LIST_ITEM tokens into a single list.
func (p *Parser) parseList() *ast.ListNode {
	list := &ast.ListNode{}
	for p.currentToken().Type == lexer.TokenListItem {
		list.Items = append(list.Items, p.parseListItem())
	}
	return list
}

// parseListItem builds one list item. Following lines indented deeper
// than the marker are absorbed as its content — wrapped text and nested
// blocks alike — so nested lists fall out of the recursive parse
// naturally. Inside a fence opened from the item, everything up to the
// closing fence is content, whatever its indentation.
func (p *Parser) parseListItem() *ast.ListItemNode {
	tok := p.currentToken()
	p.nextToken()

	indent := leadingSpaces(tok.Literal)
	content := strings.TrimSpace(tok.Literal)
	content = strings.TrimSpace(content[2:]) // drop the "- " / "* " / "+ " marker

	lines := []string{content}
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
		lines = append(lines, dedent(next.Literal, indent+2))
		p.nextToken()
	}

	return &ast.ListItemNode{Blocks: parseSource(strings.Join(lines, "\n"))}
}

// parseSource re-lexes and re-parses a chunk of markdown source. It is used
// for container blocks whose inner content reaches the parser as raw lines.
func parseSource(src string) []ast.BlockNode {
	tokens := lexer.NewLexer(src).Tokenize()
	return NewParser(tokens).Parse().Blocks
}

// textInlines wraps raw text as a single TextNode. Inline markup (emphasis,
// links, code spans) is a separate phase: package inline later joins and
// re-splits these raw nodes via inline.Process. Here text stays raw.
func textInlines(text string) []ast.InlineNode {
	if text == "" {
		return nil
	}
	return []ast.InlineNode{&ast.TextNode{Value: text}}
}

// trimClosingHashes removes an optional ATX closing sequence
// ("## Heading ##" -> "Heading"). The run of hashes only counts as a
// closing sequence when preceded by a space or when it is the whole content.
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

// stripQuoteMarker removes one ">" (plus a single following space) from a
// blockquote line, yielding the raw source of the quoted content. Nested
// markers (">> deep") survive and are handled by the recursive parse.
func stripQuoteMarker(line string) string {
	s := strings.TrimPrefix(strings.TrimSpace(line), ">")
	return strings.TrimPrefix(s, " ")
}

// leadingSpaces returns the count of leading space characters in s.
func leadingSpaces(s string) int {
	n := 0
	for n < len(s) && s[n] == ' ' {
		n++
	}
	return n
}

// dedent removes up to n leading spaces from s.
func dedent(s string, n int) string {
	i := 0
	for i < n && i < len(s) && s[i] == ' ' {
		i++
	}
	return s[i:]
}
