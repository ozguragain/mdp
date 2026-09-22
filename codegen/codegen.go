// Package codegen renders a markdown AST to an HTML fragment.
package codegen

import (
	"fmt"
	"html"
	"strings"

	"github.com/ozguragain/mdp/ast"
)

// RenderHTML converts a document tree into an HTML fragment string.
// All text is escaped; no <html>/<body> wrapper is produced. Each
// top-level block ends with a newline.
func RenderHTML(doc *ast.DocumentNode) string {
	var b strings.Builder
	renderBlocks(&b, doc.Blocks)
	return b.String()
}

func renderBlocks(b *strings.Builder, blocks []ast.BlockNode) {
	for _, block := range blocks {
		renderBlock(b, block)
	}
}

func renderBlock(b *strings.Builder, block ast.BlockNode) {
	switch n := block.(type) {
	case *ast.HeadingNode:
		fmt.Fprintf(b, "<h%d>%s</h%d>\n", n.Level, renderInlines(n.Inlines), n.Level)
	case *ast.ParagraphNode:
		fmt.Fprintf(b, "<p>%s</p>\n", renderInlines(n.Inlines))
	case *ast.FencedCodeBlockNode:
		renderCodeBlock(b, n)
	case *ast.BlockquoteNode:
		b.WriteString("<blockquote>\n")
		renderBlocks(b, n.Blocks)
		b.WriteString("</blockquote>\n")
	case *ast.ListNode:
		b.WriteString("<ul>\n")
		for _, item := range n.Items {
			renderListItem(b, item)
		}
		b.WriteString("</ul>\n")
	}
}

// renderInlines joins the inline nodes of a block. Today the parser only
// produces TextNodes (one per source line); the newlines between them act
// as soft breaks in HTML.
func renderInlines(inlines []ast.InlineNode) string {
	parts := make([]string, 0, len(inlines))
	for _, inline := range inlines {
		if text, ok := inline.(*ast.TextNode); ok {
			parts = append(parts, html.EscapeString(text.Value))
		}
	}
	return strings.Join(parts, "\n")
}

func renderCodeBlock(b *strings.Builder, n *ast.FencedCodeBlockNode) {
	if lang := firstWord(n.Info); lang != "" {
		fmt.Fprintf(b, "<pre><code class=\"language-%s\">", html.EscapeString(lang))
	} else {
		b.WriteString("<pre><code>")
	}
	if len(n.Lines) > 0 {
		b.WriteString(html.EscapeString(strings.Join(n.Lines, "\n")))
		b.WriteString("\n")
	}
	b.WriteString("</code></pre>\n")
}

// renderListItem writes one <li>. An item whose content is a single
// paragraph renders its text directly (tight list style); anything richer
// renders its full block tree.
func renderListItem(b *strings.Builder, item *ast.ListItemNode) {
	if len(item.Blocks) == 1 {
		if para, ok := item.Blocks[0].(*ast.ParagraphNode); ok {
			fmt.Fprintf(b, "<li>%s</li>\n", renderInlines(para.Inlines))
			return
		}
	}
	b.WriteString("<li>\n")
	renderBlocks(b, item.Blocks)
	b.WriteString("</li>\n")
}

// firstWord extracts the language name from a raw info string
// ("js {hl_lines=[1]}" -> "js"). Highlighting attributes are not
// interpreted yet; that is a future concern.
func firstWord(info string) string {
	fields := strings.Fields(info)
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}
