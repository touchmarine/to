package transformer

import (
	"fmt"
	"github.com/touchmarine/to/node"
)

type sequencer struct {
	seqMap map[string]map[int]int
}

func Sequence(nodes []node.Node) []node.Node {
	s := &sequencer{
		seqMap: make(map[string]map[int]int),
	}
	nodes = s.sequence(nodes)
	return nodes
}

func (s *sequencer) sequence(nodes []node.Node) []node.Node {
	for i := 0; i < len(nodes); i++ {
		n := nodes[i]

		if m, ok := n.(node.Ranked); ok {
			name := n.Node()
			rank := m.Rank()

			s.increment(name, rank)
			seqNums := s.seqNums(name, rank)

			nodes[i] = &node.SeqNumBox{n, seqNums}
		}
	}
	return nodes
}

func (s *sequencer) increment(name string, rank int) {
	if _, ok := s.seqMap[name]; !ok {
		s.seqMap[name] = make(map[int]int)
	}

	s.seqMap[name][rank]++

	for r, _ := range s.seqMap[name] {
		if r > rank {
			// clear deeper rank
			s.seqMap[name][r] = 0
		}
	}
}

func (s *sequencer) seqNums(name string, rank int) []int {
	m, ok := s.seqMap[name]
	if !ok {
		panic(fmt.Sprintf("transformer: missing map for node %s", name))
	}

	var seq []int
	for i := 2; i <= rank; i++ {
		seq = append(seq, m[i])
	}

	return seq
}
