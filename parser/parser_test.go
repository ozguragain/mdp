package parser

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/ozguragain/mdp/ast"
	"github.com/ozguragain/mdp/lexer"
)

// parse runs the full lexer -> parser pipeline over src.
func parse(src string) *ast.DocumentNode {
	tokens := lexer.NewLexer(src).Tokenize()
	return NewParser(tokens).Parse()
}

// text is shorthand for the inline slice of a single-text block.
func text(value string) []ast.InlineNode {
	return []ast.InlineNode{&ast.TextNode{Value: value}}
}

// para is shorthand for a paragraph with one text node per argument.
func para(lines ...string) *ast.ParagraphNode {
	var inlines []ast.InlineNode
	for _, line := range lines {
		inlines = append(inlines, &ast.TextNode{Value: line})
	}
	return &ast.ParagraphNode{Inlines: inlines}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want []ast.BlockNode
	}{
		{
			name: "empty input produces no blocks",
			src:  "",
			want: nil,
		},
		{
			name: "blank lines only produce no blocks",
			src:  "\n\n",
			want: nil,
		},
		{
			name: "heading keeps level and content",
			src:  "# Title\n### Sub",
			want: []ast.BlockNode{
				&ast.HeadingNode{Level: 1, Inlines: text("Title")},
				&ast.HeadingNode{Level: 3, Inlines: text("Sub")},
			},
		},
		{
			name: "heading closing hashes are stripped",
			src:  "## closed heading ##",
			want: []ast.BlockNode{
				&ast.HeadingNode{Level: 2, Inlines: text("closed heading")},
			},
		},
		{
			name: "hash inside heading content is kept",
			src:  "# C# rocks",
			want: []ast.BlockNode{
				&ast.HeadingNode{Level: 1, Inlines: text("C# rocks")},
			},
		},
		{
			name: "indented heading keeps content only",
			src:  "   ### indented",
			want: []ast.BlockNode{
				&ast.HeadingNode{Level: 3, Inlines: text("indented")},
			},
		},
		{
			name: "consecutive text lines form one paragraph, one node per line",
			src:  "first line\nsecond line",
			want: []ast.BlockNode{para("first line", "second line")},
		},
		{
			name: "blank line splits paragraphs",
			src:  "one\n\ntwo",
			want: []ast.BlockNode{para("one"), para("two")},
		},
		{
			name: "fenced code block keeps raw info and lines",
			src:  "```go\nfmt.Println(\"hi\")\n```",
			want: []ast.BlockNode{
				&ast.FencedCodeBlockNode{Info: "go", Lines: []string{`fmt.Println("hi")`}},
			},
		},
		{
			name: "info string is kept raw for consumers",
			src:  "```js {hl_lines=[1]}\ncode\n```",
			want: []ast.BlockNode{
				&ast.FencedCodeBlockNode{Info: "js {hl_lines=[1]}", Lines: []string{"code"}},
			},
		},
		{
			name: "unclosed fence runs to EOF",
			src:  "```\na := 1\nnever closed",
			want: []ast.BlockNode{
				&ast.FencedCodeBlockNode{Lines: []string{"a := 1", "never closed"}},
			},
		},
		{
			name: "empty code block has no lines",
			src:  "```\n```",
			want: []ast.BlockNode{&ast.FencedCodeBlockNode{}},
		},
		{
			name: "block markers inside code are protected",
			src:  "```\n# not a heading\n- not a list\n```",
			want: []ast.BlockNode{
				&ast.FencedCodeBlockNode{Lines: []string{"# not a heading", "- not a list"}},
			},
		},
		{
			name: "blockquote content is parsed as blocks",
			src:  "> # head\n> text",
			want: []ast.BlockNode{
				&ast.BlockquoteNode{Blocks: []ast.BlockNode{
					&ast.HeadingNode{Level: 1, Inlines: text("head")},
					para("text"),
				}},
			},
		},
		{
			name: "nested quote markers nest containers",
			src:  ">> deep",
			want: []ast.BlockNode{
				&ast.BlockquoteNode{Blocks: []ast.BlockNode{
					&ast.BlockquoteNode{Blocks: []ast.BlockNode{para("deep")}},
				}},
			},
		},
		{
			name: "consecutive list items form one list",
			src:  "- a\n* b\n+ c",
			want: []ast.BlockNode{
				&ast.ListNode{Items: []*ast.ListItemNode{
					{Blocks: []ast.BlockNode{para("a")}},
					{Blocks: []ast.BlockNode{para("b")}},
					{Blocks: []ast.BlockNode{para("c")}},
				}},
			},
		},
		{
			name: "deeper indented items nest inside the previous item",
			src:  "- top\n  - nested",
			want: []ast.BlockNode{
				&ast.ListNode{Items: []*ast.ListItemNode{
					{Blocks: []ast.BlockNode{
						para("top"),
						&ast.ListNode{Items: []*ast.ListItemNode{
							{Blocks: []ast.BlockNode{para("nested")}},
						}},
					}},
				}},
			},
		},
		{
			name: "indented continuation lines stay inside the item",
			src:  "- item one\n  continued line\n- item two",
			want: []ast.BlockNode{
				&ast.ListNode{Items: []*ast.ListItemNode{
					{Blocks: []ast.BlockNode{para("item one", "continued line")}},
					{Blocks: []ast.BlockNode{para("item two")}},
				}},
			},
		},
		{
			name: "fence inside an item keeps the list together",
			src:  "- example\n  ```\n  x := 1\n\n  y := 2\n  ```",
			want: []ast.BlockNode{
				&ast.ListNode{Items: []*ast.ListItemNode{
					{Blocks: []ast.BlockNode{
						para("example"),
						&ast.FencedCodeBlockNode{Lines: []string{"x := 1", "", "y := 2"}},
					}},
				}},
			},
		},
		{
			name: "item fence may close at column zero",
			src:  "- a\n  ```\n  x\n```",
			want: []ast.BlockNode{
				&ast.ListNode{Items: []*ast.ListItemNode{
					{Blocks: []ast.BlockNode{
						para("a"),
						&ast.FencedCodeBlockNode{Lines: []string{"x"}},
					}},
				}},
			},
		},
		{
			name: "blank line ends a list",
			src:  "- a\n\n- b",
			want: []ast.BlockNode{
				&ast.ListNode{Items: []*ast.ListItemNode{{Blocks: []ast.BlockNode{para("a")}}}},
				&ast.ListNode{Items: []*ast.ListItemNode{{Blocks: []ast.BlockNode{para("b")}}}},
			},
		},
		{
			name: "mixed document",
			src:  "# Doc\n\nintro text\n\n- item\n\n```go\ncode\n```\n\n> quote\n",
			want: []ast.BlockNode{
				&ast.HeadingNode{Level: 1, Inlines: text("Doc")},
				para("intro text"),
				&ast.ListNode{Items: []*ast.ListItemNode{{Blocks: []ast.BlockNode{para("item")}}}},
				&ast.FencedCodeBlockNode{Info: "go", Lines: []string{"code"}},
				&ast.BlockquoteNode{Blocks: []ast.BlockNode{para("quote")}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parse(tt.src)
			want := &ast.DocumentNode{Blocks: tt.want}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("parse(%q) mismatch:\n got:\n%s\nwant:\n%s",
					tt.src, formatDoc(got), formatDoc(want))
			}
		})
	}
}

// formatDoc renders an AST as an indented outline for failure messages,
// mirroring the formatTokens convention in the lexer tests.
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
		}
	}
}

func formatInlines(inlines []ast.InlineNode) string {
	parts := make([]string, 0, len(inlines))
	for _, inline := range inlines {
		if t, ok := inline.(*ast.TextNode); ok {
			parts = append(parts, strconv.Quote(t.Value))
		}
	}
	return strings.Join(parts, " + ")
}
