package inline

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/ozguragain/mdp/ast"
	"github.com/ozguragain/mdp/lexer"
	"github.com/ozguragain/mdp/parser"
)

// Constructors for expected inline trees, mirroring the parser_test style.
// (Named txt rather than text to stay clear of the *testing.T parameter.)
func txt(value string) ast.InlineNode { return &ast.TextNode{Value: value} }

func em(children ...ast.InlineNode) ast.InlineNode {
	return &ast.EmphasisNode{Children: children}
}

func strong(children ...ast.InlineNode) ast.InlineNode {
	return &ast.StrongNode{Children: children}
}

func code(value string) ast.InlineNode { return &ast.CodeNode{Value: value} }

func link(dest, title string, children ...ast.InlineNode) ast.InlineNode {
	return &ast.LinkNode{Destination: dest, Title: title, Children: children}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want []ast.InlineNode
	}{
		{
			name: "empty input produces no inlines",
			src:  "",
			want: nil,
		},
		{
			name: "plain text stays one text node",
			src:  "hello world",
			want: []ast.InlineNode{txt("hello world")},
		},
		{
			name: "asterisk emphasis",
			src:  "*x*",
			want: []ast.InlineNode{em(txt("x"))},
		},
		{
			name: "underscore emphasis",
			src:  "_x_",
			want: []ast.InlineNode{em(txt("x"))},
		},
		{
			name: "asterisk strong",
			src:  "**x**",
			want: []ast.InlineNode{strong(txt("x"))},
		},
		{
			name: "underscore strong",
			src:  "__x__",
			want: []ast.InlineNode{strong(txt("x"))},
		},
		{
			name: "triple run is em around strong",
			src:  "***x***",
			want: []ast.InlineNode{em(strong(txt("x")))},
		},
		{
			name: "quadruple run is literal",
			src:  "****x****",
			want: []ast.InlineNode{txt("****x****")},
		},
		{
			name: "intraword asterisk emphasis",
			src:  "un*frigging*believable",
			want: []ast.InlineNode{txt("un"), em(txt("frigging")), txt("believable")},
		},
		{
			name: "intraword underscore is literal",
			src:  "snake_case_name",
			want: []ast.InlineNode{txt("snake_case_name")},
		},
		{
			name: "two emphases in one line",
			src:  "*a* b *c*",
			want: []ast.InlineNode{em(txt("a")), txt(" b "), em(txt("c"))},
		},
		{
			name: "emphasis nested in strong",
			src:  "**a *b* c**",
			want: []ast.InlineNode{strong(txt("a "), em(txt("b")), txt(" c"))},
		},
		{
			name: "same delimiter nests via stack semantics",
			src:  "*a *b* c*",
			want: []ast.InlineNode{em(txt("a "), em(txt("b")), txt(" c"))},
		},
		{
			name: "unequal run lengths stay literal",
			src:  "**foo*",
			want: []ast.InlineNode{txt("**foo*")},
		},
		{
			name: "unequal run lengths stay literal the other way",
			src:  "*foo**",
			want: []ast.InlineNode{txt("*foo**")},
		},
		{
			name: "spaced delimiters are literal",
			src:  "a * b * c",
			want: []ast.InlineNode{txt("a * b * c")},
		},
		{
			name: "unmatched opener is literal",
			src:  "*x",
			want: []ast.InlineNode{txt("*x")},
		},
		{
			name: "code span shields emphasis",
			src:  "`x*y`",
			want: []ast.InlineNode{code("x*y")},
		},
		{
			name: "double backtick code span holds one backtick",
			src:  "``a`b``",
			want: []ast.InlineNode{code("a`b")},
		},
		{
			name: "code span strips one space on each side",
			src:  "` x `",
			want: []ast.InlineNode{code("x")},
		},
		{
			name: "unmatched backtick run is literal",
			src:  "`x",
			want: []ast.InlineNode{txt("`x")},
		},
		{
			name: "simple link",
			src:  "[t](/u)",
			want: []ast.InlineNode{link("/u", "", txt("t"))},
		},
		{
			name: "link with title",
			src:  `[t](/u "T")`,
			want: []ast.InlineNode{link("/u", "T", txt("t"))},
		},
		{
			name: "emphasis inside link label",
			src:  "[*t*](/u)",
			want: []ast.InlineNode{link("/u", "", em(txt("t")))},
		},
		{
			name: "balanced parentheses in destination",
			src:  "[t](/a_(b))",
			want: []ast.InlineNode{link("/a_(b)", "", txt("t"))},
		},
		{
			name: "nested link syntax is suppressed",
			src:  "[a [b](/c) d](/e)",
			want: []ast.InlineNode{link("/e", "", txt("a [b](/c) d"))},
		},
		{
			name: "empty destination is a valid link",
			src:  "[a]()",
			want: []ast.InlineNode{link("", "", txt("a"))},
		},
		{
			name: "unclosed link is literal",
			src:  "[a](/b",
			want: []ast.InlineNode{txt("[a](/b")},
		},
		{
			name: "text after label is literal",
			src:  "[a] b",
			want: []ast.InlineNode{txt("[a] b")},
		},
		{
			name: "bare label is literal",
			src:  "[a]",
			want: []ast.InlineNode{txt("[a]")},
		},
		{
			name: "escapes shield delimiters",
			src:  `\*x\*`,
			want: []ast.InlineNode{txt("*x*")},
		},
		{
			name: "escaped backslash is one backslash",
			src:  `a\\b`,
			want: []ast.InlineNode{txt(`a\b`)},
		},
		{
			name: "backslash before newline is literal and soft breaks",
			src:  "c\\\nd",
			want: []ast.InlineNode{txt("c\\\nd")},
		},
		{
			name: "emphasized label survives an unsafe URL",
			src:  "[**x**](javascript:alert(1))",
			want: []ast.InlineNode{strong(txt("x"))},
		},
		{
			name: "javascript URL unwraps to label",
			src:  "[x](javascript:alert(1))",
			want: []ast.InlineNode{txt("x")},
		},
		{
			name: "case mangled javascript URL unwraps to label",
			src:  "[x](JaVaScRiPt:x)",
			want: []ast.InlineNode{txt("x")},
		},
		{
			name: "data URL unwraps to label",
			src:  "[x](data:text/html,x)",
			want: []ast.InlineNode{txt("x")},
		},
		{
			name: "vbscript URL unwraps to label",
			src:  "[x](vbscript:x)",
			want: []ast.InlineNode{txt("x")},
		},
		{
			name: "entity encoded scheme unwraps to label",
			src:  "[x](&#106;avascript:x)",
			want: []ast.InlineNode{txt("x")},
		},
		{
			name: "fragment destination is safe",
			src:  "[x](#f)",
			want: []ast.InlineNode{link("#f", "", txt("x"))},
		},
		{
			name: "relative destination is safe",
			src:  "[x](/r)",
			want: []ast.InlineNode{link("/r", "", txt("x"))},
		},
		{
			name: "https destination is safe",
			src:  "[x](https://ok)",
			want: []ast.InlineNode{link("https://ok", "", txt("x"))},
		},
		{
			name: "mailto destination is safe",
			src:  "[x](mailto:a@b)",
			want: []ast.InlineNode{link("mailto:a@b", "", txt("x"))},
		},
		{
			name: "emphasis cannot close across escaped star",
			src:  "*a \\* b*",
			want: []ast.InlineNode{em(txt("a * b"))},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Parse(tt.src)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse(%q) mismatch:\n got: %s\nwant: %s",
					tt.src, formatInlines(got), formatInlines(tt.want))
			}
		})
	}
}

