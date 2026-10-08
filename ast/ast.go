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
}

// InlineNode is the sealed interface for inline nodes.
type InlineNode interface {
	inlineNode()
}

// DocumentNode is the root of a parsed markdown document. Frontmatter is
// stripped before parsing and is not part of the tree.
type DocumentNode struct {
	Blocks []BlockNode
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
	Items   []*ListItemNode
	Ordered bool // false for "-"/"*"/"+", true for "1."/"1)"
	Start   int  // first marker's number; 1 for bullet lists
	Span    Span
}

// ThematicBreakNode is a thematic break ("---", "***", "___").
type ThematicBreakNode struct {
	Span Span // the whole break line
}

// Align is a table column's horizontal alignment, taken from the delimiter row.
type Align int

const (
	AlignNone   Align = iota // no ":" in the delimiter cell
	AlignLeft                // ":---"
	AlignCenter              // ":---:"
	AlignRight               // "---:"
)

// TableCellNode is one table cell. It is structural, not a block or inline
// node of its own.
type TableCellNode struct {
	Inlines []InlineNode
	Span    Span
}

// TableNode is a GFM pipe table: one header row, a delimiter row that is
// consumed (never part of the tree), then zero or more body rows. Every row
// is padded or truncated to len(Alignments) columns.
type TableNode struct {
	Header     []TableCellNode
	Alignments []Align
	Rows       [][]TableCellNode
	Span       Span // from the header row to the last body row
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

// ImageNode is an inline image `![alt](src "title")`. Alt is literal text —
// the spec parses no markup inside it. Destination is always URL-sanitized:
// an unsafe source never becomes an ImageNode.
type ImageNode struct {
	Alt         string // literal alt text
	Destination string // raw destination as written
	Title       string // optional; "" when absent
	Span        Span   // from "![" to just past the closing ")"
}

// StrikethroughNode wraps inline children in `~~...~~`.
type StrikethroughNode struct {
	Children []InlineNode
	Span     Span
}

// HardBreakNode is an explicit line break inside a block: two or more
// trailing spaces, or a backslash, immediately before the newline.
type HardBreakNode struct {
	Span Span // covering the newline that produced the break
}

// AutolinkNode is `<https://example.com>` or `<user@example.com>`. The
// destination is URL-sanitized like any other link.
type AutolinkNode struct {
	Destination string // the href; for emails this is "mailto:" + the text
	Text        string // literal text shown to the reader
	Span        Span   // including both angle brackets
}

func (*HeadingNode) blockNode()         {}
func (*ParagraphNode) blockNode()       {}
func (*FencedCodeBlockNode) blockNode() {}
func (*BlockquoteNode) blockNode()      {}
func (*ListNode) blockNode()            {}
func (*ListItemNode) blockNode()        {}
func (*ThematicBreakNode) blockNode()   {}
func (*TableNode) blockNode()           {}

func (*TextNode) inlineNode()          {}
func (*EmphasisNode) inlineNode()      {}
func (*StrongNode) inlineNode()        {}
func (*CodeNode) inlineNode()          {}
func (*LinkNode) inlineNode()          {}
func (*ImageNode) inlineNode()         {}
func (*StrikethroughNode) inlineNode() {}
func (*HardBreakNode) inlineNode()     {}
func (*AutolinkNode) inlineNode()      {}
