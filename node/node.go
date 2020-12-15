package node

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

const tab = ".   "

type ListType string

// list types
const (
	UnorderedList ListType = "unordered"
	NumberedList           = "numbers"
	//LowercaseLetters                = "lowercaseLetters"
	//UppercaseLetters                = "uppercaseLetters"
	//LowercaseRomanNumerals          = "lowercaseRomanNumerals"
	//UppercaseRomanNumerals          = "uppercaseRomanNumerals"
)

type Node interface {
	node()                    // dummy method to conform to interface
	Pretty(indent int) string // prettified string representation; usually for testing and debugging
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
	return Pretty("Document", map[string]interface{}{
		"Children": d.Children,
	}, 1)
}

type Paragraph struct {
	Children []Inline
}

func (p Paragraph) node()  {}
func (p Paragraph) block() {}
func (p *Paragraph) Pretty(indent int) string {
	return Pretty("Paragraph", map[string]interface{}{
		"Children": InlinesToNodes(p.Children),
	}, indent)
}

type Text struct {
	Value string
}

func (t Text) node()   {}
func (t Text) inline() {}
func (t *Text) Pretty(indent int) string {
	return fmt.Sprintf("Text(%s)", t.Value)
}

type Emphasis struct {
	Children []Inline
}

func (e Emphasis) node()   {}
func (e Emphasis) inline() {}
func (e *Emphasis) Pretty(indent int) string {
	return Pretty("Emphasis", map[string]interface{}{
		"Children": InlinesToNodes(e.Children),
	}, indent)
}

type Strong struct {
	Children []Inline
}

func (s Strong) node()   {}
func (s Strong) inline() {}
func (s *Strong) Pretty(indent int) string {
	return Pretty("Strong", map[string]interface{}{
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
func (h *Heading) Pretty(indent int) string {
	return Pretty("Heading", map[string]interface{}{
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
func (l *Link) Pretty(indent int) string {
	return Pretty("Link", map[string]interface{}{
		"Destination": l.Destination,
		"Children":    InlinesToNodes(l.Children),
	}, indent)
}

type CodeBlock struct {
	Language    string
	Filename    string
	MetadataRaw string
	Body        string
}

func (cb CodeBlock) node()  {}
func (cb CodeBlock) block() {}
func (cb *CodeBlock) Pretty(indent int) string {
	return Pretty("CodeBlock", map[string]interface{}{
		"Language":    cb.Language,
		"Filename":    cb.Filename,
		"MetadataRaw": cb.MetadataRaw,
		"Body":        cb.Body,
	}, indent)
}

type List struct {
	//IsContinued bool     // whether counting continues onward from the previous list
	Type      ListType // unordered or numbering type if ordered
	ListItems [][]Node
}

func (l List) node()  {}
func (l List) block() {}
func (l *List) Pretty(indent int) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("[%d]", len(l.ListItems)))
	b.WriteString("[\n" + strings.Repeat(tab, indent+1))

	for i, li := range l.ListItems {
		if i > 0 {
			b.WriteString(",\n" + strings.Repeat(tab, indent+1))
		}

		b.WriteString(PrettyNodes(li, indent+2))
	}

	b.WriteString("\n" + strings.Repeat(tab, indent) + "]")

	return Pretty("List", map[string]interface{}{
		"Type":      string(l.Type),
		"ListItems": b.String(),
	}, indent)
}

// Pretty returns a prettified string representation of an element; usually used
// for testing and debugging.
func Pretty(name string, fields map[string]interface{}, indent int) string {
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
			b.WriteString(PrettyNodes(v, indent+1))
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

// PrettyNodes returns a prettified string representation of nodes.
func PrettyNodes(nodes []Node, indent int) string {
	if len(nodes) == 0 {
		return "[]"
	}

	var b strings.Builder
	b.WriteString("[\n")

	for _, node := range nodes {
		b.WriteString(strings.Repeat(tab, indent) + node.Pretty(indent+1) + ",\n")
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

// ListsToNodes converts []*List to []Node.
func ListsToNodes(lists []*List) []Node {
	nodes := make([]Node, len(lists))
	for i, v := range lists {
		nodes[i] = Node(v)
	}
	return nodes
}
