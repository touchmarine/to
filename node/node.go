package node

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

//go:generate stringer -type=Type
type Type int

// Types of nodes
const (
	// special
	TypeError Type = iota
	TypeContainer

	// blocks
	TypeWalled
	TypeVerbatimWalled
	TypeHanging
	TypeRankedHanging
	TypeFenced
	TypeVerbatimLine
	TypeLeaf

	// inlines
	TypeUniform
	TypeEscaped
	TypePrefixed
	TypeText
)

func (t *Type) UnmarshalText(text []byte) error {
	switch s := strings.ToLower(string(text)); s {
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
	case "verbatimline":
		*t = TypeVerbatimLine
	case "leaf":
		*t = TypeLeaf
	case "uniform":
		*t = TypeUniform
	case "escaped":
		*t = TypeEscaped
	case "prefixed":
		*t = TypePrefixed
	case "text":
		*t = TypeText
	default:
		return fmt.Errorf("unexpected node.Type value: %q", s)
	}
	return nil
}

func IsBlock(t Type) bool {
	return t >= TypeWalled && t <= TypeLeaf
}

func IsInline(t Type) bool {
	return t >= TypeUniform
}

func HasDelimiter(t Type) bool {
	return t == TypeWalled || t == TypeVerbatimWalled || t == TypeHanging ||
		t == TypeRankedHanging || t == TypeFenced || t == TypeVerbatimLine ||
		t == TypeUniform || t == TypeEscaped || t == TypePrefixed
}

type Node struct {
	Element string // element name
	Type    Type
	Data    Data // additional data, like rank

	Value string

	Parent      *Node
	FirstChild  *Node
	LastChild   *Node
	PrevSibling *Node
	NextSibling *Node
}

type Data map[string]interface{}

// String is used for debugging and can change at any time.
func (n Node) String() string {
	return fmt.Sprintf("%s(%s)", n.Type.String()[len("Type"):], n.Element)
}

func (n Node) IsBlock() bool {
	return IsBlock(n.Type)
}

func (n Node) IsInline() bool {
	return IsInline(n.Type)
}

// IsElementBlock is like IsBlock but also determines if the node is an element
// not just a plain container.
//func (n Node) IsElementBlock() bool {
//	return n.IsBlock() && n.Element != ""
//}

// IsPlainContainer reports whether the node is a plain container.
//
// A plain container is just a convenient wrapper for multiple nodes and does
// not represent any element or hold data.
//func (n Node) IsPlainContainer() bool {
//	return n.Element == "" && (n.Type == TypeContainer || n.Type == TypeInlineContainer)
//}

// IsGroupContainer reports whether the node is a group container.
//
// A group container is a node wrapper that represents an element and holds
// data. It
//func (n Node) IsGroupContainer() bool {
//	return n.Element != "" && (n.Type == TypeContainer || n.Type == TypeInlineContainer)
//}

func (n *Node) InsertBefore(newChild, oldChild *Node) {
	if newChild.Parent != nil || newChild.PrevSibling != nil || newChild.NextSibling != nil {
		panic("node: InsertBefore called for an attached child Node")
	}
	var prev, next *Node
	if oldChild != nil {
		prev, next = oldChild.PrevSibling, oldChild
	} else {
		prev = n.LastChild
	}
	if prev != nil {
		prev.NextSibling = newChild
	} else {
		n.FirstChild = newChild
	}
	if next != nil {
		next.PrevSibling = newChild
	} else {
		n.LastChild = newChild
	}
	newChild.Parent = n
	newChild.PrevSibling = prev
	newChild.NextSibling = next
}

func (n *Node) AppendChild(c *Node) {
	if c.Parent != nil || c.PrevSibling != nil || c.NextSibling != nil {
		panic("node: AppendChild called for an attached child Node")
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

func (n *Node) RemoveChild(c *Node) {
	if c.Parent != n {
		panic("node: RemoveChild called for a non-child Node")
	}
	if n.FirstChild == c {
		n.FirstChild = c.NextSibling
	}
	if c.NextSibling != nil {
		c.NextSibling.PrevSibling = c.PrevSibling
	}
	if n.LastChild == c {
		n.LastChild = c.PrevSibling
	}
	if c.PrevSibling != nil {
		c.PrevSibling.NextSibling = c.NextSibling
	}
	c.Parent = nil
	c.PrevSibling = nil
	c.NextSibling = nil
}

// TextContent returns the text content of the node and its descendants.
func (n Node) TextContent() string {
	var b strings.Builder
	n.textContent(&b)
	return b.String()
}

func (n Node) textContent(w io.StringWriter) {
	if n.Value != "" && n.FirstChild != nil {
		panic(fmt.Sprintf("node: node has both data and children (%s)", n))
	} else if n.Value != "" {
		lines := strings.Split(n.Value, "\n")
		isFilled := false
		for _, line := range lines {
			if strings.Trim(line, " \t") != "" {
				isFilled = true
				break
			}
		}

		if isFilled {
			w.WriteString(n.Value)
		}
	} else if n.FirstChild != nil {
		i := 0
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if i > 0 && c.IsBlock() {
				w.WriteString("\n")
			}

			c.textContent(w)
			i++
		}
	}
}

/*
func (n *Node) Clone() *Node {
	return &Node{
		Element: n.Element,
		Type:    n.Type,
		Data:    n.Data,

		Value: n.Value,
	}
}

func (n *Node) FullClone() *Node {
	m := &Node{
		Element: n.Element,
		Type: n.Type,
		Data: n.Data,

		Value: n.Value,
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		m.AppendChild(m.FullClone())
	}
	return m
}

func ReparentChildren(dst, src *Node) {
	for {
		c := src.FirstChild
		if c == nil {
			break
		}
		src.RemoveChild(c)
		dst.AppendChild(c)
	}
}
*/

/*
type Node interface {
	Element() string // name of element
	Type() Type
	Data() interface{}
	Position() Position
	Attributes() Attributes
}

type Position struct {
	Filename string
	Start    int
	End      int
}

type Point struct {
	Offset int
	Line   int
	Column int
}

type Attributes map[string][]string

type Parent interface {
	Parent() *Node
	FirstChild() *Node
	LastChild() *Node
	PreviousSibling() *Node
	NextSibling() *Node
}

type ManipulatableParent interface {
	Parent

	InsertBefore(*Node, *Node)
	AppendChild(*Node)
	RemoveChild()
}

type Literal interface {
	Node
	Value() string
}
*/

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
	//return fmt.Sprintf("SequentialNumberBox(%s%s)", s.Nod.Name, s.SequentialNumber())
	return ""
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
