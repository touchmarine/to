// Package aggregator defines an interface for aggregating data from node trees.
//
// Aggregators traverse the node tree and aggregate (collect) data we are
// interested in. For example, we use the sequential number aggregator to
// generate table of contents by aggregating heading nodes and attaching their
// sequential numbers to them.
//
// In our context:
// - aggregate=result
// - aggregator=implementation
package aggregator

import (
	"github.com/touchmarine/to/node"
)

// Aggregators is a map of aggregator names to Aggregators.
type Aggregators map[string]Aggregator

// Aggregator aggregates (collects) data from node trees.
type Aggregator interface {
	Aggregate(n *node.Node) Aggregate
}

// AggregatorFunc is a convenience type that implements the Aggregator interface
// for the given function.
type AggregatorFunc func(*node.Node) Aggregate

// Aggregate implements the Aggregator interface.
func (a AggregatorFunc) Aggregate(n *node.Node) Aggregate {
	return a(n)
}

// Aggregates is a map of aggregate names to Aggregates.
type Aggregates map[string]Aggregate

// Aggregate is an aggregate of data we are interested in.
type Aggregate interface {
	AnAggregate() // dummy method to avoid type errors
}

// Apply applies the given aggregators and returns the resulting aggregates.
func Apply(n *node.Node, aggregators Aggregators) Aggregates {
	m := Aggregates{}
	for name, a := range aggregators {
		m[name] = a.Aggregate(n)
	}
	return m
}
