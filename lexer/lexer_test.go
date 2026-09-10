package lexer

import (
	"fmt"
	"strings"
	"testing"
)

// tokensEqual compares two token slices and returns a human readable
// diff message when they differ.
func tokensEqual(got, want []Token) string {
	if len(got) != len(want) {
		return fmt.Sprintf("token count mismatch:\n got %d: %s\nwant %d: %s",
			len(got), formatTokens(got), len(want), formatTokens(want))
	}
	for i := range got {
		g, w := got[i], want[i]
		if g.Type != w.Type || g.Literal != w.Literal || g.Meta != w.Meta {
			return fmt.Sprintf("token[%d] mismatch:\n got: {Type: %s, Literal: %q, Meta: %q}\nwant: {Type: %s, Literal: %q, Meta: %q}",
				i, g.Type, g.Literal, g.Meta, w.Type, w.Literal, w.Meta)
		}
	}
	return ""
}

func formatTokens(tokens []Token) string {
	var b strings.Builder
	for i, t := range tokens {
		fmt.Fprintf(&b, "{%s %q meta=%q}", t.Type, t.Literal, t.Meta)
		if i < len(tokens)-1 {
			b.WriteString(", ")
		}
	}
	return b.String()
}

// tokenize runs the lexer over input and asserts that the token stream
// always terminates with exactly one EOF token.
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

// runTokenTests executes a table of input -> expected tokens cases.
func runTokenTests(t *testing.T, cases []struct {
	name  string
	input string
	want  []Token
}) {
	t.Helper()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tokenize(t, tc.input)
			if msg := tokensEqual(got, tc.want); msg != "" {
				t.Error(msg)
			}
		})
	}
}

func TestTokenizeEmptyInput(t *testing.T) {
	runTokenTests(t, []struct {
		name  string
		input string
		want  []Token
	}{
		{
			name:  "empty input yields only EOF",
			input: "",
			want: []Token{
				{Type: TokenEOF},
			},
		},
		{
			name:  "single newline yields one blank line then EOF",
			input: "\n",
			want: []Token{
				{Type: TokenBlankLine, Literal: ""},
				{Type: TokenEOF},
			},
		},
		{
			name:  "trailing newline does not emit an extra blank line",
			input: "text\n",
			want: []Token{
				{Type: TokenTextLine, Literal: "text"},
				{Type: TokenEOF},
			},
		},
		{
			name:  "double trailing newline keeps one real blank line",
			input: "text\n\n",
			want: []Token{
				{Type: TokenTextLine, Literal: "text"},
				{Type: TokenBlankLine, Literal: ""},
				{Type: TokenEOF},
			},
		},
	})
}

func TestTokenizeHeadings(t *testing.T) {
	runTokenTests(t, []struct {
		name  string
		input string
		want  []Token
	}{
		{
			name:  "h1 through h6 are recognized with correct level in Meta",
			input: "# h1\n## h2\n### h3\n#### h4\n##### h5\n###### h6",
			want: []Token{
				{Type: TokenHeading, Literal: "# h1", Meta: "1"},
				{Type: TokenHeading, Literal: "## h2", Meta: "2"},
				{Type: TokenHeading, Literal: "### h3", Meta: "3"},
				{Type: TokenHeading, Literal: "#### h4", Meta: "4"},
				{Type: TokenHeading, Literal: "##### h5", Meta: "5"},
				{Type: TokenHeading, Literal: "###### h6", Meta: "6"},
				{Type: TokenEOF},
			},
		},
		{
			name:  "seven hashes is not a heading",
			input: "####### seven",
			want: []Token{
				{Type: TokenTextLine, Literal: "####### seven"},
				{Type: TokenEOF},
			},
		},
		{
			name:  "hash without trailing space is text",
			input: "#nospace",
			want: []Token{
				{Type: TokenTextLine, Literal: "#nospace"},
				{Type: TokenEOF},
			},
		},
		{
			name:  "lone hash without space or content is text",
			input: "#",
			want: []Token{
				{Type: TokenTextLine, Literal: "#"},
				{Type: TokenEOF},
			},
		},
		{
			name:  "trailing hashes still produce a heading",
			input: "## closed heading ##",
			want: []Token{
				{Type: TokenHeading, Literal: "## closed heading ##", Meta: "2"},
				{Type: TokenEOF},
			},
		},
		{
			name:  "indented heading keeps untrimmed literal but detects level",
			input: "   ### indented heading",
			want: []Token{
				{Type: TokenHeading, Literal: "   ### indented heading", Meta: "3"},
				{Type: TokenEOF},
			},
		},
		{
			name:  "hash followed by tab is not a heading",
			input: "#\ttabbed",
			want: []Token{
				{Type: TokenTextLine, Literal: "#\ttabbed"},
				{Type: TokenEOF},
			},
		},
	})
}

