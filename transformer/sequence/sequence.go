package sequence

import (
	"fmt"

	"github.com/touchmarine/to/node"
)

const Element = "sequentialNumber"

type Transformer struct {
	seqMap map[string]map[int]int
}

type target struct {
	node              *node.Node
	sequentialNumbers []int
}

func (t Transformer) Transform(n *node.Node) *node.Node {
	var targets []target

	walk(n, func(n *node.Node) bool {
		if n.Type == node.TypeRankedHanging {
			rank, isInt := n.Data.(int)
			if isInt {
				t.increment(n.Element, rank)

				targets = append(targets, target{
					node:              n,
					sequentialNumbers: t.sequentialNumbers(n.Element, rank),
				})
			}
		}

		return true
	})

	for _, target := range targets {
		container := &node.Node{
			Element: Element,
			Type:    node.TypeContainer,
			Data:    target.sequentialNumbers,
		}
		target.node.Parent.InsertBefore(container, target.node)
		target.node.Parent.RemoveChild(target.node)
		container.AppendChild(target.node)
	}

	return n
}

func (t *Transformer) increment(name string, rank int) {
	if t.seqMap == nil {
		t.seqMap = map[string]map[int]int{}
	}

	if _, ok := t.seqMap[name]; !ok {
		t.seqMap[name] = make(map[int]int)
	}

	t.seqMap[name][rank]++

	for r, _ := range t.seqMap[name] {
		if r > rank {
			// clear deeper rank
			t.seqMap[name][r] = 0
		}
	}
}

func (t Transformer) sequentialNumbers(element string, rank int) []int {
	m, ok := t.seqMap[element]
	if !ok {
		panic(fmt.Sprintf("transformer: missing map for element (%s)", element))
	}

	var seq []int
	for i := 2; i <= rank; i++ {
		seq = append(seq, m[i])
	}

	return seq
}

func walk(n *node.Node, fn func(n *node.Node) bool) {
	if fn(n) {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c, fn)
		}
	}
}
