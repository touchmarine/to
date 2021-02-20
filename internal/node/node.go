package node

//go:generate stringer -type=Category
type Category uint

// Categories of nodes
const (
	CategoryBlock Category = iota
	CategoryInline
)

//go:generate stringer -type=Type
type Type uint

// Types of nodes
const (
	// blocks
	TypeLine Type = iota
	TypeWalled

	// inlines
	TypeText
)

// TypeCategory is used by parser to determine node category based on type.
func TypeCategory(typ Type) Category {
	if typ > 1 {
		return CategoryInline
	}
	return CategoryBlock
}

// Node represents an element.
type Node interface {
	Node() string
}

type Block interface {
	Node
	Block()
}

type Inline interface {
	Node
	Inline()
}

type Content interface {
	Content() []byte
}

type BlockChildren interface {
	BlockChildren() []Block
}

type InlineChildren interface {
	InlineChildren() []Inline
}

// BlocksToNodes converts blocks to nodes.
func BlocksToNodes(blocks []Block) []Node {
	nodes := make([]Node, len(blocks))
	for i, b := range blocks {
		nodes[i] = Node(b)
	}
	return nodes
}

// InlinesToNodes converts inlines to nodes.
func InlinesToNodes(inlines []Inline) []Node {
	nodes := make([]Node, len(inlines))
	for i, v := range inlines {
		nodes[i] = Node(v)
	}
	return nodes
}

type Line struct {
	Name     string
	Children []Inline
}

func (l Line) Node() string {
	return l.Name
}

func (l Line) Block() {}

func (l *Line) InlineChildren() []Inline {
	return l.Children
}

type Walled struct {
	Name     string
	Children []Block
}

func (w Walled) Node() string {
	return w.Name
}

func (w Walled) Block() {}

func (w *Walled) BlockChildren() []Block {
	return w.Children
}

// Text represents text—an atomic, inline node.
type Text []byte

// Node returns the node's name.
func (t Text) Node() string {
	return "Text"
}

func (t Text) Inline() {}

// Content returns the text.
func (t Text) Content() []byte {
	return t
}