func TestTokenizeCodeBlocks(t *testing.T) {
	runTokenTests(t, []struct {
		name  string
		input string
		want  []Token
	}{
		{
			name:  "fenced block with language",
			input: "```go\nfmt.Println(\"hi\")\n```",
			want: []Token{
				{Type: TokenCodeFence, Literal: "```go", Meta: "go"},
				{Type: TokenCodeLine, Literal: "fmt.Println(\"hi\")"},
				{Type: TokenCodeFence, Literal: "```", Meta: ""},
				{Type: TokenEOF},
			},
		},
		{
			name:  "fence without language has empty meta",
			input: "```\nplain code\n```",
			want: []Token{
				{Type: TokenCodeFence, Literal: "```", Meta: ""},
				{Type: TokenCodeLine, Literal: "plain code"},
				{Type: TokenCodeFence, Literal: "```", Meta: ""},
				{Type: TokenEOF},
			},
		},
		{
			name:  "four backtick fence keeps full info string in meta",
			input: "````go\ncode\n````",
			want: []Token{
				{Type: TokenCodeFence, Literal: "````go", Meta: "go"},
				{Type: TokenCodeLine, Literal: "code"},
				{Type: TokenCodeFence, Literal: "````", Meta: ""},
				{Type: TokenEOF},
			},
		},
		{
			name:  "info string with attributes is preserved in meta",
			input: "```js {hl_lines=[1]}\ncode\n```",
			want: []Token{
				{Type: TokenCodeFence, Literal: "```js {hl_lines=[1]}", Meta: "js {hl_lines=[1]}"},
				{Type: TokenCodeLine, Literal: "code"},
				{Type: TokenCodeFence, Literal: "```", Meta: ""},
				{Type: TokenEOF},
			},
		},
		{
			name:  "markdown syntax inside code block is protected",
			input: "```\n# not a heading\n- not a list\n> not a quote\n```",
			want: []Token{
				{Type: TokenCodeFence, Literal: "```", Meta: ""},
				{Type: TokenCodeLine, Literal: "# not a heading"},
				{Type: TokenCodeLine, Literal: "- not a list"},
				{Type: TokenCodeLine, Literal: "> not a quote"},
				{Type: TokenCodeFence, Literal: "```", Meta: ""},
				{Type: TokenEOF},
			},
		},
		{
			name:  "unclosed code block consumes rest of input",
			input: "```go\na := 1\nnever closed",
			want: []Token{
				{Type: TokenCodeFence, Literal: "```go", Meta: "go"},
				{Type: TokenCodeLine, Literal: "a := 1"},
				{Type: TokenCodeLine, Literal: "never closed"},
				{Type: TokenEOF},
			},
		},
		{
			name:  "two sequential code blocks",
			input: "```\nfirst\n```\n\n```\nsecond\n```",
			want: []Token{
				{Type: TokenCodeFence, Literal: "```", Meta: ""},
				{Type: TokenCodeLine, Literal: "first"},
				{Type: TokenCodeFence, Literal: "```", Meta: ""},
				{Type: TokenBlankLine, Literal: ""},
				{Type: TokenCodeFence, Literal: "```", Meta: ""},
				{Type: TokenCodeLine, Literal: "second"},
				{Type: TokenCodeFence, Literal: "```", Meta: ""},
				{Type: TokenEOF},
			},
		},
		{
			name:  "indented fence opens a code block",
			input: "  ```go\ncode\n  ```",
			want: []Token{
				{Type: TokenCodeFence, Literal: "  ```go", Meta: "go"},
				{Type: TokenCodeLine, Literal: "code"},
				{Type: TokenCodeFence, Literal: "  ```", Meta: ""},
				{Type: TokenEOF},
			},
		},
		{
			name:  "blank line inside code block stays a code line",
			input: "```\n\nstill code\n```",
			want: []Token{
				{Type: TokenCodeFence, Literal: "```", Meta: ""},
				{Type: TokenCodeLine, Literal: ""},
				{Type: TokenCodeLine, Literal: "still code"},
				{Type: TokenCodeFence, Literal: "```", Meta: ""},
				{Type: TokenEOF},
			},
		},
	})
}

