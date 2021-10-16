package paragraph

import (
	"log"

	"github.com/touchmarine/to/node"
)

const trace = false

type Transformer struct {
	Name string // paragraph name
}

func (t Transformer) Transform(n *node.Node) *node.Node {
	var targets []*node.Node

	walk(n, func(n *node.Node) bool {
		if n.Type == node.TypeLeaf && (n.PrevSibling != nil || n.NextSibling != nil) {
			if trace {
				log.Printf("add target element %s", n.Element)
			}

			targets = append(targets, n)
			return false
		}
		return true
	})

	for _, target := range targets {
		p := &node.Node{
			Element: t.Name,
			Type:    node.TypeContainer,
		}
		target.Parent.InsertBefore(p, target)
		target.Parent.RemoveChild(target)
		p.AppendChild(target)
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