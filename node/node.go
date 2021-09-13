package node

import (
	"fmt"
	"strconv"
	"strings"
)

//go:generate stringer -type=Category
type Category int

// Categories of nodes
const (
	CategoryError Category = iota
	CategoryBlock
	CategoryInline
)

//go:generate stringer -type=Type
type Type int

// Types of nodes
const (
	TypeError Type = iota

	// blocks
	TypeContainer
	TypeLeaf
	TypeVerbatimLine
	TypeWalled
	TypeVerbatimWalled
	TypeHanging
	TypeRankedHanging
	TypeFenced

	// inlines
	TypeText
	TypeUniform
	TypeEscaped
	TypePrefixed
)

func (t *Type) UnmarshalText(text []byte) error {
	switch s := strings.ToLower(string(text)); s {
	case "leaf":
		*t = TypeLeaf
	case "verbatimline":
		*t = TypeVerbatimLine
	case "walled":
		*t = TypeWalled
	case "verbatimwalled":
		*t = TypeVerbatimWalled
	case "hanging":
		*t = TypeHanging
	case "rankedhanging":
		*t = TypeRankedHanging
	case "fenced":
		*t = TypeFenced
	case "text":
		*t = TypeText
	case "uniform":
		*t = TypeUniform
	case "escaped":
		*t = TypeEscaped
	case "prefixed":
		*t = TypePrefixed
	default:
		return fmt.Errorf("unexpected node.Type value: %q", s)
	}
	return nil
}

// TypeCategory is used by parser to determine node category based on type.
func TypeCategory(typ Type) Category {
	if typ == 0 {
		return CategoryError
	} else if typ <= 8 {
		return CategoryBlock
	} else {
		return CategoryInline
	}
}

type Node struct {
	Name string
	Type Type
	Data string

	Parent      *Node
	FirstChild  *Node
	LastChild   *Node
	PrevSibling *Node
	NextSibling *Node
}

func (n Node) TypeCategory() Category {
	return TypeCategory(n.Type)
}

func (n *Node) AppendChild(c *Node) {
	if c.Parent != nil || c.PrevSibling != nil || c.NextSibling != nil {
		panic("node: child already attached")
	}

	last := n.LastChild
	if last != nil {
		last.NextSibling = c
	} else {
		n.FirstChild = c
	}

	n.LastChild = c
	c.Parent = n
	c.PrevSibling = last
}

type Block interface {
	Block()
}

type Inline interface {
	Inline()
}

type Content interface {
	Content() []byte
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
	Rank() int
}

type Boxed interface {
	Unbox() Node
}

// NodesToBlocks converts nodes to blocks.
func NodesToBlocks(nodes []Node) []Block {
	return nil
}

// NodesToInlines converts nodes to blocks.
func NodesToInlines(nodes []Node) []Inline {
	return nil
}

// BlocksToNodes converts blocks to nodes.
func BlocksToNodes(blocks []Block) []Node {
	return nil
}

// InlinesToNodes converts inlines to nodes.
func InlinesToNodes(inlines []Inline) []Node {
	return nil
}

type Leaf struct {
	Name     string
	Children []Inline
}

func (l Leaf) Node() string {
	return l.Name
}

func (l Leaf) Block() {}

func (l *Leaf) InlineChildren() []Inline {
	return l.Children
}

func (l *Leaf) SetInlineChildren(children []Inline) {
	l.Children = children
}

type VerbatimLine struct {
	Name     string
	Content0 []byte
}

func (l VerbatimLine) Node() string {
	return l.Name
}

func (l VerbatimLine) Block() {}

