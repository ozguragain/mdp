package parser

import (
	"github.com/ozguragain/mdp/ast"
	"github.com/ozguragain/mdp/lexer"
)

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

func (p *Parser) currentToken() lexer.Token {
	if p.position >= len(p.tokens) {
		return lexer.Token{Type: lexer.TokenEOF, Literal: ""}
	}
	return p.tokens[p.position]
}

func (p *Parser) nextToken() {
	if p.position < len(p.tokens) {
		p.position++
	}
}

func (p *Parser) Parse() *ast.DocumentNode {
	doc := &ast.DocumentNode{
		Blocks: []ast.BlockNode{},
	}

	// TODO

	return doc
}
