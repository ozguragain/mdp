package ast

type Node interface {
	Type() string
}

type BlockNode interface {
	Node
	blocknode() // dummy method to distinguish BlockNode from other Node types
}

// root markdown document node (whole markdown document)
type DocumentNode struct {
	Blocks []BlockNode
}

func (d *DocumentNode) Type() string {
	return "Document"
}
