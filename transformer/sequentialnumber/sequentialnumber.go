package sequentialnumber

import (
	"strconv"
	"strings"

	"github.com/touchmarine/to/node"
	"github.com/touchmarine/to/parser"
)

const Key = "sequentialNumber"

func Transform(n *node.Node) *node.Node {
	m := map[string]map[int]int{}
	walk(n, func(n *node.Node) bool {
		rank, ok := n.Data[parser.KeyRank].(int)
		if ok {
			if _, ok := m[n.Element]; !ok {
				m[n.Element] = map[int]int{}
			}
			m[n.Element][rank]++

			for r, _ := range m[n.Element] {
				if r > rank {
					// clear deeper rank
					m[n.Element][r] = 0
				}
			}

			var sequentialNumbers []int
			for i := 2; i <= rank; i++ {
				sequentialNumbers = append(sequentialNumbers, m[n.Element][i])
			}

			n.Data[Key] = sequentialNumber(sequentialNumbers)
		}
		return true
	})
	return n
}

func walk(n *node.Node, fn func(n *node.Node) bool) {
	if fn(n) {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c, fn)
		}
	}
}

func sequentialNumber(sequentialNumbers []int) string {
	var p []string
	for _, n := range sequentialNumbers {
		p = append(p, strconv.FormatUint(uint64(n), 10))
	}
	return strings.Join(p, ".")
}
