package transformer

import (
	"github.com/touchmarine/to/config"
	"github.com/touchmarine/to/node"
	"github.com/touchmarine/to/transformer/composite"
	"github.com/touchmarine/to/transformer/group"
	"github.com/touchmarine/to/transformer/paragraph"
	"github.com/touchmarine/to/transformer/sequence"
	"github.com/touchmarine/to/transformer/sticky"
)

type Transformer interface {
	Transform(nodes []node.Node) []node.Node
}

type TransformerFunc func([]node.Node) []node.Node

func (t TransformerFunc) Transform(nodes []node.Node) []node.Node {
	return t(nodes)
}

func Defaults(conf *config.Config) []Transformer {
	grouper := group.Transformer{conf.Groups}
	compositer := composite.Transformer{conf.TransformerComposites()}
	stickyer := sticky.Transformer{conf.Stickies}

	return []Transformer{
		TransformerFunc(paragraph.Transform),
		grouper,
		compositer,
		stickyer,
		TransformerFunc(sequence.Transform),
	}
}

func Apply(nodes []node.Node, transformers []Transformer) []node.Node {
	for _, transformer := range transformers {
		nodes = transformer.Transform(nodes)
	}
	return nodes
}
