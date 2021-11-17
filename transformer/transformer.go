// package transformer provides a transformer interface and related utility
// functions. Transformers transform a node tree. They usually traverse the node
// tree, recognize patterns, and transform the tree by attaching, adding, or
// removing nodes.
package transformer

import (
	"github.com/touchmarine/to/node"
)

// Transformer transforms the given node tree and returns the transformed
// variant.
//
// Transformers can also first inspect the given tree and transform it based on
// the inspection results. This is common in To—transformers first recognize
// patterns (like groups) and then, based on the results, transform the tree
// (add the groups).
type Transformer interface {
	Transform(node *node.Node) *node.Node
}

// Func is like what http.HandlerFunc is to http.Handler—an adapter to allow the
// use of ordinary functions as Transformers. If f is a function with the
// appropriate signature, Func(f) is a Transformer that calls and returns f(n).
type Func func(*node.Node) *node.Node

// Transform calls and returns f(n).
func (f Func) Transform(n *node.Node) *node.Node {
	return f(n)
}

// Group is a list of transformers that is itself a transformer. It can be used
// to group related transformers.
type Group []Transformer

// Transform implements the Transformer interface.
func (g Group) Transform(n *node.Node) *node.Node {
	for _, t := range g {
		n = t.Transform(n)
	}
	return n
}
