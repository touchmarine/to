package group

import (
	"log"

	"github.com/touchmarine/to/node"
)

const trace = false

// Map maps Groups by Elements.
type Map map[string]Group

type Group struct {
	Name    string
	Element string
}

type Transformer struct {
	GroupMap Map
}

type target struct {
	name     string // group name
	children []*node.Node
}

func (t Transformer) Transform(n *node.Node) *node.Node {
	var targets []target

	if trace {
		log.Printf("t.GroupMap = %+v\n", t.GroupMap)
	}

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
					targets = append(targets, target{
						name:     name,
						children: nodes[start:end],
					})
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
			targets = append(targets, target{
				name:     name,
				children: nodes[start:end],
			})
		}
	})

	if trace {
		log.Printf("targets = %+v\n", targets)
	}
	for _, target := range targets {
		group := &node.Node{
			Element: target.name,
			Type:    node.TypeContainer,
		}

		if len(target.children) > 0 {
			target.children[0].Parent.InsertBefore(group, target.children[0])
		}

		for _, child := range target.children {
			if trace {
				log.Printf("child = %+v\n", child)
			}

			child.Parent.RemoveChild(child)
			group.AppendChild(child)
		}
	}

	return n
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
