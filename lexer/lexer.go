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

	// Empty input produces no lines at all; only an EOF token will be emitted.
	var lines []string
	if input != "" {
		lines = strings.Split(input, "\n")
	}

	lines = strings.Split(input, "\n")
	if strings.HasSuffix(input, "\n") && len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	return &Lexer{
		lines:       lines,
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
