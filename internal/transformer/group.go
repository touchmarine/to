package transformer

import (
	"fmt"
	"to/internal/config"
	"to/internal/node"
)

const trace = false

func Group(groups []config.Group, nodes []node.Node) []node.Node {
	var group *config.Group
	var open string
	var pos int

	for i := 0; i < len(nodes); i++ {
		n := nodes[i]
		name := n.Node()

		if open == "" {
			if isGrouped(groups, name) {
				group = elementGroup(groups, name)
				if trace {
					printf("open  %s for %s (i=%d) [1]", group.Name, name, i)
				}

				open = name
				pos = i
			}
		} else if name != open {
			end := i - 1

			if trace {
				printf("close %s for %s (i=%d, group=%d-%d)", group.Name, open, i, pos, end)
			}

			children := node.NodesToBlocks(nodes[pos : end+1])
			g := &node.Group{group.Name, children}

			nodes[pos] = g
			if end-pos > 0 {
				if trace {
					printf("cut nodes %d-%d [1]", pos+1, end+1)
				}

				nodes = cut(nodes, pos+1, end+1)
				i -= end - pos
			}

			if isGrouped(groups, name) {
				group = elementGroup(groups, name)
				if trace {
					printf("open  %s for %s (i=%d) [2]", group.Name, name, i)
				}

				open = name
				pos = i
			} else {
				open = ""
				pos = 0
			}
		}

		if group != nil && group.GroupNested {
			if m, ok := n.(node.SettableBlockChildren); ok {
				grouped := Group(groups, node.BlocksToNodes(m.BlockChildren()))
				m.SetBlockChildren(node.NodesToBlocks(grouped))
			}
		}
	}

	if open != "" {
		l := len(nodes)
		end := l - 1

		if trace {
			printf("close %s for %s (i=%d, group=%d-%d) [last]", group.Name, open, l, pos, end)
		}

		children := node.NodesToBlocks(nodes[pos:])
		g := &node.Group{group.Name, children}

		nodes[pos] = g
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

func elementGroup(groups []config.Group, element string) *config.Group {
	for _, group := range groups {
		if group.Element == element {
			return &group
		}
	}
	panic(fmt.Sprintf("element %s group not found", element))
}

func isGrouped(groups []config.Group, s string) bool {
	for _, group := range groups {
		if group.Element == s {
			return true
		}
	}
	return false
}
