package group

import (
	"github.com/touchmarine/to/node"
)

// Map maps Groups by Elements.
type Map map[string]Group

type Group struct {
	Name    string
	Element string
}

type Transformer struct {
	GroupMap Map
}

func (t Transformer) Transform(n *node.Node) *node.Node {
	var ops []func()
	walkBreadthFirstStack(n, func(nodes []*node.Node) {
		name, start, end := "", -1, 0

		for i, n := range nodes {
			group, found := t.GroupMap[n.Element]

			if name != "" {
				// a group is open
				if found && group.Name == name {
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
				name = group.Name
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
