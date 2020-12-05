package node

import (
	"fmt"
	"strings"
)

type Node interface {
	node()                    // dummy method to conform to interface
	String(indent int) string // string representation for testing and debugging
}

type Block interface {
	Node
	block() // dummy method
}

type Inline interface {
	Node
	inline() // dummy method
}

type Document struct {
	Children []Node
}

func (d *Document) String() string {
	return String(d.Children, "Document", 1)
}

type Paragraph struct {
	Children []Inline
}

func (p Paragraph) node()  {}
func (p Paragraph) block() {}
func (p *Paragraph) String(indent int) string {
	return String(InlinesToNodes(p.Children), "Paragraph", indent)
}

type Text struct {
	Value string
}

func (t Text) node()   {}
func (t Text) inline() {}
func (t *Text) String(indent int) string {
	return fmt.Sprintf("Text(%s)", t.Value)
}

type Emphasis struct {
	Children []Inline
}

func (e Emphasis) node()   {}
func (e Emphasis) inline() {}
func (e *Emphasis) String(indent int) string {
	return String(InlinesToNodes(e.Children), "Emphasis", indent)
}

type Strong struct {
	Children []Inline
}

func (s Strong) node()   {}
func (s Strong) inline() {}
func (s *Strong) String(indent int) string {
	return String(InlinesToNodes(s.Children), "Strong", indent)
}

// String representation of nodes.
func String(nodes []Node, name string, indent int) string {
	if len(nodes) == 0 {
		return name + "([])"
	}

	var b strings.Builder
	b.WriteString(name + "([\n")

	const tab = ".   "
	for _, node := range nodes {
		b.WriteString(strings.Repeat(tab, indent) + node.String(indent+1) + "\n")
	}

	b.WriteString(strings.Repeat(tab, indent-1) + "])")

	return b.String()
}

// InlinesToNodes converts []Inline to []Node.
func InlinesToNodes(inlines []Inline) []Node {
	nodes := make([]Node, len(inlines))
	for i, v := range inlines {
		nodes[i] = Node(v)
	}
	return nodes
}
