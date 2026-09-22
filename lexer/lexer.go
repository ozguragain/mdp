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

	lines := strings.Split(input, "\n")
	// Split always yields at least one element; the last one is "" only for
	// empty input or a trailing newline. In both cases it is a ghost that
	// must not become a blank line token.
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	return &Lexer{lines: lines}
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
		return Token{Type: TokenEOF}
	}

	line := l.lines[l.position]
	l.position++

	trimmed := strings.TrimSpace(line)

	// Code block end state check
	if l.inCodeBlock {
		if backtickRun(trimmed) >= 3 {
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
	// A fenced code block opens with a run of 3 or more backticks (CommonMark).
	if btCount := backtickRun(trimmed); btCount >= 3 {
		l.inCodeBlock = true
		language := strings.TrimSpace(trimmed[btCount:])
		return Token{Type: TokenCodeFence, Literal: line, Meta: language}
	}

	// Headings
	if strings.HasPrefix(trimmed, "#") {
		hashCount := len(trimmed) - len(strings.TrimLeft(trimmed, "#"))
		if hashCount <= 6 && len(trimmed) > hashCount && trimmed[hashCount] == ' ' {
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

// backtickRun returns the number of leading backtick characters in s.
func backtickRun(s string) int {
	n := 0
	for _, ch := range s {
		if ch != '`' {
			break
		}
		n++
	}
	return n
}
