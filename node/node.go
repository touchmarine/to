package node

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

const tab = ".   "

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
	return String("Document", map[string]interface{}{
		"Children": d.Children,
	}, 1)
}

type Paragraph struct {
	Children []Inline
}

func (p Paragraph) node()  {}
func (p Paragraph) block() {}
func (p *Paragraph) String(indent int) string {
	return String("Paragraph", map[string]interface{}{
		"Children": InlinesToNodes(p.Children),
	}, indent)
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
	return String("Emphasis", map[string]interface{}{
		"Children": InlinesToNodes(e.Children),
	}, indent)
}

type Strong struct {
	Children []Inline
}

func (s Strong) node()   {}
func (s Strong) inline() {}
func (s *Strong) String(indent int) string {
	return String("Strong", map[string]interface{}{
		"Children": InlinesToNodes(s.Children),
	}, indent)
}

type Heading struct {
	Level      int
	IsNumbered bool
	Children   []Inline
}

func (h Heading) node()  {}
func (h Heading) block() {}
func (h *Heading) String(indent int) string {
	return String("Heading", map[string]interface{}{
		"Level":      strconv.Itoa(h.Level),
		"IsNumbered": strconv.FormatBool(h.IsNumbered),
		"Children":   InlinesToNodes(h.Children),
	}, indent)
}

type Link struct {
	Destination string
	Children    []Inline
}

func (l Link) node()   {}
func (l Link) inline() {}
func (l *Link) String(indent int) string {
	return String("Link", map[string]interface{}{
		"Destination": l.Destination,
		"Children":    InlinesToNodes(l.Children),
	}, indent)
}

// String representation of an element.
func String(name string, fields map[string]interface{}, indent int) string {
	var b strings.Builder
	b.WriteString(name + "{")

	// sort map keys
	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// print fields (key-value pairs) in separate lines, indented
	for _, k := range keys {
		i := fields[k]
		b.WriteString(fmt.Sprintf("\n%s%s: ", strings.Repeat(tab, indent), k))

		switch v := i.(type) {
		case string:
			b.WriteString(v)
		case []Node:
			b.WriteString(NodesString(v, indent+1))
		}

		b.WriteString(",")
	}

	// move into next line and indent closing '}' if fields are present
	if len(fields) > 0 {
		b.WriteString("\n" + strings.Repeat(tab, indent-1))
	}

	b.WriteString("}")

	return b.String()
}

// NodesString returns a string representation of nodes.
func NodesString(nodes []Node, indent int) string {
	if len(nodes) == 0 {
		return "[]"
	}

	var b strings.Builder
	b.WriteString("[\n")

	for _, node := range nodes {
		b.WriteString(strings.Repeat(tab, indent) + node.String(indent+1) + ",\n")
	}

	b.WriteString(strings.Repeat(tab, indent-1) + "]")

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
