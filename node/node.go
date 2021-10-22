package node

import (
	"fmt"
	"io"
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

	Parent          *Node
	FirstChild      *Node
	LastChild       *Node
	PreviousSibling *Node
	NextSibling     *Node
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

func (n *Node) InsertBefore(newChild, oldChild *Node) {
	if newChild.Parent != nil || newChild.PreviousSibling != nil || newChild.NextSibling != nil {
		panic("node: InsertBefore called for an attached child Node")
	}
	var prev, next *Node
	if oldChild != nil {
		prev, next = oldChild.PreviousSibling, oldChild
	} else {
		prev = n.LastChild
	}
	if prev != nil {
		prev.NextSibling = newChild
	} else {
		n.FirstChild = newChild
	}
	if next != nil {
		next.PreviousSibling = newChild
	} else {
		n.LastChild = newChild
	}
	newChild.Parent = n
	newChild.PreviousSibling = prev
	newChild.NextSibling = next
}

func (n *Node) AppendChild(c *Node) {
	if c.Parent != nil || c.PreviousSibling != nil || c.NextSibling != nil {
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
	c.PreviousSibling = last
}

func (n *Node) RemoveChild(c *Node) {
	if c.Parent != n {
		panic("node: RemoveChild called for a non-child Node")
	}
	if n.FirstChild == c {
		n.FirstChild = c.NextSibling
	}
	if c.NextSibling != nil {
		c.NextSibling.PreviousSibling = c.PreviousSibling
	}
	if n.LastChild == c {
		n.LastChild = c.PreviousSibling
	}
	if c.PreviousSibling != nil {
		c.PreviousSibling.NextSibling = c.NextSibling
	}
	c.Parent = nil
	c.PreviousSibling = nil
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
