package node

import (
	"bytes"
	"fmt"
	"strings"
)

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
	TypeHanging
	TypeFenced

	// inlines
	TypeText
	TypeUniform
	TypeEscaped
	TypeForward
)

func (t *Type) UnmarshalText(text []byte) error {
	switch strings.ToLower(string(text)) {
	case "line":
		*t = TypeLine
	case "walled":
		*t = TypeWalled
	case "hanging":
		*t = TypeHanging
	case "fenced":
		*t = TypeFenced
	case "text":
		*t = TypeText
	case "uniform":
		*t = TypeUniform
	case "escaped":
		*t = TypeEscaped
	case "forward":
		*t = TypeForward
	default:
		return fmt.Errorf("unexpected node.Type value: %s", text)
	}
	return nil
}

// TypeCategory is used by parser to determine node category based on type.
func TypeCategory(typ Type) Category {
	if typ > 3 {
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

type ContentInlineChildren interface {
	Content
	InlineChildren
}

type Content interface {
	Content() []byte
}

type HeadBody interface {
	Head() []byte
	Body() []byte
}

type BlockChildren interface {
	BlockChildren() []Block
}

type SettableBlockChildren interface {
	BlockChildren
	SetBlockChildren(children []Block)
}

type InlineChildren interface {
	InlineChildren() []Inline
}

type Ranked interface {
	Rank() uint
}

// NodesToBlocks converts nodes to blocks.
func NodesToBlocks(nodes []Node) []Block {
	blocks := make([]Block, len(nodes))
	for i, n := range nodes {
		block, ok := n.(Block)
		if !ok {
			panic(fmt.Sprintf("node: node %s does not implement node.Block", n.Node()))
		}
		blocks[i] = block
	}
	return blocks
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

type Hanging struct {
	Name     string
	Rank0    uint
	Children []Block
}

func (h Hanging) Node() string {
	return h.Name
}

func (h Hanging) Block() {}

func (h *Hanging) Rank() uint {
	return h.Rank0
}

func (h *Hanging) BlockChildren() []Block {
	return h.Children
}

func (h *Hanging) SetBlockChildren(children []Block) {
	h.Children = children
}

type Fenced struct {
	Name  string
	Lines [][]byte
}

func (f Fenced) Node() string {
	return f.Name
}

func (f Fenced) Block() {}

func (f Fenced) Head() []byte {
	if len(f.Lines) == 0 {
		return nil
	}
	return f.Lines[0]
}

func (f Fenced) Body() []byte {
	if len(f.Lines) == 0 {
		return nil
	}
	return bytes.Join(f.Lines[1:], []byte("\n"))
}

type Uniform struct {
	Name     string
	Children []Inline
}

func (u Uniform) Node() string {
	return u.Name
}

func (u Uniform) Inline() {}

func (u *Uniform) InlineChildren() []Inline {
	return u.Children
}

type Escaped struct {
	Name     string
	Content0 []byte
}

func (e Escaped) Node() string {
	return e.Name
}

func (e Escaped) Inline() {}

func (e *Escaped) Content() []byte {
	return e.Content0
}

type Forward struct {
	Name      string
	Content0  []byte
	Children0 []Inline
}

func (f Forward) Node() string {
	return f.Name
}

func (f Forward) Inline() {}

func (f *Forward) Content() []byte {
	return f.Content0
}

func (f *Forward) InlineChildren() []Inline {
	return f.Children0
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

// LineComment represents text—an atomic, inline node.
type LineComment []byte

// Node returns the node's name.
func (c LineComment) Node() string {
	return "LineComment"
}

func (c LineComment) Inline() {}

// Content returns the LineComment's text.
func (c LineComment) Content() []byte {
	return c
}

type Group struct {
	Name     string
	Children []Block
}

func (g Group) Node() string {
	return g.Name
}

func (g Group) Block() {}

func (g *Group) BlockChildren() []Block {
	return g.Children
}
