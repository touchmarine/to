// Package node provides a type used to represent node trees.
package node

import (
	"fmt"
	"io"
	"strings"
)

// Node represents an abstract unit of Touch formatted text: text, an element,
// or a container of other nodes.
type Node struct {
	Element string // element name (blank if not an element)
	Type    Type   // node type
	Data    Data   // extra data (e.g. rank)

	Value string // text node's content

	// relationships
	Parent          *Node
	FirstChild      *Node
	LastChild       *Node
	PreviousSibling *Node
	NextSibling     *Node

	Start, End int
	Location   Location // location in source
}

// Data holds any extra data associated with a node.
type Data map[string]interface{}

// Location represents a location inside a resource, such as a line inside a
// text file.
type Location struct {
	URI   DocumentURI
	Range Range
}

// DocumentURI is an URI of a document.
//
// 	  foo://example.com:8042/over/there?name=ferret#nose
// 	  \_/   \______________/\_________/ \_________/ \__/
// 	   |           |            |            |        |
// 	scheme     authority       path        query   fragment
// 	   |   _____________________|__
// 	  / \ /                        \
// 	  urn:example:animal:ferret:nose
//
// https://microsoft.github.io/language-server-protocol/specifications/specification-3-17/#uri
// https://datatracker.ietf.org/doc/html/rfc3986
type DocumentURI string

// Range is like a selection in an editor (zero-based).
type Range struct {
	Start, End Position
}

// Position is like an 'insert' cursor in an editor (zero-based).
type Position struct {
	Offset int // zero-based
	Line   int // zero-based
	Column int // zero-based, byte-count
}

// String returns the node's type and element value. It is used for debugging
// and can change at any time.
func (n Node) String() string {
	return fmt.Sprintf("%s(%s)", n.Type.String(), n.Element)
}

// IsBlock reports whether the node type is a member of the block type set.
func (n Node) IsBlock() bool {
	return IsBlock(n.Type)
}

// IsInline reports whether the node type is a member of the inline type set.
func (n Node) IsInline() bool {
	return IsInline(n.Type)
}

// InsertBefore inserts the newChild immediately before the oldChild.
// newChild is inserted after existing children if oldChild is nil.
//
// https://pkg.go.dev/golang.org/x/net/html#Node.InsertBefore
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

// AppendChild appends the given node (at the end).
//
// It will panic if the given node already has a parent or siblings.
//
// https://pkg.go.dev/golang.org/x/net/html#Node.AppendChild
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

// RemoveChild removes the given child. Afterwards, the removed child will have
// no parent or siblings.
//
// https://pkg.go.dev/golang.org/x/net/html#Node.RemoveChild
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
		panic(fmt.Sprintf("node: node has both value and children (%s)", n))
	}

	if n.Value != "" {
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
