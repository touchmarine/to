package transformer

import (
	"fmt"
	"strconv"
	"strings"
	"to/internal/config"
	"to/internal/node"
)

type sequencer struct {
	elems  []config.Element
	seqMap map[string]map[uint]uint
}

func Sequence(elements []config.Element, nodes []node.Node) []node.Node {
	s := &sequencer{
		elems:  elements,
		seqMap: make(map[string]map[uint]uint),
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

			s.incSeqNum(name, rank)
			seqNum := s.seqNum(name, rank)

			nodes[i] = &node.SeqNumBox{n, seqNum}
		}
	}
	return nodes
}

func (s *sequencer) incSeqNum(name string, rank uint) {
	if _, ok := s.seqMap[name]; !ok {
		s.seqMap[name] = make(map[uint]uint)
	}

	s.seqMap[name][rank]++

	for rk, _ := range s.seqMap[name] {
		if rk > rank {
			s.seqMap[name][rk] = 0
		}
	}
}

func (s *sequencer) seqNum(name string, rank uint) string {
	m, ok := s.seqMap[name]
	if !ok {
		panic(fmt.Sprintf("sequencer.seqNum: missing map for ranked node %s", name))
	}

	el, ok := s.element(name)
	if !ok {
		panic(fmt.Sprintf("sequencer.seqNum: missing element in config for node %s", name))
	}

	minRank := el.MinRank

	var seq []string
	for i := minRank; i <= rank; i++ {
		var seqNum uint64
		u, ok := m[i]
		if ok {
			seqNum = uint64(u)
		}
		seq = append(seq, strconv.FormatUint(seqNum, 10))
	}

	return strings.Join(seq, ".")
}

func (s *sequencer) element(name string) (config.Element, bool) {
	for _, el := range s.elems {
		if el.Name == name {
			return el, true
		}
	}
	return config.Element{}, false
}
