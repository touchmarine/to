// package aggregator contains functions and types related to aggregators.
//
// Aggregators traverse nodes and aggregate (collect) data we are interested in.
// For example, we use the sequential number aggregator to generate table of
// contents by aggregating heading nodes and attaching their sequential numbers
// to them.
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
	Aggregate(nodes []node.Node) Aggregate
}

// AggregatorFunc is a convenience type that implements the Aggregator interface
// for the given function.
type AggregatorFunc func([]node.Node) Aggregate

// Aggregate implements the Aggregator interface.
func (a AggregatorFunc) Aggregate(nodes []node.Node) Aggregate {
	return a(nodes)
}

// AggregateMap=map["aggregatorName"]map["aggregateName"]Aggregate
//
// For example:
// 	aggregatorName="sequentialNumbers",
// 	aggregateName="numberedHeadings" and
// 	[]Aggregate is the result.
type AggregateMap map[string]map[string]Aggregate

// Aggregate is an aggregate of Particles.
type Aggregate interface {
	Particles() []Particle
}

// Particle is an element of the aggregate.
type Particle interface {
	Particle()
}

// Apply applies the given aggregators to the given nodes.
func Apply(nodes []node.Node, aggregatorMap AggregatorMap) AggregateMap {
	m := AggregateMap{}
	for namea, ma := range aggregatorMap {
		if m[namea] == nil {
			m[namea] = map[string]Aggregate{}
		}

		for nameb, aggregator := range ma {
			m[namea][nameb] = aggregator.Aggregate(nodes)
		}
	}
	return m
}
