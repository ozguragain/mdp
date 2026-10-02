package lexer

import (
	"strconv"
	"strings"

	"github.com/ozguragain/mdp/ast"
)

type Lexer struct {
	lines       []string
	lineStarts  []int // byte offset of each line start in the normalized input
	position    int
	inCodeBlock bool
}

func NewLexer(input string) *Lexer {
	input = strings.ReplaceAll(input, "\r\n", "\n")

	lines := strings.Split(input, "\n")
	// The split's last "" (empty input or trailing newline) is a ghost line
	// that must not become a blank line token.
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	starts := make([]int, len(lines))
	offset := 0
	for i, line := range lines {
		starts[i] = offset
		offset += len(line) + 1 // the newline
	}

	return &Lexer{lines: lines, lineStarts: starts}
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
		return Token{Type: TokenEOF, Span: l.eofSpan()}
	}

	line := l.lines[l.position]
	l.position++
	span := l.lineSpan()

	trimmed := strings.TrimSpace(line)

	if l.inCodeBlock {
		if backtickRun(trimmed) >= 3 {
			l.inCodeBlock = false
			return Token{Type: TokenCodeFence, Literal: line, Span: span}
		}
		return Token{Type: TokenCodeLine, Literal: line, Span: span}
	}

	if trimmed == "" {
		return Token{Type: TokenBlankLine, Literal: line, Span: span}
	}

	// A fenced code block opens with a run of 3 or more backticks (CommonMark).
	if btCount := backtickRun(trimmed); btCount >= 3 {
		l.inCodeBlock = true
		language := strings.TrimSpace(trimmed[btCount:])
		return Token{Type: TokenCodeFence, Literal: line, Meta: language, Span: span}
	}

	if strings.HasPrefix(trimmed, "#") {
		hashCount := len(trimmed) - len(strings.TrimLeft(trimmed, "#"))
		if hashCount <= 6 && len(trimmed) > hashCount && trimmed[hashCount] == ' ' {
			return Token{Type: TokenHeading, Literal: line, Meta: strconv.Itoa(hashCount), Span: span} // level of heading
		}
	}

	if strings.HasPrefix(trimmed, ">") {
		return Token{Type: TokenBlockquote, Literal: line, Span: span}
	}

	if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") || strings.HasPrefix(trimmed, "+ ") {
		return Token{Type: TokenListItem, Literal: line, Span: span}
	}

	return Token{Type: TokenTextLine, Literal: line, Span: span}
}

// lineSpan returns the span of the line the cursor just moved past,
// excluding its newline.
func (l *Lexer) lineSpan() ast.Span {
	i := l.position - 1
	line := l.lines[i]
	start := ast.Pos{Offset: l.lineStarts[i], Line: i + 1, Column: 1}
	end := ast.Pos{Offset: start.Offset + len(line), Line: i + 1, Column: len(line) + 1}
	return ast.Span{Start: start, End: end}
}

// eofSpan is a zero-width span at the end of the input.
func (l *Lexer) eofSpan() ast.Span {
	if len(l.lines) == 0 {
		p := ast.Pos{Offset: 0, Line: 1, Column: 1}
		return ast.Span{Start: p, End: p}
	}
	i := len(l.lines) - 1
	last := l.lines[i]
	p := ast.Pos{Offset: l.lineStarts[i] + len(last), Line: i + 1, Column: len(last) + 1}
	return ast.Span{Start: p, End: p}
}

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
