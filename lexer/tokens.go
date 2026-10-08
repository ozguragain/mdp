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

	// TokenThematicBreak: "---", "***" or "___" alone on its line.
	TokenThematicBreak TokenType = "THEMATIC_BREAK"
	// TokenOrderedListItem: "1." / "1)" followed by content. Meta holds the
	// marker exactly as written ("1.", "42)").
	TokenOrderedListItem TokenType = "ORDERED_LIST_ITEM"
	// TokenFrontmatter: the whole leading "---" ... "---" block, consumed as
	// one token so none of it reaches the block parser. Meta holds the raw
	// YAML body. Emitted only by NewDocumentLexer.
	TokenFrontmatter TokenType = "FRONTMATTER"
	// TokenTableHeader / TokenTableDelimiter / TokenTableRow are the three
	// lines of a GFM pipe table. Meta on the delimiter holds the raw line so
	// the parser can read the column alignments.
	TokenTableHeader    TokenType = "TABLE_HEADER"
	TokenTableDelimiter TokenType = "TABLE_DELIMITER"
	TokenTableRow       TokenType = "TABLE_ROW"
)

type Token struct {
	Type    TokenType
	Literal string   // Original text of the token
	Meta    string   // Extra information about the token
	Span    ast.Span // Source extent of the token (a whole line, or a point at EOF)
}