// parseSrc runs the full lexer -> parser -> inline pipeline over src.
func parseSrc(src string) *ast.DocumentNode {
	tokens := lexer.NewLexer(src).Tokenize()
	doc := parser.NewParser(tokens).Parse()
	Process(doc)
	return doc
}

func TestProcess(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want []ast.BlockNode
	}{
		{
			name: "paragraph lines join with a soft break",
			src:  "a\nb",
			want: []ast.BlockNode{
				&ast.ParagraphNode{Inlines: []ast.InlineNode{txt("a\nb")}},
			},
		},
		{
			name: "heading inlines are parsed",
			src:  "# *x*",
			want: []ast.BlockNode{
				&ast.HeadingNode{Level: 1, Inlines: []ast.InlineNode{em(txt("x"))}},
			},
		},
		{
			name: "blockquote content is processed recursively",
			src:  "> *x*",
			want: []ast.BlockNode{
				&ast.BlockquoteNode{Blocks: []ast.BlockNode{
					&ast.ParagraphNode{Inlines: []ast.InlineNode{em(txt("x"))}},
				}},
			},
		},
		{
			name: "list item content is processed recursively",
			src:  "- *x*",
			want: []ast.BlockNode{
				&ast.ListNode{Items: []*ast.ListItemNode{
					{Blocks: []ast.BlockNode{
						&ast.ParagraphNode{Inlines: []ast.InlineNode{em(txt("x"))}},
					}},
				}},
			},
		},
		{
			name: "fenced code block is left untouched",
			src:  "```\n*x*\n```",
			want: []ast.BlockNode{
				&ast.FencedCodeBlockNode{Lines: []string{"*x*"}},
			},
		},
		{
			name: "emphasis spans a soft break",
			src:  "*a\nb*",
			want: []ast.BlockNode{
				&ast.ParagraphNode{Inlines: []ast.InlineNode{em(txt("a\nb"))}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseSrc(tt.src)
			want := &ast.DocumentNode{Blocks: tt.want}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("process(%q) mismatch:\n got: %s\nwant: %s",
					tt.src, formatDoc(got), formatDoc(want))
			}
		})
	}
}

