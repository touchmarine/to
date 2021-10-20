package paragraph

import (
	"github.com/touchmarine/to/node"
)

type Transformer struct {
	ParagraphMap map[node.Type]string // map[nodeType]groupName
}

func (t Transformer) Transform(n *node.Node) *node.Node {
	var ops []func()
	walk(n, func(n *node.Node) bool {
		if name, ok := t.ParagraphMap[n.Type]; ok && (n.PreviousSibling != nil || n.NextSibling != nil) {
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
