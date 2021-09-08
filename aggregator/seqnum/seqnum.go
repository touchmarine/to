package seqnum

import (
	"fmt"
	"github.com/touchmarine/to/aggregator"
	"github.com/touchmarine/to/node"
)

type Aggregate []Particle

func (a Aggregate) Particles() []aggregator.Particle {
	return toAggregatorParticles(a)
}

func (a Aggregator) Aggregate(nodes []node.Node) aggregator.Aggregate {
	s := sequencer{
		elements: a.Elements,
	}
	s.aggregate(nodes)
	return aggregator.Aggregate(s.particles)
}

type Particle struct {
	Element           string
	ID                string
	Text              string
	SequentialNumbers []int
	SequentialNumber  string
}

func (Particle) Particle() {}

type Aggregator struct {
	Elements []string
}

type sequencer struct {
	elements  []string
	particles Aggregate
}

func (s *sequencer) aggregate(nodes []node.Node) {
	for _, n := range nodes {
		var p Particle

		switch m := n.(type) {
		case *node.Sticky:
			s.aggregate(node.BlocksToNodes(m.BlockChildren()))
		case node.Boxed:
			switch k := n.(type) {
			case *node.SequentialNumberBox:
				p.SequentialNumbers = k.SequentialNumbers
				p.SequentialNumber = k.SequentialNumber()
			default:
				panic(fmt.Sprintf("seqnum: unexpected Boxed node %T", n))
			}

			n = m.Unbox()
		}

		name := n.Node()
		for _, e := range s.elements {
			if e == name {
				p.Element = name
				txt := node.ExtractText(n)
				p.ID = txt
				p.Text = txt

				s.particles = append(s.particles, p)
			}
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
