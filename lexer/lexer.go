package lexer

import (
	"strconv"
	"strings"
)

type Lexer struct {
	lines       []string
	position    int
	inCodeBlock bool
}

func NewLexer(input string) *Lexer {
	input = strings.ReplaceAll(input, "\r\n", "\n")

	return &Lexer{
		lines:       strings.Split(input, "\n"),
		position:    0,
		inCodeBlock: false,
	}
}

func (l *Lexer) Tokenize() []Token {
	var tokens []Token
	for {
		token := l.NextToken()
		tokens = append(tokens, token)

		if token.Type == TokenEOF {
			break
		}
	}
	return tokens
}

func (l *Lexer) NextToken() Token {
	if l.position >= len(l.lines) {
		return Token{Type: TokenEOF, Literal: "", Meta: ""}
	}

	line := l.lines[l.position]
	l.position++

	trimmed := strings.TrimSpace(line)

	// Code block end state check
	if l.inCodeBlock {
		if strings.HasPrefix(trimmed, "```") {
			l.inCodeBlock = false
			return Token{Type: TokenCodeFence, Literal: line}
		}
		return Token{Type: TokenCodeLine, Literal: line}
	}

	// null
	if trimmed == "" {
		return Token{Type: TokenBlankLine, Literal: line}
	}

	// Code block start state check
	if strings.HasPrefix(trimmed, "```") {
		l.inCodeBlock = true
		language := strings.TrimPrefix(trimmed, "```")
		return Token{Type: TokenCodeFence, Literal: line, Meta: strings.TrimSpace(language)}
	}

	// Headings
	if strings.HasPrefix(trimmed, "#") {
		hashCount := 0
		for _, ch := range trimmed {
			if ch == '#' {
				hashCount++
			} else {
				break
			}
		}

		if hashCount > 0 && hashCount <= 6 && len(trimmed) > hashCount && trimmed[hashCount] == ' ' {
			return Token{Type: TokenHeading, Literal: line, Meta: strconv.Itoa(hashCount)} // level of heading
		}
	}

	// Blockquote
	if strings.HasPrefix(trimmed, ">") {
		return Token{Type: TokenBlockquote, Literal: line}
	}

	// List items
	if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") || strings.HasPrefix(trimmed, "+ ") {
		return Token{Type: TokenListItem, Literal: line}
	}

	return Token{Type: TokenTextLine, Literal: line}
}
