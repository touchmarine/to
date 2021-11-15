package sticky

import (
	"github.com/touchmarine/to/node"
)

const Key = "sticky"

// Map maps Stickies to Elements.
type Map map[string]Sticky

type Sticky struct {
	Name   string
	Target string
	After  bool
}

type Transformer struct {
	Stickies Map
}

// Transform recognizes sticky patterns and creates Sticky nodes. Note that
// it mutates the given nodes.
//
// Transform recognizes only immediately placed sticky elements. Out of
// multiple consecutive sticky elements, only the one closest to a non-sticky
// element is grouped. One element can have one sticky element before it and one
// after it.
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
		if sticky, ok := t.Stickies[s.Element]; ok {
			var x *node.Node
			if sticky.After {
				x = s.PreviousSibling
			} else {
				x = s.NextSibling
			}
			if x != nil {
				if _, ok := t.Stickies[x.Element]; (!ok || ok && x.Element != s.Element) && (sticky.Target == "" || sticky.Target != "" && sticky.Target == x.Element) {
					ops = append(ops, makeDo(sticky, s, x))
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

func makeDo(sticky Sticky, stickynode, targetnode *node.Node) func() {
	return func() {
		s := &node.Node{
			Element: sticky.Name,
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
