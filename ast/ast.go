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

// TextNode carries a raw, not-yet-inline-parsed piece of text.
type TextNode struct {
	Value string
}

func (*HeadingNode) blockNode()         {}
func (*ParagraphNode) blockNode()       {}
func (*FencedCodeBlockNode) blockNode() {}
func (*BlockquoteNode) blockNode()      {}
func (*ListNode) blockNode()            {}
func (*ListItemNode) blockNode()        {}

func (*TextNode) inlineNode() {}
