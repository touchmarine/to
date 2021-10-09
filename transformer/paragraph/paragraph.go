package paragraph

import (
	"log"

	"github.com/touchmarine/to/node"
)

const trace = false

func NewTransformer(paragraphName string) *transformer {
	return &transformer{name: paragraphName}
}

type transformer struct {
	name    string       // paragraph name
	targets []*node.Node // nodes to put in paragraph buffer
}

func (t transformer) Transform(n *node.Node) *node.Node {
	walk(n, func(n *node.Node) bool {
		if n.Type == node.TypeLeaf && (n.PrevSibling != nil || n.NextSibling != nil) {
			if trace {
				log.Printf("add target element %s", n.Element)
			}

			t.targets = append(t.targets, n)
			return false
		}
		return true
	})

	for _, target := range t.targets {
		p := &node.Node{
			Element: t.name,
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

/*
func rewrite(n *node.Node, fn func(n *node.Node) (*node.Node, bool)) *node.Node {
	new, cont := fn(n)
	if cont {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			x := rewrite(c, fn)
			if x != nil {
				c=x
			}
		}
	}
	return new
}
*/
