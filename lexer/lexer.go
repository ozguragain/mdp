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

	inTable       bool
	needDelimiter bool // the delimiter row follows a freshly opened table

	// fm is the frontmatter token, emitted as the very first token and only
	// by NewDocumentLexer.
	fm *Token
}

// NewLexer lexes a markdown fragment — the content of a container, or a
// document whose frontmatter someone else already dealt with. A leading
// "---" fence is never read as frontmatter here.
func NewLexer(input string) *Lexer {
	return newLexer(input, false)
}

// NewDocumentLexer lexes a whole document. When the source opens with a
// frontmatter block ("---" ... "---" or "..."), that block is emitted as a
// single TokenFrontmatter before any block token.
func NewDocumentLexer(input string) *Lexer {
	return newLexer(input, true)
}

func newLexer(input string, document bool) *Lexer {
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

	l := &Lexer{lines: lines, lineStarts: starts}
	if document {
		l.fm = scanFrontmatter(lines, starts)
	}
	return l
}

// scanFrontmatter reads a frontmatter block at the head of the document.
// Without a closing fence there is no frontmatter — the leading "---" stays a
// thematic break and is lexed normally.
func scanFrontmatter(lines []string, starts []int) *Token {
	if len(lines) < 2 || lines[0] != "---" {
		return nil
	}
	for i := 1; i < len(lines); i++ {
		t := strings.TrimSpace(lines[i])
		if t != "---" && t != "..." {
			continue
		}
		start := ast.Pos{Offset: starts[0], Line: 1, Column: 1}
		last := lines[i]
		end := ast.Pos{Offset: starts[i] + len(last), Line: i + 1, Column: len(last) + 1}
		return &Token{
			Type:    TokenFrontmatter,
			Literal: strings.Join(lines[1:i], "\n"),
			Span:    ast.Span{Start: start, End: end},
		}
	}
	return nil
}

func (l *Lexer) Tokenize() []Token {
	var tokens []Token
	for {
		token := l.nextToken()
		tokens = append(tokens, token)

		if token.Type == TokenEOF {
			break
		}
	}
	return tokens
}

func (l *Lexer) nextToken() Token {
	if l.fm != nil {
		t := *l.fm
		l.fm = nil
		l.skipPastFrontmatter(t)
		return t
	}

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
		l.inTable = false
		return Token{Type: TokenBlankLine, Literal: line, Span: span}
	}

	// A table that is already open keeps claiming lines with a pipe, whatever
	// block markers they begin with; anything else ends it.
	if l.inTable {
		if l.needDelimiter {
			l.needDelimiter = false
			return Token{Type: TokenTableDelimiter, Literal: line, Meta: line, Span: span}
		}
		if strings.Contains(trimmed, "|") {
			return Token{Type: TokenTableRow, Literal: line, Span: span}
		}
		l.inTable = false
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

	// Thematic break before the list markers: "- - -" is a break, not a list.
	if isThematicBreak(trimmed) {
		return Token{Type: TokenThematicBreak, Literal: line, Span: span}
	}

	if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") || strings.HasPrefix(trimmed, "+ ") {
		return Token{Type: TokenListItem, Literal: line, Span: span}
	}

	if marker := orderedMarker(trimmed); marker != "" {
		return Token{Type: TokenOrderedListItem, Literal: line, Meta: marker, Span: span}
	}

	// A table opens only when the next line is a real delimiter row, so a
	// paragraph that merely mentions "a | b" stays a paragraph.
	if strings.Contains(trimmed, "|") && l.isTableDelimiterAt(l.position) {
		l.inTable = true
		l.needDelimiter = true
		return Token{Type: TokenTableHeader, Literal: line, Span: span}
	}

	return Token{Type: TokenTextLine, Literal: line, Span: span}
}

// skipPastFrontmatter moves the cursor past the block the token covered, so
// the next token is the first line after the closing fence.
func (l *Lexer) skipPastFrontmatter(t Token) {
	i := 0
	for i < len(l.lines) && l.lineStarts[i] < t.Span.End.Offset {
		i++
	}
	l.position = i
}

// isTableDelimiterAt reports whether the line at idx is a GFM delimiter row.
func (l *Lexer) isTableDelimiterAt(idx int) bool {
	if idx >= len(l.lines) {
		return false
	}
	return isTableDelimiter(l.lines[idx])
}

// isTableDelimiter reports whether one line is a GFM table delimiter row:
// cells of optional ":", one or more "-", optional ":", joined by pipes. An
// opening or closing pipe contributes no cell.
func isTableDelimiter(line string) bool {
	trimmed := strings.TrimSpace(line)
	if !strings.Contains(trimmed, "-") {
		return false
	}
	cells := 0
	for _, cell := range strings.Split(trimmed, "|") {
		c := strings.TrimSpace(cell)
		if c == "" {
			continue
		}
		cells++
		c = strings.TrimPrefix(c, ":")
		c = strings.TrimSuffix(c, ":")
		if c == "" || strings.Trim(c, "-") != "" {
			return false
		}
	}
	return cells > 0
}

// isThematicBreak reports whether a line is "---", "***" or "___": three or
// more of one of those characters with only spaces between them.
func isThematicBreak(trimmed string) bool {
	if trimmed == "" {
		return false
	}
	c := trimmed[0]
	if c != '-' && c != '*' && c != '_' {
		return false
	}
	n := 0
	for i := 0; i < len(trimmed); i++ {
		switch trimmed[i] {
		case c:
			n++
		case ' ':
		default:
			return false
		}
	}
	return n >= 3
}

// orderedMarker returns the marker when the line opens an ordered list item
// ("1." or "1)" followed by a space), else "".
func orderedMarker(trimmed string) string {
	i := 0
	for i < len(trimmed) && trimmed[i] >= '0' && trimmed[i] <= '9' {
		i++
	}
	if i == 0 || i > 9 || i+1 >= len(trimmed) {
		return ""
	}
	if trimmed[i] != '.' && trimmed[i] != ')' {
		return ""
	}
	if trimmed[i+1] != ' ' && trimmed[i+1] != '\t' {
		return ""
	}
	return trimmed[:i+1]
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
