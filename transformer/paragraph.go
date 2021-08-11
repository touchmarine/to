package transformer

import (
	"fmt"
	"github.com/touchmarine/to/node"
	"log"
)

func Paragraph(nodes []node.Node) []node.Node {
	for i := 0; i < len(nodes); i++ {
		n := nodes[i]

		if len(nodes) > 1 {
			if block, ok := n.(*node.BasicBlock); ok && n.Node() == "TextBlock" {
				if trace {
					log.Printf("add paragraph")
				}

				p := &node.Group{"Paragraph", []node.Block{block}}
				nodes[i] = p
			}
		}

		if m, ok := n.(node.SettableBlockChildren); ok {
			breaked := Paragraph(node.BlocksToNodes(m.BlockChildren()))
			m.SetBlockChildren(node.NodesToBlocks(breaked))
		} else {
			if _, ok := n.(node.BlockChildren); ok {
				panic(fmt.Sprintf("transformer: node %T does not implement SettableBlockChildren", n))
			}
		}
	}

	return nodes
}
