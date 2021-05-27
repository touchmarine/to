package transformer

import (
	"bytes"
	"fmt"
	"github.com/touchmarine/to/node"
)

func BlankLine(nodes []node.Node) []node.Node {
	for i := 0; i < len(nodes); i++ {
		n := nodes[i]

		if _, ok := n.(node.Boxed); ok {
			switch m := n.(type) {
			case *node.Hat:
				b := BlankLine([]node.Node{m.Unbox()})

				var k node.Node
				if l := len(b); l == 1 {
					k = b[0]
				} else if l > 1 {
					panic("transformer: expected 1 node")
				}
				nodes[i] = &node.Hat{m.Lines(), k}
				continue
			case *node.SeqNumBox:
				n = m.Unbox()
			default:
				panic(fmt.Sprintf("transformer: unexpected Boxed node %T", n))
			}

			if n == nil {
				continue
			}
		}

		if ln, ok := n.(*node.Line); ok && n.Node() == "Line" && isBlank(ln) {
			nodes = append(nodes[:i], nodes[i+1:]...) // delete
			i--
			continue
		}

		if _, ok := n.(node.Lines); ok {
			if m, ok := n.(node.SettableLines); ok {
				var lns [][]byte
				for _, ln := range m.Lines() {
					l := bytes.Trim(ln, " \t")
					if len(l) > 0 {
						lns = append(lns, l)
					}
				}
				m.SetLines(lns)
				continue
			}
		}

		if m, ok := n.(node.SettableBlockChildren); ok {
			a := BlankLine(node.BlocksToNodes(m.BlockChildren()))
			m.SetBlockChildren(node.NodesToBlocks(a))
		} else {
			if _, ok := n.(node.BlockChildren); ok {
				panic(fmt.Sprintf("transformer: node %T does not implement SettableBlockChildren", n))
			}
		}
	}

	return nodes
}
