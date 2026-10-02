package lexer

import "github.com/ozguragain/mdp/ast"

type TokenType string

const (
	TokenBlankLine  TokenType = "BLANK_LINE"
	TokenHeading    TokenType = "HEADING"
	TokenCodeFence  TokenType = "CODE_FENCE"
	TokenCodeLine   TokenType = "CODE_LINE"
	TokenListItem   TokenType = "LIST_ITEM"
	TokenBlockquote TokenType = "BLOCKQUOTE"
	TokenTextLine   TokenType = "TEXT_LINE"
	TokenEOF        TokenType = "EOF"
)

type Token struct {
	Type    TokenType
	Literal string   // Original text of the token
	Meta    string   // Extra information about the token
	Span    ast.Span // Source extent of the token (a whole line, or a point at EOF)
}
