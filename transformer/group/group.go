// Package group provides a transformer for recognizing and adding groups to
// node trees.
package group

import (
	"github.com/touchmarine/to/node"
)

// Map is a map of group names (keys) to elements (values). Elements tell the
// transformer which elements can it group.
type Map map[string]string

func (m Map) firstByElement(e string) (string, bool) {
	for g, el := range m {
		if el == e {
			return g, true
		}
	}
	return "", false
}

// Transformer recognizes groups based on the given Groups and adds them to the
// given tree (it mutates the tree). A group is a contiguous sequence of the
// same sibling elements.
//
// Transformer supports only one group per element. Using multiple groups
// targeting the same element is considered undefined behaviour.
type Transformer struct {
	Groups Map
}

// Transform implements the Transformer interface.
func (t Transformer) Transform(n *node.Node) *node.Node {
	var ops []func()
	walkBreadthFirstStack(n, func(nodes []*node.Node) {
		name, start, end := "", -1, 0

		for i, n := range nodes {
			gname, found := t.Groups.firstByElement(n.Element)

			if name != "" {
				// a group is open
				if found && gname == name {
					// group continues
					end++
					continue
				} else {
					// group ends
					ops = append(ops, makeAddGroup(name, nodes[start:end]))
					name, start, end = "", -1, 0
				}
			}

			if found {
				// start of a new group
				name = gname
				start = i
				end = i + 1
			}
		}

		if name != "" {
			ops = append(ops, makeAddGroup(name, nodes[start:end]))
		}
	})
	for _, op := range ops {
		op()
	}
	return n
}

func makeAddGroup(name string, children []*node.Node) func() {
	return func() {
		g := &node.Node{
			Element: name,
			Type:    node.TypeContainer,
		}
		if len(children) > 0 {
			children[0].Parent.InsertBefore(g, children[0])
		}
		for _, c := range children {
			c.Parent.RemoveChild(c)
			g.AppendChild(c)
		}
	}
}

func walkBreadthFirstStack(n *node.Node, fn func(nodes []*node.Node)) {
	nodes := []*node.Node{n}

	for s := n.NextSibling; s != nil; s = s.NextSibling {
		nodes = append(nodes, s)
	}

	fn(nodes)

	for _, n := range nodes {
		if n.FirstChild != nil {
			walkBreadthFirstStack(n.FirstChild, fn)
		}
	}
}
