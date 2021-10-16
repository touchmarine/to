package seqnum

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/touchmarine/to/aggregator"
	"github.com/touchmarine/to/node"
	"github.com/touchmarine/to/transformer/sequence"
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
	Element           string
	ID                string
	Text              string
	SequentialNumbers []int
}

func (Particle) Particle() {}

func (p Particle) SequentialNumber() string {
	var a []string
	for _, n := range p.SequentialNumbers {
		a = append(a, strconv.FormatUint(uint64(n), 10))
	}
	return strings.Join(a, ".")
}

type Aggregator struct {
	Elements []string
}

type sequencer struct {
	elements  []string
	particles Aggregate
}

func (s *sequencer) aggregate(n *node.Node) {
	walk(n, func(n *node.Node) bool {
		if n.Element == sequence.Element && n.Data != nil {
			sequentialNumbers, ok := n.Data.([]int)
			if !ok {
				panic(fmt.Sprintf("aggregator: unexpected node Data type %#v (%s)", n.Data, n))
			}

			for child := n.FirstChild; child != nil; child = child.NextSibling {
				// walk first level children
				for _, e := range s.elements {
					if e == child.Element {
						s.particles = append(s.particles, Particle{
							Element:           child.Element,
							ID:                n.TextContent(),
							Text:              n.TextContent(),
							SequentialNumbers: sequentialNumbers,
						})
					}
				}
			}
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
