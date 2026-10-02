package ast

// RemapSpan maps a chunk-coordinate span to the parent space, given the
// parent position of each chunk line's first byte. Byte columns make this
// exact even for lines stripped of markers. Remapping composes: mapping into
// an intermediate space and then again equals one mapping into the final one.
func RemapSpan(s Span, bases []Pos) Span {
	return Span{Start: remapPos(s.Start, bases), End: remapPos(s.End, bases)}
}

func remapPos(p Pos, bases []Pos) Pos {
	if len(bases) == 0 {
		return p
	}
	i := p.Line - 1
	if i < 0 {
		i = 0
	}
	if i >= len(bases) {
		i = len(bases) - 1
	}
	b := bases[i]
	return Pos{
		Offset: b.Offset + p.Column - 1,
		Line:   b.Line,
		Column: b.Column + p.Column - 1,
	}
}

// RemapBlocks remaps every block and inline span in the tree.
func RemapBlocks(blocks []BlockNode, bases []Pos) {
	for _, block := range blocks {
		switch n := block.(type) {
		case *HeadingNode:
			n.Span = RemapSpan(n.Span, bases)
			RemapInlines(n.Inlines, bases)
		case *ParagraphNode:
			n.Span = RemapSpan(n.Span, bases)
			RemapInlines(n.Inlines, bases)
		case *FencedCodeBlockNode:
			n.Span = RemapSpan(n.Span, bases)
		case *BlockquoteNode:
			n.Span = RemapSpan(n.Span, bases)
			RemapBlocks(n.Blocks, bases)
		case *ListNode:
			n.Span = RemapSpan(n.Span, bases)
			for _, item := range n.Items {
				item.Span = RemapSpan(item.Span, bases)
				RemapBlocks(item.Blocks, bases)
			}
		case *ListItemNode:
			n.Span = RemapSpan(n.Span, bases)
			RemapBlocks(n.Blocks, bases)
		}
	}
}

// RemapInlines remaps inline spans, recursively.
func RemapInlines(inlines []InlineNode, bases []Pos) {
	for _, in := range inlines {
		switch n := in.(type) {
		case *TextNode:
			n.Span = RemapSpan(n.Span, bases)
		case *EmphasisNode:
			n.Span = RemapSpan(n.Span, bases)
			RemapInlines(n.Children, bases)
		case *StrongNode:
			n.Span = RemapSpan(n.Span, bases)
			RemapInlines(n.Children, bases)
		case *CodeNode:
			n.Span = RemapSpan(n.Span, bases)
		case *LinkNode:
			n.Span = RemapSpan(n.Span, bases)
			RemapInlines(n.Children, bases)
		}
	}
}
