// Package ast defines the markdown AST node types produced by the parser.
package ast

// Pos is a source position: 0-based byte Offset, 1-based Line, 1-based byte
// Column (in the source after "\r\n" → "\n" normalization).
type Pos struct {
	Offset int
	Line   int
	Column int
}

// Span is a half-open source range: Start included, End excluded.
type Span struct {
	Start Pos
	End   Pos
}

// BlockNode is the sealed interface for block-level nodes.
type BlockNode interface {
	blockNode()
	SourceSpan() Span
}

// InlineNode is the sealed interface for inline nodes.
type InlineNode interface {
	inlineNode()
	SourceSpan() Span
}

// DocumentNode is the root of a parsed markdown document.
type DocumentNode struct {
	Blocks []BlockNode
	Span   Span // union of the block spans; zero when empty
}

// HeadingNode represents an ATX heading (`# ...` through `###### ...`).
type HeadingNode struct {
	Level   int // 1..6
	Inlines []InlineNode
	Span    Span // the whole heading line
}

// ParagraphNode represents a paragraph built from consecutive text lines.
type ParagraphNode struct {
	Inlines []InlineNode
	Span    Span // from the first line start to the last line end
}

// FencedCodeBlockNode represents a backtick-fenced code block.
type FencedCodeBlockNode struct {
	Info  string   // raw info string ("js", "js {hl_lines=[1]}"); may be empty
	Lines []string // raw code lines, excluding the fence lines
	Span  Span     // from the opening fence to the closing fence (or EOF)
}

// BlockquoteNode represents a block quote (container).
type BlockquoteNode struct {
	Blocks []BlockNode
	Span   Span
}

// ListNode represents a list built from consecutive list items.
type ListNode struct {
	Items []*ListItemNode
	Span  Span
}

// ListItemNode represents a single list item (container).
type ListItemNode struct {
	Blocks []BlockNode
	Span   Span // from the list marker to the end of the item content
}

// EmphasisNode wraps inline children in emphasis (`*x*` / `_x_`).
type EmphasisNode struct {
	Children []InlineNode
	Span     Span // from the opening delimiter to the closing delimiter
}

// StrongNode wraps inline children in strong emphasis (`**x**` / `__x__`).
type StrongNode struct {
	Children []InlineNode
	Span     Span
}

// CodeNode is a code span; Value is literal text, never further parsed.
type CodeNode struct {
	Value string
	Span  Span // including the backtick runs on both sides
}

// LinkNode is an inline link `[text](dest "title")`. Destination is always
// URL-sanitized (unsafe links never become a LinkNode).
type LinkNode struct {
	Destination string       // raw destination as written
	Title       string       // optional; "" when absent
	Children    []InlineNode // label, recursively parsed; nested links suppressed
	Span        Span         // from "[" to just past the closing ")"
}

// TextNode carries raw per-line text before inline.Process, then literal
// runs in which "\n" acts as a soft line break.
type TextNode struct {
	Value string
	Span  Span
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

func (n *DocumentNode) SourceSpan() Span        { return n.Span }
func (n *HeadingNode) SourceSpan() Span         { return n.Span }
func (n *ParagraphNode) SourceSpan() Span       { return n.Span }
func (n *FencedCodeBlockNode) SourceSpan() Span { return n.Span }
func (n *BlockquoteNode) SourceSpan() Span      { return n.Span }
func (n *ListNode) SourceSpan() Span            { return n.Span }
func (n *ListItemNode) SourceSpan() Span        { return n.Span }
func (n *TextNode) SourceSpan() Span            { return n.Span }
func (n *EmphasisNode) SourceSpan() Span        { return n.Span }
func (n *StrongNode) SourceSpan() Span          { return n.Span }
func (n *CodeNode) SourceSpan() Span            { return n.Span }
func (n *LinkNode) SourceSpan() Span            { return n.Span }
