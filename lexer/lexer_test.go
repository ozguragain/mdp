package lexer

import (
	"strings"
	"testing"
)

// tokenize asserts the stream ends with exactly one EOF token.
func tokenize(t *testing.T, input string) []Token {
	t.Helper()
	tokens := NewLexer(input).Tokenize()
	if len(tokens) == 0 {
		t.Fatalf("Tokenize() returned no tokens, expected at least an EOF token")
	}
	last := tokens[len(tokens)-1]
	if last.Type != TokenEOF {
		t.Fatalf("last token is %s (%q), want EOF", last.Type, last.Literal)
	}
	for i, tok := range tokens[:len(tokens)-1] {
		if tok.Type == TokenEOF {
			t.Fatalf("unexpected EOF at index %d; EOF must only appear as the final token", i)
		}
	}
	return tokens
}

// FuzzTokenize guards the failure modes no finite E2E corpus can cover
// (AGENTS.md rule 5): panic, hang, and a malformed EOF stream.
func FuzzTokenize(f *testing.F) {
	seeds := []string{
		"",
		"\n",
		"# h\n```go\ncode\n```\n- l\n> q\n",
		"````\n``` \nunclosed",
		"#######\r\n#\r\n-\r\n>",
		strings.Repeat("```", 1000),
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, input string) {
		tokenize(t, input)
	})
}
