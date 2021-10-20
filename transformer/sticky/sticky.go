package sticky

import (
	"github.com/touchmarine/to/node"
)

const Key = "sticky"

// Map maps Stickies to Elements.
type Map map[string]Sticky

type Sticky struct {
	Name    string `json:"name"`
	Element string `json:"element"`
	Target  string `json:"target"`
	After   bool   `json:"after"`
}

type Transformer struct {
	StickyMap Map
}

// Transform recognizes sticky patterns and creates Sticky nodes. Note that
// it mutates the given nodes.
//
// Transform recognizes only immediately placed sticky elements. Out of
// multiple consecutive sticky elements, only the one closest to a non-sticky
// element is grouped. One element can have one sticky element before it and one
// after it.
func (t Transformer) Transform(n *node.Node) *node.Node {
	targets := t.search(n)
	t.modify(n, targets)
	return n
}

type target struct {
	name           string // sticky name
	position       string // "before" | "after"
	child1, child2 *node.Node
}

// search walks breadth-first and searches for a before or after sticky.
func (t Transformer) search(n *node.Node) []target {
	var targets []target
	for s := n; s != nil; s = s.NextSibling {
		if sticky, ok := t.StickyMap[s.Element]; ok {
			var x *node.Node
			if sticky.After {
				x = s.PreviousSibling
			} else {
				x = s.NextSibling
			}
			if x != nil {
				if _, ok := t.StickyMap[x.Element]; !ok && (sticky.Target == "" || sticky.Target != "" && sticky.Target == x.Element) {
					tt := target{
						name: sticky.Name,
					}
					if sticky.After {
						tt.position = "after"
						tt.child1 = x
						tt.child2 = s
					} else {
						tt.position = "before"
						tt.child1 = s
						tt.child2 = x
					}
					targets = append(targets, tt)
				}
			}
		}
	}

	// walk children
	for s := n; s != nil; s = s.NextSibling {
		if s.FirstChild != nil {
			newTargets := t.search(s.FirstChild)
			targets = append(targets, newTargets...)
		}
	}

	return targets
}

func (t Transformer) modify(n *node.Node, targets []target) {
	for _, target := range targets {
		sticky := &node.Node{
			Element: target.name,
			Type:    node.TypeContainer,
			Data: node.Data{
				Key: target.position,
			},
		}

		target.child1.Parent.InsertBefore(sticky, target.child1)

		target.child1.Parent.RemoveChild(target.child1)
		sticky.AppendChild(target.child1)
		target.child2.Parent.RemoveChild(target.child2)
		sticky.AppendChild(target.child2)
	}
}
