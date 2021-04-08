package transformer

import (
	"fmt"
	"strings"
	"to/internal/config"
	"to/internal/node"
)

const trace = false

type grouper struct {
	groups []config.Group
	depth  int

	indent int
}

func Group(groups []config.Group, nodes []node.Node) []node.Node {
	g := grouper{groups, -1, 0}
	return g.group(nodes)
}

func (g *grouper) group(nodes []node.Node) []node.Node {
	g.depth++
	defer func() {
		g.depth--
	}()

	if trace {
		defer g.tracef("group (depth=%d)", g.depth)()
	}

	var grp config.Group
	var open string
	var pos int

	for i := 0; i < len(nodes); i++ {
		n := nodes[i]
		name := n.Node()

		if open == "" {
			var ok bool
			if grp, ok = g.elementGroup(name); ok && g.isGroupable(grp) {
				if trace {
					g.printf("open  %s for %s (i=%d) [1]", grp.Name, name, i)
				}

				open = name
				pos = i
			}
		} else if name != open {
			end := i - 1

			if trace {
				g.printf("close %s for %s (i=%d, group=%d-%d)", grp.Name, open, i, pos, end)
			}

			children := node.NodesToBlocks(nodes[pos : end+1])
			grpNode := &node.Group{grp.Name, children}

			nodes[pos] = grpNode
			if end-pos > 0 {
				if trace {
					g.printf("cut nodes %d-%d [1]", pos+1, end+1)
				}

				nodes = cut(nodes, pos+1, end+1)
				i -= end - pos
			}

			var ok bool
			if grp, ok = g.elementGroup(name); ok && g.isGroupable(grp) {
				if trace {
					g.printf("open  %s for %s (i=%d) [2]", grp.Name, name, i)
				}

				open = name
				pos = i
			} else {
				open = ""
				pos = 0
			}
		}

		if m, ok := n.(node.SettableBlockChildren); ok {
			grouped := g.group(node.BlocksToNodes(m.BlockChildren()))
			m.SetBlockChildren(node.NodesToBlocks(grouped))
		}
	}

	if open != "" {
		l := len(nodes)
		end := l - 1

		if trace {
			g.printf("close %s for %s (i=%d, group=%d-%d) [last]", grp.Name, open, l, pos, end)
		}

		children := node.NodesToBlocks(nodes[pos:])
		grpNode := &node.Group{grp.Name, children}

		nodes[pos] = grpNode
		if end-pos > 0 {
			if trace {
				g.printf("cut nodes %d-%d [2]", pos+1, end+1)
			}

			nodes = cut(nodes, pos+1, end+1)
		}
	}

	return nodes
}

func (g *grouper) elementGroup(element string) (config.Group, bool) {
	for _, grp := range g.groups {
		if grp.Element == element {
			return grp, true
		}
	}
	return config.Group{}, false
}

func (g *grouper) isGroupable(grp config.Group) bool {
	if trace {
		defer g.tracef("isGroupable (%s, groupNested=%t)", grp.Name, grp.GroupNested)()
	}

	var t bool
	if grp.GroupNested {
		t = true
	} else {
		t = g.depth < 1
	}

	if trace {
		g.printf("return %t", t)
	}

	return t
}

func (g *grouper) tracef(format string, v ...interface{}) func() {
	return g.trace(fmt.Sprintf(format, v...))
}

func (g *grouper) trace(msg string) func() {
	g.printf("%s (", msg)
	g.indent++

	return func() {
		g.indent--
		g.print(")")
	}
}

func (g *grouper) printf(format string, v ...interface{}) {
	g.print(fmt.Sprintf(format, v...))
}

func (g *grouper) print(msg string) {
	fmt.Println(strings.Repeat("\t", g.indent) + msg)
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
