package transformer

import (
	"fmt"
	"github.com/touchmarine/to/config"
	"github.com/touchmarine/to/node"
	"strings"
)

type grouper struct {
	groups []config.Group
	indent int
}

func Group(groups []config.Group, nodes []node.Node) []node.Node {
	g := grouper{groups, 0}
	return g.group(nodes)
}

func (g *grouper) group(nodes []node.Node) []node.Node {
	if trace {
		defer g.trace("group")()
	}

	var grp config.Group
	var lines [][]byte
	var open string
	var pos int

	for i := 0; i < len(nodes); i++ {
		n := nodes[i]
		name := n.Node()

		if m, ok := n.(node.Boxed); ok {
			if open != "" && lines != nil {
				// close
				end := i - 1

				if trace {
					g.printf("close %s for %s (i=%d, group=%d-%d)", grp.Name, open, i, pos, end)
				}

				children := node.NodesToBlocks(nodes[pos : end+1])
				grpNode := &node.Group{grp.Name, children}
				nod := &node.Hat{lines, grpNode}

				nodes[pos] = nod
				if end-pos > 0 {
					if trace {
						g.printf("cut nodes %d-%d [1]", pos+1, end+1)
					}

					nodes = cut(nodes, pos+1, end+1)
					i -= end - pos
				}

				lines = nil
				open = ""
				pos = 0
			}

			switch k := n.(type) {
			case *node.Hat:
				lines = k.Lines()
			case *node.SeqNumBox:
			default:
				panic(fmt.Sprintf("transformer: unexpected Boxed node %T", n))
			}

			n = m.Unbox()
			if n == nil {
				continue
			}

			name = n.Node()

			if trace {
				g.printf("unboxed %T into %s", m, name)
			}
		}

		if open == "" {
			var ok bool
			if grp, ok = g.elementGroup(name); ok {
				if trace {
					g.printf("open  %s for %s (i=%d) [1]", grp.Name, name, i)
				}

				open = name
				pos = i
			} else {
				lines = nil
			}
		} else if name != open {
			end := i - 1

			if trace {
				g.printf("close %s for %s (i=%d, group=%d-%d)", grp.Name, open, i, pos, end)
			}

			children := node.NodesToBlocks(nodes[pos : end+1])
			grpNode := &node.Group{grp.Name, children}

			var nod node.Node
			if lines == nil {
				nod = grpNode
			} else {
				nod = &node.Hat{lines, grpNode}
				lines = nil
			}

			nodes[pos] = nod
			if end-pos > 0 {
				if trace {
					g.printf("cut nodes %d-%d [1]", pos+1, end+1)
				}

				nodes = cut(nodes, pos+1, end+1)
				i -= end - pos
			}

			var ok bool
			if grp, ok = g.elementGroup(name); ok {
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
		} else {
			_, ok1 := n.(node.BlockChildren)
			_, ok2 := n.(*node.Group)
			if ok1 && !(ok2 && n.Node() == "Paragraph") {
				panic(fmt.Sprintf("transformer: node %T does not implement SettableBlockChildren", n))
			}
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

		var nod node.Node
		if lines == nil {
			nod = grpNode
		} else {
			nod = &node.Hat{lines, grpNode}
		}

		nodes[pos] = nod
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
