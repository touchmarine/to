// Package paragraph provides a transformer for recognizing and adding
// paragraphs to node trees.
package paragraph

import (
	"github.com/touchmarine/to/node"
)

// Map is a map of paragraph names (keys) to node types (values). Node types
// tell the transformer the element types that can be grouped into paragraphs.
type Map map[string]node.Type

// Transformer recognizes paragraphs based on the given Paragraphs and adds them
// to the given tree (it mutates the tree).
//
// A paragraph is any node that matches the given type and has at least one
// sibling.
type Transformer struct {
	Paragraphs Map
}

// Transform implements the Transformer interface.
func (t Transformer) Transform(n *node.Node) *node.Node {
	var ops []func()
	walk(n, func(n *node.Node) bool {
		if n.PreviousSibling != nil || n.NextSibling != nil {
			for name, typ := range t.Paragraphs {
				if typ == n.Type {
					ops = append(ops, func() {
						p := &node.Node{
							Element: name,
							Type:    node.TypeContainer,
						}
						n.Parent.InsertBefore(p, n)
						n.Parent.RemoveChild(n)
						p.AppendChild(n)
					})
					return false
				}
			}
		}
		return true
	})

	for _, op := range ops {
		op()
	}
	return n
}

func walk(n *node.Node, fn func(n *node.Node) bool) {
	if fn(n) {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c, fn)
		}
	}
}