func TestIsSafeURL(t *testing.T) {
	tests := []struct {
		raw  string
		want bool
	}{
		{"/rel/path", true},
		{"#frag", true},
		{"?q=1", true},
		{"//evil.test/x", true}, // protocol-relative: scheme comes from the page
		{"foo/bar:baz", true},   // relative path, colon after "/"
		{"nocolon", true},
		{"https://ok", true},
		{"HTTP://ok", true},
		{"mailto:a@b", true},
		{"javascript:alert(1)", false},
		{"JaVaScRiPt:alert(1)", false},
		{"java\tscript:alert(1)", false},
		{"java script:alert(1)", false},
		{"&#106;avascript:alert(1)", false},
		{"data:text/html,x", false},
		{"vbscript:x", false},
		{"file:///etc/passwd", false},
		{"blob:https://ok", false},
		{"ftp://host/x", false}, // not in the allowlist
	}

	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			if got := isSafeURL(tt.raw); got != tt.want {
				t.Errorf("isSafeURL(%q) = %v, want %v", tt.raw, got, tt.want)
			}
		})
	}
}

// formatDoc renders an AST as an indented outline for failure messages,
// extending the formatTokens convention in the other packages.
func formatDoc(doc *ast.DocumentNode) string {
	var b strings.Builder
	b.WriteString("Document\n")
	formatBlocks(&b, doc.Blocks, 1)
	return b.String()
}

func formatBlocks(b *strings.Builder, blocks []ast.BlockNode, depth int) {
	pad := strings.Repeat("  ", depth)
	for _, block := range blocks {
		switch n := block.(type) {
		case *ast.HeadingNode:
			fmt.Fprintf(b, "%sHeading(%d) %s\n", pad, n.Level, formatInlines(n.Inlines))
		case *ast.ParagraphNode:
			fmt.Fprintf(b, "%sParagraph %s\n", pad, formatInlines(n.Inlines))
		case *ast.FencedCodeBlockNode:
			fmt.Fprintf(b, "%sCode(info=%q) %q\n", pad, n.Info, n.Lines)
		case *ast.BlockquoteNode:
			b.WriteString(pad + "Blockquote\n")
			formatBlocks(b, n.Blocks, depth+1)
		case *ast.ListNode:
			b.WriteString(pad + "List\n")
			for _, item := range n.Items {
				b.WriteString(pad + "  Item\n")
				formatBlocks(b, item.Blocks, depth+2)
			}
		case *ast.ListItemNode:
			b.WriteString(pad + "Item\n")
			formatBlocks(b, n.Blocks, depth+1)
		}
	}
}

// formatInlines renders inline nodes as a flat outline for failure messages.
func formatInlines(inlines []ast.InlineNode) string {
	parts := make([]string, 0, len(inlines))
	for _, in := range inlines {
		parts = append(parts, formatInline(in))
	}
	return strings.Join(parts, " + ")
}

func formatInline(in ast.InlineNode) string {
	switch n := in.(type) {
	case *ast.TextNode:
		return strconv.Quote(n.Value)
	case *ast.EmphasisNode:
		return "Em(" + formatInlines(n.Children) + ")"
	case *ast.StrongNode:
		return "Strong(" + formatInlines(n.Children) + ")"
	case *ast.CodeNode:
		return "Code(" + strconv.Quote(n.Value) + ")"
	case *ast.LinkNode:
		return fmt.Sprintf("Link(dest=%q title=%q %s)", n.Destination, n.Title, formatInlines(n.Children))
	}
	return "?"
}
