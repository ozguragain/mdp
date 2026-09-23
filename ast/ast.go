// Package ast defines the markdown AST node types produced by the parser.
package ast

// BlockNode is the common interface for block-level nodes in the document
// tree. It is sealed: only types in this package can implement it.
type BlockNode interface {
	blockNode()
}

// InlineNode is the common interface for inline nodes within block content.
// It is sealed: only types in this package can implement it.
type InlineNode interface {
	inlineNode()
}

// DocumentNode is the root of a parsed markdown document.
type DocumentNode struct {
	Blocks []BlockNode
}

// HeadingNode represents an ATX heading (`# ...` through `###### ...`).
type HeadingNode struct {
	Level   int // 1..6
	Inlines []InlineNode
}

// ParagraphNode represents a paragraph built from consecutive text lines.
type ParagraphNode struct {
	Inlines []InlineNode
}

// FencedCodeBlockNode represents a backtick-fenced code block.
type FencedCodeBlockNode struct {
	Info  string   // raw info string ("js", "js {hl_lines=[1]}"); may be empty
	Lines []string // raw code lines, excluding the fence lines
}

// BlockquoteNode represents a block quote (container).
type BlockquoteNode struct {
	Blocks []BlockNode
}

// ListNode represents a list built from consecutive list items.
type ListNode struct {
	Items []*ListItemNode
}

// ListItemNode represents a single list item (container).
type ListItemNode struct {
	Blocks []BlockNode
}

// EmphasisNode wraps inline children in emphasis (`*x*` / `_x_`).
type EmphasisNode struct {
	Children []InlineNode
}

// StrongNode wraps inline children in strong emphasis (`**x**` / `__x__`).
type StrongNode struct {
	Children []InlineNode
}

// CodeNode is a code span; Value is literal text, never further parsed.
type CodeNode struct {
	Value string
}

// LinkNode is an inline link `[text](dest "title")`. Destination is always
// URL-sanitized (unsafe links never become a LinkNode).
type LinkNode struct {
	Destination string       // raw destination as written
	Title       string       // optional; "" when absent
	Children    []InlineNode // label, recursively parsed; nested links suppressed
}

// TextNode carries text. Before the inline phase it holds raw, unparsed
// source text (one node per source line); after inline.Process it holds a
// literal run of text in which "\n" acts as a soft line break.
type TextNode struct {
	Value string
}

func (*HeadingNode) blockNode()         {}
func (*ParagraphNode) blockNode()       {}
func (*FencedCodeBlockNode) blockNode() {}
func (*BlockquoteNode) blockNode()      {}
func (*ListNode) blockNode()            {}
func (*ListItemNode) blockNode()        {}

func (*TextNode) inlineNode()     {}
func (*EmphasisNode) inlineNode() {}
func (*StrongNode) inlineNode()   {}
func (*CodeNode) inlineNode()     {}
func (*LinkNode) inlineNode()     {}
