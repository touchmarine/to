package transformer

import (
	"fmt"
	"to/internal/node"
)

const trace = false

var groupMap = map[string]string{
	"NumberedListDot":   "NumberedListItemDot",
	"NumberedListParen": "NumberedListItemParen",
}

func Group(nodes []node.Node) []node.Node {
	var open string
	var pos int

	for i := 0; i < len(nodes); i++ {
		n := nodes[i]
		name := n.Node()

		if open == "" {
			if isGrouped(name) {
				if trace {
					printf("open  %s for %s (i=%d) [1]", groupName(name), name, i)
				}

				open = name
				pos = i
			}
		} else if name != open {
			gname := groupName(open)
			end := i - 1

			if trace {
				printf("close %s for %s (i=%d, group=%d-%d)", gname, open, i, pos, end)
			}

			children := node.NodesToBlocks(nodes[pos : end+1])
			group := &node.Group{gname, children}

			nodes[end] = group
			if end-pos > 0 {
				if trace {
					printf("cut nodes %d-%d [1]", pos, end)
				}

				nodes = cut(nodes, pos, end)
				i -= end - pos
			}

			if isGrouped(name) {
				if trace {
					printf("open  %s for %s (i=%d) [2]", groupName(name), name, i)
				}

				open = name
				pos = i
			} else {
				open = ""
				pos = 0
			}
		}
	}

	if open != "" {
		gname := groupName(open)
		l := len(nodes)
		end := l - 1

		if trace {
			printf("close %s for %s (i=%d, group=%d-%d) [last]", gname, open, l, pos, end)
		}

		children := node.NodesToBlocks(nodes[pos:])
		group := &node.Group{gname, children}

		nodes[pos] = group
		if end-pos > 0 {
			if trace {
				printf("cut nodes %d-%d [2]", pos+1, end+1)
			}

			nodes = cut(nodes, pos+1, end+1)
		}
	}

	return nodes
}

func printf(format string, a ...interface{}) {
	fmt.Printf(format+"\n", a...)
}

// https://github.com/golang/go/wiki/SliceTricks
func cut(a []node.Node, i, j int) []node.Node {
	copy(a[i:], a[j:])
	for k, n := len(a)-j+i, len(a); k < n; k++ {
		a[k] = nil
	}
	a = a[:len(a)-j+i]
	return a
}

func isGrouped(s string) bool {
	for _, node := range groupMap {
		if node == s {
			return true
		}
	}
	return false
}

func groupName(s string) string {
	for group, node := range groupMap {
		if node == s {
			return group
		}
	}
	panic(fmt.Sprintf("node %s group name not found", s))
}
