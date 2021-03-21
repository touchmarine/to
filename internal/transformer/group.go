package transformer

import (
	"fmt"
	"to/internal/config"
	"to/internal/node"
)

const trace = false

func Group(groups []config.Group, nodes []node.Node) []node.Node {
	var open string
	var pos int

	for i := 0; i < len(nodes); i++ {
		n := nodes[i]
		name := n.Node()

		if open == "" {
			if isGrouped(groups, name) {
				if trace {
					printf("open  %s for %s (i=%d) [1]", groupName(groups, name), name, i)
				}

				open = name
				pos = i
			}
		} else if name != open {
			gname := groupName(groups, open)
			end := i - 1

			if trace {
				printf("close %s for %s (i=%d, group=%d-%d)", gname, open, i, pos, end)
			}

			children := node.NodesToBlocks(nodes[pos : end+1])
			group := &node.Group{gname, children}

			nodes[pos] = group
			if end-pos > 0 {
				if trace {
					printf("cut nodes %d-%d [1]", pos+1, end+1)
				}

				nodes = cut(nodes, pos+1, end+1)
				i -= end - pos
			}

			if isGrouped(groups, name) {
				if trace {
					printf("open  %s for %s (i=%d) [2]", groupName(groups, name), name, i)
				}

				open = name
				pos = i
			} else {
				open = ""
				pos = 0
			}
		}

		if m, ok := n.(node.SettableBlockChildren); ok {
			grouped := Group(groups, node.BlocksToNodes(m.BlockChildren()))
			m.SetBlockChildren(node.NodesToBlocks(grouped))
		}
	}

	if open != "" {
		gname := groupName(groups, open)
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

func isGrouped(groups []config.Group, s string) bool {
	for _, group := range groups {
		if group.Element == s {
			return true
		}
	}
	return false
}

func groupName(groups []config.Group, s string) string {
	for _, group := range groups {
		if group.Element == s {
			return group.Name
		}
	}
	panic(fmt.Sprintf("node %s group name not found", s))
}
