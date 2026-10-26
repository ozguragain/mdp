// Package codegen renders a markdown AST to an HTML fragment.
package codegen

import (
	"html"
	"strconv"
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
		level := strconv.Itoa(n.Level)
		b.WriteString("<h" + level + ">")
		b.WriteString(renderInlines(n.Inlines))
		b.WriteString("</h" + level + ">\n")
	case *ast.ParagraphNode:
		b.WriteString("<p>")
		b.WriteString(renderInlines(n.Inlines))
		b.WriteString("</p>\n")
	case *ast.FencedCodeBlockNode:
		renderCodeBlock(b, n)
	case *ast.BlockquoteNode:
		b.WriteString("<blockquote>\n")
		renderBlocks(b, n.Blocks)
		b.WriteString("</blockquote>\n")
	case *ast.ListNode:
		if n.Ordered {
			if n.Start != 1 {
				b.WriteString("<ol start=\"" + strconv.Itoa(n.Start) + "\">\n")
			} else {
				b.WriteString("<ol>\n")
			}
		} else {
			b.WriteString("<ul>\n")
		}
		for _, item := range n.Items {
			renderListItem(b, item)
		}
		if n.Ordered {
			b.WriteString("</ol>\n")
		} else {
			b.WriteString("</ul>\n")
		}
	case *ast.ThematicBreakNode:
		b.WriteString("<hr>\n")
	case *ast.TableNode:
		renderTable(b, n)
	}
}

// renderTable writes one <table>: a <thead> holding the header row, then a
// <tbody> holding the body rows. Empty rows are omitted entirely.
func renderTable(b *strings.Builder, n *ast.TableNode) {
	b.WriteString("<table>\n<thead>\n<tr>\n")
	for i, cell := range n.Header {
		writeCell(b, "th", cell, n.Alignments, i)
	}
	b.WriteString("</tr>\n</thead>\n")
	if len(n.Rows) > 0 {
		b.WriteString("<tbody>\n")
		for _, row := range n.Rows {
			b.WriteString("<tr>\n")
			for i, cell := range row {
				writeCell(b, "td", cell, n.Alignments, i)
			}
			b.WriteString("</tr>\n")
		}
		b.WriteString("</tbody>\n")
	}
	b.WriteString("</table>\n")
}

// writeCell writes one <th>/<td>, adding align= when the delimiter row asked
// for a specific alignment on that column.
func writeCell(b *strings.Builder, tag string, cell ast.TableCellNode, aligns []ast.Align, col int) {
	b.WriteString("<")
	b.WriteString(tag)
	if col < len(aligns) {
		switch aligns[col] {
		case ast.AlignLeft:
			b.WriteString(" align=\"left\"")
		case ast.AlignCenter:
			b.WriteString(" align=\"center\"")
		case ast.AlignRight:
			b.WriteString(" align=\"right\"")
		}
	}
	b.WriteString(">")
	b.WriteString(renderInlines(cell.Inlines))
	b.WriteString("</")
	b.WriteString(tag)
	b.WriteString(">\n")
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
		b.WriteString("<a href=\"" + html.EscapeString(n.Destination) + "\"")
		if n.Title != "" {
			b.WriteString(" title=\"" + html.EscapeString(n.Title) + "\"")
		}
		b.WriteString(">")
		b.WriteString(renderInlines(n.Children))
		b.WriteString("</a>")
	case *ast.ImageNode:
		b.WriteString("<img src=\"" + html.EscapeString(n.Destination) + "\" alt=\"" + html.EscapeString(n.Alt) + "\"")
		if n.Title != "" {
			b.WriteString(" title=\"" + html.EscapeString(n.Title) + "\"")
		}
		b.WriteString(">")
	case *ast.StrikethroughNode:
		b.WriteString("<del>")
		b.WriteString(renderInlines(n.Children))
		b.WriteString("</del>")
	case *ast.HardBreakNode:
		b.WriteString("<br>\n")
	case *ast.AutolinkNode:
		b.WriteString("<a href=\"" + html.EscapeString(n.Destination) + "\">" + html.EscapeString(n.Text) + "</a>")
	}
}

func renderCodeBlock(b *strings.Builder, n *ast.FencedCodeBlockNode) {
	if lang := firstWord(n.Info); lang != "" {
		b.WriteString("<pre><code class=\"language-" + html.EscapeString(lang) + "\">")
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
			b.WriteString("<li>" + renderInlines(para.Inlines) + "</li>\n")
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
