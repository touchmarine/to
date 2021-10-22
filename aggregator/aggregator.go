// package aggregator contains functions and types related to aggregators.
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

// AggregatorMap=map["aggregatorName"]["aggregateName"]Aggregator
type AggregatorMap map[string]map[string]Aggregator

// Aggregator is an object that aggregates (collects) nodes.
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

// AggregateMap=map["aggregatorName"]map["aggregateName"]Aggregate
//
// For example:
// 	aggregatorName="sequentialNumbers",
// 	aggregateName="numberedHeadings" and
// 	Aggregate is the result.
type AggregateMap map[string]map[string]Aggregate

// Aggregate is an aggregate of data we are interested in.
type Aggregate interface {
	AnAggregate() // dummy method to avoid type errors
}

// Apply applies the given aggregators to the given nodes.
func Apply(n *node.Node, aggregatorMap AggregatorMap) AggregateMap {
	m := AggregateMap{}
	for namea, ma := range aggregatorMap {
		if m[namea] == nil {
			m[namea] = map[string]Aggregate{}
		}
		for nameb, aggregator := range ma {
			m[namea][nameb] = aggregator.Aggregate(n)
		}
	}
	return m
}
