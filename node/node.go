package node

import (
	"fmt"
	"strconv"
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

type LinesTrailingText interface {
	Lines
	TrailingText
}

type LinesBoxed interface {
	Lines
	Boxed
}

type Lines interface {
	Lines() [][]byte
}

type SettableLines interface {
	Lines
	SetLines(lines [][]byte)
}

type TrailingText interface {
	TrailingText() []byte
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

type SettableInlineChildren interface {
	InlineChildren
	SetInlineChildren(children []Inline)
}

type Composited interface {
	Primary() Inline
	Secondary() Inline
}

type Ranked interface {
	Rank() uint
}

type Boxed interface {
	Unbox() Node
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

// NodesToInlines converts nodes to blocks.
func NodesToInlines(nodes []Node) []Inline {
	inlines := make([]Inline, len(nodes))
	for i, n := range nodes {
		inline, ok := n.(Inline)
		if !ok {
			panic(fmt.Sprintf("node: node %s does not implement node.Inline", n.Node()))
		}
		inlines[i] = inline
	}
	return inlines
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

func (l *Line) SetInlineChildren(children []Inline) {
	l.Children = children
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

func (w *Walled) SetBlockChildren(children []Block) {
	w.Children = children
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

type HangingVerbatim struct {
	Name   string
	Rank0  uint
	Lines0 [][]byte
}

func (hv HangingVerbatim) Node() string {
	return hv.Name
}

func (hv HangingVerbatim) Block() {}

func (hv *HangingVerbatim) Rank() uint {
	return hv.Rank0
}

func (hv *HangingVerbatim) Lines() [][]byte {
	return hv.Lines0
}

func (hv *HangingVerbatim) SetLines(lines [][]byte) {
	hv.Lines0 = lines
}

type Fenced struct {
	Name          string
	Lines0        [][]byte
	TrailingText0 []byte
}

func (f Fenced) Node() string {
	return f.Name
}

func (f Fenced) Block() {}

func (f Fenced) Lines() [][]byte {
	return f.Lines0
}

func (f Fenced) TrailingText() []byte {
	return f.TrailingText0
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

func (g *Group) SetBlockChildren(children []Block) {
	g.Children = children
}

type Composite struct {
	Name       string
	Primary0   Inline
	Secondary0 Inline
}

func (c Composite) Node() string {
	return c.Name
}

func (c Composite) Inline() {}

func (c *Composite) Primary() Inline {
	return c.Primary0
}

func (c *Composite) Secondary() Inline {
	return c.Secondary0
}

type Hat struct {
	Lines0 [][]byte
	Nod    Node
}

func (h Hat) Node() string {
	if h.Nod == nil {
		return "Hat()"
	}
	return fmt.Sprintf("Hat(%s)", h.Nod.Node())
}

func (h Hat) Block() {}

func (h *Hat) Lines() [][]byte {
	return h.Lines0
}

func (h *Hat) Unbox() Node {
	return h.Nod
}

type SeqNumBox struct {
	Nod     Node
	SeqNums []uint
}

func (s SeqNumBox) Node() string {
	return fmt.Sprintf("SeqNumBox(%s%s)", s.Nod.Node(), s.SeqNum())
}

func (s *SeqNumBox) Unbox() Node {
	return s.Nod
}

func (s *SeqNumBox) SeqNum() string {
	var a []string
	for _, n := range s.SeqNums {
		a = append(a, strconv.FormatUint(uint64(n), 10))
	}
	return strings.Join(a, ".")
}

func ExtractText(n Node) string {
	var b strings.Builder

	switch n.(type) {
	case BlockChildren, InlineChildren, Content, Lines, Composited, Boxed:
	default:
		panic(fmt.Sprintf("ExtractText: unexpected node type %T", n))
	}

	if m, ok := n.(Boxed); ok {
		unboxed := m.Unbox()
		if unboxed == nil {
			return ""
		}
		return ExtractText(unboxed)
	}

	if m, ok := n.(Composited); ok {
		return ExtractText(m.Primary())
	}

	if m, ok := n.(BlockChildren); ok {
		for i, c := range m.BlockChildren() {
			if i > 0 {
				b.WriteString("\n")
			}

			b.WriteString(ExtractText(c))
		}
	}

	if m, ok := n.(ContentInlineChildren); ok {
		for _, c := range m.InlineChildren() {
			b.WriteString(ExtractText(c))
		}
	} else if m, ok := n.(InlineChildren); ok {
		for _, c := range m.InlineChildren() {
			b.WriteString(ExtractText(c))
		}
	} else if m, ok := n.(Content); ok {
		b.Write(m.Content())
	}

	if m, ok := n.(Lines); ok {
		for _, line := range m.Lines() {
			b.Write(line)
		}
	}

	return b.String()
}