func (l *VerbatimLine) Content() []byte {
	return l.Content0
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

type VerbatimWalled struct {
	Name   string
	Lines0 [][]byte
}

func (b VerbatimWalled) Node() string {
	return b.Name
}

func (b VerbatimWalled) Block() {}

func (b *VerbatimWalled) Lines() [][]byte {
	return b.Lines0
}

type Hanging struct {
	Name     string
	Children []Block
}

func (h Hanging) Node() string {
	return h.Name
}

func (h Hanging) Block() {}

func (h *Hanging) BlockChildren() []Block {
	return h.Children
}

func (h *Hanging) SetBlockChildren(children []Block) {
	h.Children = children
}

type RankedHanging struct {
	Name     string
	Rank0    int
	Children []Block
}

func (h RankedHanging) Node() string {
	return h.Name
}

func (h RankedHanging) Block() {}

func (h *RankedHanging) Rank() int {
	return h.Rank0
}

func (h *RankedHanging) BlockChildren() []Block {
	return h.Children
}

func (h *RankedHanging) SetBlockChildren(children []Block) {
	h.Children = children
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

type Prefixed struct {
	Name     string
	Content0 []byte
}

func (p Prefixed) Node() string {
	return p.Name
}

func (Prefixed) Inline() {}

func (p *Prefixed) Content() []byte {
	return p.Content0
}

// Text represents text—an atomic, inline node.
type Text struct {
	Name     string
	Content0 []byte
}

// Node returns the node's name.
func (t Text) Node() string {
	return t.Name
}

func (t Text) Inline() {}

// Content returns the text.
func (t *Text) Content() []byte {
	return t.Content0
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

type Sticky struct {
	Name     string
	After    bool
	Children []Block
}

func (s Sticky) Node() string {
	return s.Name
}

func (s Sticky) Block() {}

func (s *Sticky) Sticky() Block {
	if l := len(s.Children); l != 2 {
		panic(fmt.Sprintf("node: unexpected number of children: %d", l))
	}

	if s.After {
		return s.Children[1]
	} else {
		return s.Children[0]
	}
}

func (s *Sticky) Target() Block {
	if l := len(s.Children); l != 2 {
		panic(fmt.Sprintf("node: unexpected number of children: %d", l))
	}

	if s.After {
		return s.Children[0]
	} else {
		return s.Children[1]
	}
}

func (s *Sticky) BlockChildren() []Block {
	return s.Children
}

func (s *Sticky) SetBlockChildren(children []Block) {
	s.Children = children
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

type SequentialNumberBox struct {
	Nod               Node
	SequentialNumbers []int
}

func (s SequentialNumberBox) Node() string {
	return fmt.Sprintf("SequentialNumberBox(%s%s)", s.Nod.Name, s.SequentialNumber())
}

func (s SequentialNumberBox) Block() {}

func (s *SequentialNumberBox) Unbox() Node {
	return s.Nod
}

func (s *SequentialNumberBox) SequentialNumber() string {
	var a []string
	for _, n := range s.SequentialNumbers {
		a = append(a, strconv.FormatUint(uint64(n), 10))
	}
	return strings.Join(a, ".")
}

func ExtractText(n Node) string {
	return ExtractTextWithReplacements(n, nil)
}

// ExtractTextWithReplacments is like ExtractText but replaces the node text
// with the replacement value.
//
// replacementMap = map[node.Name]text
func ExtractTextWithReplacements(n Node, replacementMap map[string]string) string {
	var b strings.Builder

	text, found := replacementMap[n.Name]
	if found {
		// found replacement
		b.WriteString(text)
		return b.String()
	}

	//if m, ok := n.(Boxed); ok {
	//	unboxed := m.Unbox()
	//	if unboxed == nil {
	//		return ""
	//	}
	//	return ExtractTextWithReplacements(unboxed, replacementMap)
	//}

	//if m, ok := n.(Composited); ok {
	//	return ExtractTextWithReplacements(m.Primary(), replacementMap)
	//}

	//if m, ok := n.(BlockChildren); ok {
	//	for i, c := range m.BlockChildren() {
	//		if i > 0 {
	//			b.WriteString("\n")
	//		}

	//		b.WriteString(ExtractTextWithReplacements(c, replacementMap))
	//	}
	//}

	//if m, ok := n.(InlineChildren); ok {
	//	for _, c := range m.InlineChildren() {
	//		b.WriteString(ExtractTextWithReplacements(c, replacementMap))
	//	}
	//} else if m, ok := n.(Content); ok {
	//	b.Write(m.Content())
	//}

	//if m, ok := n.(Lines); ok {
	//	for _, line := range m.Lines() {
	//		b.Write(line)
	//	}
	//}

	return b.String()

	//var b strings.Builder

	//switch n.(type) {
	//case BlockChildren, InlineChildren, Content, Lines, Composited, Boxed:
	//default:
	//	panic(fmt.Sprintf("ExtractTextWithReplacements: unexpected node type %T", n))
	//}

	//for name, text := range replacementMap {
	//	if name == n.Node() {
	//		// found replacement

	//		b.WriteString(text)
	//		return b.String()
	//	}
	//}

	//if m, ok := n.(Boxed); ok {
	//	unboxed := m.Unbox()
	//	if unboxed == nil {
	//		return ""
	//	}
	//	return ExtractTextWithReplacements(unboxed, replacementMap)
	//}

	//if m, ok := n.(Composited); ok {
	//	return ExtractTextWithReplacements(m.Primary(), replacementMap)
	//}

	//if m, ok := n.(BlockChildren); ok {
	//	for i, c := range m.BlockChildren() {
	//		if i > 0 {
	//			b.WriteString("\n")
	//		}

	//		b.WriteString(ExtractTextWithReplacements(c, replacementMap))
	//	}
	//}

	//if m, ok := n.(InlineChildren); ok {
	//	for _, c := range m.InlineChildren() {
	//		b.WriteString(ExtractTextWithReplacements(c, replacementMap))
	//	}
	//} else if m, ok := n.(Content); ok {
	//	b.Write(m.Content())
	//}

	//if m, ok := n.(Lines); ok {
	//	for _, line := range m.Lines() {
	//		b.Write(line)
	//	}
	//}

	//return b.String()
}
