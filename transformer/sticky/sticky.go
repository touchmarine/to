// Package sticky provides a transformer for recognizing and adding stickies to
// node trees.
package sticky

import (
	"github.com/touchmarine/to/node"
)

// Key is a key to sticky's position ("before" or "after") in node.Data.
const Key = "sticky"

// Map is a map of sticky names (keys) to Sticky structs (values).
type Map map[string]Sticky

func (m Map) containsElement(e string) bool {
	for _, s := range m {
		if s.Element == e {
			return true
		}
	}
	return false
}

// Sticky holds the information the transformer uses to recognize a sticky
// element.
type Sticky struct {
	Element string // the sticky element
	Target  string // the element onto which the sticky element sticks to
	After   bool   // whether the sticky must be placed after the target
}

// Transformer recognizes sticky elements based on the given Stickies and adds
// them to the given tree (it mutates the tree).
//
// Transformer recognizes only immediately placed sticky elements. Out of
// multiple consecutive sticky elements, only the one closest to a non-sticky
// element is recognized. An element can have only one sticky element before it
// and one after it.
type Transformer struct {
	Stickies Map
}

// Transform implements the Transformer interface.
func (t Transformer) Transform(n *node.Node) *node.Node {
	ops := t.transform(n)
	for _, op := range ops {
		op()
	}
	return n
}

func (t Transformer) transform(n *node.Node) []func() {
	var ops []func()
	for s := n; s != nil; s = s.NextSibling {
		for name, sticky := range t.Stickies {
			if s.Element != "" && sticky.Element == s.Element {
				var x *node.Node // target node
				if sticky.After {
					x = s.PreviousSibling
				} else {
					x = s.NextSibling
				}
				if x != nil {
					isXSticky := x.Element != "" && t.Stickies.containsElement(x.Element)
					if (!isXSticky || isXSticky && x.Element != s.Element) &&
						(sticky.Target == "" || sticky.Target != "" && sticky.Target == x.Element) {
						ops = append(ops, makeDo(name, sticky, s, x))
					}
				}
			}
		}
	}

	// walk children
	for s := n; s != nil; s = s.NextSibling {
		if s.FirstChild != nil {
			newOps := t.transform(s.FirstChild)
			ops = append(ops, newOps...)
		}
	}

	return ops
}

func makeDo(name string, sticky Sticky, stickynode, targetnode *node.Node) func() {
	return func() {
		s := &node.Node{
			Element: name,
			Type:    node.TypeContainer,
			Data:    node.Data{},
		}

		var child1, child2 *node.Node
		if sticky.After {
			s.Data[Key] = "after"
			child1 = targetnode
			child2 = stickynode
		} else {
			s.Data[Key] = "before"
			child1 = stickynode
			child2 = targetnode
		}

		child1.Parent.InsertBefore(s, child1)

		child1.Parent.RemoveChild(child1)
		s.AppendChild(child1)
		child2.Parent.RemoveChild(child2)
		s.AppendChild(child2)
	}
}
