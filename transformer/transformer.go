package transformer

import (
	"github.com/touchmarine/to/node"
)

type Transformer interface {
	Transform(node *node.Node) *node.Node
}

type Func func(*node.Node) *node.Node

func (f Func) Transform(n *node.Node) *node.Node {
	return f(n)
}

func Apply(n *node.Node, transformers []Transformer) *node.Node {
	for _, transformer := range transformers {
		n = transformer.Transform(n)
	}
	return n
}
