// Package codegen renders a markdown AST to an HTML fragment.
package codegen

import (
	"fmt"
	"html"
	"strings"

	"github.com/ozguragain/mdp/ast"
)

// RenderHTML renders a document tree as an escaped HTML fragment: no
// wrapper, and each top-level block ends with a newline.
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

// renderInlines renders a block's inline nodes back to back (soft breaks
// live inside text values).
func renderInlines(inlines []ast.InlineNode) string {
	var b strings.Builder
	for _, inline := range inlines {
		renderInline(&b, inline)
	}
	return b.String()
}

// renderInline writes one node; all user text goes through html.EscapeString
// so nothing can break out of markup or an attribute.
func renderInline(b *strings.Builder, node ast.InlineNode) {
	switch n := node.(type) {
	case *ast.TextNode:
		b.WriteString(html.EscapeString(n.Value))
	case *ast.EmphasisNode:
		b.WriteString("<em>")
		b.WriteString(renderInlines(n.Children))
		b.WriteString("</em>")
	case *ast.StrongNode:
		b.WriteString("<strong>")
		b.WriteString(renderInlines(n.Children))
		b.WriteString("</strong>")
	case *ast.CodeNode:
		b.WriteString("<code>")
		b.WriteString(html.EscapeString(n.Value))
		b.WriteString("</code>")
	case *ast.LinkNode:
		fmt.Fprintf(b, "<a href=\"%s\"", html.EscapeString(n.Destination))
		if n.Title != "" {
			fmt.Fprintf(b, " title=\"%s\"", html.EscapeString(n.Title))
		}
		b.WriteString(">")
		b.WriteString(renderInlines(n.Children))
		b.WriteString("</a>")
	}
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

// renderListItem writes one <li>, tight style when the item is a single
// paragraph.
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

// firstWord takes the language from a raw info string
// ("js {hl_lines=[1]}" -> "js").
func firstWord(info string) string {
	fields := strings.Fields(info)
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}