func TestTokenizeLists(t *testing.T) {
	runTokenTests(t, []struct {
		name  string
		input string
		want  []Token
	}{
		{
			name:  "dash, asterisk and plus markers",
			input: "- dash\n* star\n+ plus",
			want: []Token{
				{Type: TokenListItem, Literal: "- dash"},
				{Type: TokenListItem, Literal: "* star"},
				{Type: TokenListItem, Literal: "+ plus"},
				{Type: TokenEOF},
			},
		},
		{
			name:  "marker without space is text",
			input: "-nospace\n*star\n+plus",
			want: []Token{
				{Type: TokenTextLine, Literal: "-nospace"},
				{Type: TokenTextLine, Literal: "*star"},
				{Type: TokenTextLine, Literal: "+plus"},
				{Type: TokenEOF},
			},
		},
		{
			name:  "lone marker without space is text",
			input: "-",
			want: []Token{
				{Type: TokenTextLine, Literal: "-"},
				{Type: TokenEOF},
			},
		},
		{
			name:  "nested list item keeps original indentation in literal",
			input: "- top\n  - nested",
			want: []Token{
				{Type: TokenListItem, Literal: "- top"},
				{Type: TokenListItem, Literal: "  - nested"},
				{Type: TokenEOF},
			},
		},
	})
}

func TestTokenizeBlockquotes(t *testing.T) {
	runTokenTests(t, []struct {
		name  string
		input string
		want  []Token
	}{
		{
			name:  "blockquote with space",
			input: "> quoted text",
			want: []Token{
				{Type: TokenBlockquote, Literal: "> quoted text"},
				{Type: TokenEOF},
			},
		},
		{
			name:  "blockquote marker alone",
			input: ">",
			want: []Token{
				{Type: TokenBlockquote, Literal: ">"},
				{Type: TokenEOF},
			},
		},
		{
			name:  "blockquote without space after marker",
			input: ">no space",
			want: []Token{
				{Type: TokenBlockquote, Literal: ">no space"},
				{Type: TokenEOF},
			},
		},
		{
			name:  "nested blockquote markers",
			input: ">> deep quote",
			want: []Token{
				{Type: TokenBlockquote, Literal: ">> deep quote"},
				{Type: TokenEOF},
			},
		},
	})
}

func TestTokenizeBlankLinesAndText(t *testing.T) {
	runTokenTests(t, []struct {
		name  string
		input string
		want  []Token
	}{
		{
			name:  "whitespace-only lines are blank lines with original literal",
			input: "text\n   \n\t\nmore",
			want: []Token{
				{Type: TokenTextLine, Literal: "text"},
				{Type: TokenBlankLine, Literal: "   "},
				{Type: TokenBlankLine, Literal: "\t"},
				{Type: TokenTextLine, Literal: "more"},
				{Type: TokenEOF},
			},
		},
		{
			name:  "crlf line endings are normalized",
			input: "# Title\r\nbody\r\n\r\n- item",
			want: []Token{
				{Type: TokenHeading, Literal: "# Title", Meta: "1"},
				{Type: TokenTextLine, Literal: "body"},
				{Type: TokenBlankLine, Literal: ""},
				{Type: TokenListItem, Literal: "- item"},
				{Type: TokenEOF},
			},
		},
		{
			name:  "plain text lines pass through unchanged",
			input: "just some text\nwith more text",
			want: []Token{
				{Type: TokenTextLine, Literal: "just some text"},
				{Type: TokenTextLine, Literal: "with more text"},
				{Type: TokenEOF},
			},
		},
	})
}

func TestNextTokenAfterEOF(t *testing.T) {
	l := NewLexer("")
	first := l.NextToken()
	if first.Type != TokenEOF {
		t.Fatalf("first NextToken() = %s, want EOF", first.Type)
	}
	second := l.NextToken()
	if second.Type != TokenEOF {
		t.Fatalf("NextToken() past end = %s, want EOF to be returned repeatedly", second.Type)
	}
}

// FuzzTokenize guards against panics and non-termination on arbitrary input,
// and checks the stream invariants enforced by tokenize().
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

