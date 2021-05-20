package transformer

import (
	"fmt"
	"github.com/touchmarine/to/internal/node"
	"log"
)

func Paragraph(nodes []node.Node) []node.Node {
	beg := -1
	end := -1
	var added bool
	for i := 0; i < len(nodes); i++ {
		n := nodes[i]

	beginning:
		if beg < 0 {
			if isLine(n) && !isBlank1(n) {
				beg = i
			}
		} else {
			if end < 0 {
				if isLine(n) && !isBlank1(n) {
					// paragraph continues
				} else {
					end = i
					if trace {
						log.Printf("para %d-%d", beg, end)
					}
				}
			}

			if end > -1 {
				if isLine(n) && isBlank1(n) {
					// do not add para if blank line
				} else {
					if trace {
						log.Printf("add para")
					}
					children := node.NodesToBlocks(nodes[beg:end])
					p := &node.Group{"Paragraph", children}

					nodes[beg] = p
					if end-beg > 1 {
						if trace {
							log.Printf("cut nodes %d-%d", beg+1, end)
						}
						nodes = cut(nodes, beg+1, end)
						i -= end - beg - 1
					}

					beg = -1
					end = -1
					added = true
					goto beginning
				}
			}
		}

		if m, ok := n.(node.SettableBlockChildren); ok {
			breaked := Paragraph(node.BlocksToNodes(m.BlockChildren()))
			m.SetBlockChildren(node.NodesToBlocks(breaked))
		} else {
			if _, ok := n.(node.BlockChildren); ok {
				panic(fmt.Sprintf("Paragraph: node %T does not implement SettableBlockChildren", n))
			}
		}
	}

	if beg > -1 && added {
		if end < 0 {
			end = len(nodes)
		}
		if trace {
			log.Printf("add para %d-%d", beg, end)
		}

		children := node.NodesToBlocks(nodes[beg:end])
		p := &node.Group{"Paragraph", children}

		nodes[beg] = p
		if end-beg > 1 {
			if trace {
				log.Printf("cut nodes %d-%d", beg+1, end)
			}
			nodes = cut(nodes, beg+1, end)
		}
	}

	return nodes
}

func isLine(n node.Node) bool {
	_, ok := n.(*node.Line)
	return ok && n.Node() == "Line"
}
