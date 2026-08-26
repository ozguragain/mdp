package lexer

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
	Literal string // Originak text of the token
	Meta    string // Extra information about the token
}
