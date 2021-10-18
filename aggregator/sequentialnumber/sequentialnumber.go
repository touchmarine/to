package sequentialnumber

import (
	"fmt"

	"github.com/touchmarine/to/aggregator"
	"github.com/touchmarine/to/node"
	"github.com/touchmarine/to/transformer/sequentialnumber"
)

type Aggregate []Particle

func (a Aggregate) Particles() []aggregator.Particle {
	return toAggregatorParticles(a)
}

func (a Aggregator) Aggregate(n *node.Node) aggregator.Aggregate {
	s := sequencer{
		elements: a.Elements,
	}
	s.aggregate(n)
	return aggregator.Aggregate(s.particles)
}

type Particle struct {
	Element          string
	ID               string
	Text             string
	SequentialNumber string
}

func (Particle) Particle() {}

type Aggregator struct {
	Elements []string
}

type sequencer struct {
	elements  []string
	particles Aggregate
}

func (s *sequencer) aggregate(n *node.Node) {
	walk(n, func(n *node.Node) bool {
		if v, ok := n.Data[sequentialnumber.Key]; ok {
			seqnum, isString := v.(string)
			if !isString {
				panic(fmt.Sprintf("aggregator: sequentialnumber is not a string (%T)", v))
			}

			s.particles = append(s.particles, Particle{
				Element:          n.Element,
				ID:               n.TextContent(),
				Text:             n.TextContent(),
				SequentialNumber: seqnum,
			})
		}
		return true
	})
}

func walk(n *node.Node, fn func(n *node.Node) bool) {
	if fn(n) {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c, fn)
		}
	}
}

func toAggregatorParticles(a []Particle) []aggregator.Particle {
	b := make([]aggregator.Particle, len(a))
	for i := range a {
		b[i] = a[i]
	}
	return b
}
