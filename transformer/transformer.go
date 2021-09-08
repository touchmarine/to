package transformer

import (
	"github.com/touchmarine/to/node"
)

type Transformer interface {
	Transform(nodes []node.Node) []node.Node
}

type Func func([]node.Node) []node.Node

func (f Func) Transform(nodes []node.Node) []node.Node {
	return f(nodes)
}

func Apply(nodes []node.Node, transformers []Transformer) []node.Node {
	for _, transformer := range transformers {
		nodes = transformer.Transform(nodes)
	}
	return nodes
}
