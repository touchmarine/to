package node

//go:generate stringer -type=Type
type Type uint

// Types of nodes
const (
	// blocks
	TypeWalled Type = iota

	// inlines
	TypeText
)

type Node interface {
	Node() (string, Type)
}

type NodeChildren interface {
	Node
	Children() []Node
}

type block struct {
	name     string
	typ      Type
	children []Node
}

func NewBlock(name string, typ Type, children []Node) Node {
	return &block{name, typ, children}
}

func (b block) Node() (string, Type) {
	return b.name, b.typ
}

func (b *block) Children() []Node {
	return b.children
}

type Text []byte

func (t Text) Node() (string, Type) {
	return "Text", TypeText
}

//go:generate stringer -type=Category
type Category uint

// Categories of nodes
const (
	CategoryBlock Category = iota
	CategoryInline
)

func TypeCategory(typ Type) Category {
	if typ > 0 {
		return CategoryInline
	}
	return CategoryBlock
}
