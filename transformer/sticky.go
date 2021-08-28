package transformer

import (
	"fmt"
	"github.com/touchmarine/to/config"
	"github.com/touchmarine/to/node"
)

// GroupStickies recognizes sticky patterns and creates Sticky nodes. Note that
// it mutates the given nodes.
//
// GroupStickies recognizes only immediately placed sticky elements. Out of
// multiple consecutive sticky elements, only the one closest to a non-sticky
// element is grouped. One element can have one sticky element before it and one
// after it.
func GroupStickies(stickies []config.Sticky, nodes []node.Node) []node.Node {
	s := stickyGrouper{
		stickies: stickies,
		nodes:    nodes,
		pos:      -1,
	}

	s.groupStickies()
	return s.nodes
}

type stickyGrouper struct {
	stickies []config.Sticky
	nodes    []node.Node

	node node.Node
	pos  int
}

func (g *stickyGrouper) next() bool {
	if g.pos+1 < len(g.nodes) {
		g.pos++
		g.node = g.nodes[g.pos]
		return true
	}

	return false
}

func (g *stickyGrouper) peek() node.Node {
	if g.pos+1 < len(g.nodes) {
		return g.nodes[g.pos+1]
	}

	return nil
}

func (g *stickyGrouper) groupStickies() {
	for g.next() {
		g.groupChildren(g.node)

	group:
		switch g.node.(type) {
		case node.Boxed:
			g.unbox()
			goto group

		case node.Block:
			peek := g.peek()

			if _, isBlock := peek.(node.Block); isBlock {
				sticky, isSticky := g.stickyByElement(g.node.Node())
				stickyPeek, isPeekSticky := g.stickyByElement(peek.Node())

				var stickyNode *node.Sticky
				if isSticky && !sticky.After && !isPeekSticky {
					// before
					stickyNode = g.createSticky(sticky.Name, false)
				} else if !isSticky && isPeekSticky && stickyPeek.After {
					// after
					stickyNode = g.createSticky(stickyPeek.Name, true)
				}

				if stickyNode != nil {
					// group peek children here as we move
					// peek into sticky right after and get
					// caught in an infinite loop otherwise
					// as it would group the same sticky
					// again and again
					g.groupChildren(peek)

					g.setNode(stickyNode)
					g.removePeek()

					if !stickyNode.After {
						// check for a sticky after the
						// new before sticky
						goto group
					}
				}
			}
		}
	}
}

func (g *stickyGrouper) groupChildren(n node.Node) {
	if m, ok := n.(node.SettableBlockChildren); ok {
		stickied := GroupStickies(g.stickies, node.BlocksToNodes(m.BlockChildren()))
		m.SetBlockChildren(node.NodesToBlocks(stickied))
	} else {
		_, isBlockChildren := n.(node.BlockChildren)
		_, isGroup := n.(*node.Group)
		if isBlockChildren && !(isGroup && n.Node() == "Paragraph") {
			panic(fmt.Sprintf("transformer: node %T does not implement SettableBlockChildren", n))
		}
	}
}

func (g *stickyGrouper) createSticky(name string, after bool) *node.Sticky {
	children := g.nodes[g.pos : g.pos+2]
	return &node.Sticky{name, after, node.NodesToBlocks(children)}
}

func (g *stickyGrouper) setNode(n node.Node) {
	g.node = n
	g.nodes[g.pos] = n
}

func (g *stickyGrouper) removePeek() {
	g.nodes = append(g.nodes[:g.pos+1], g.nodes[g.pos+2:]...)
}

// unbox unboxes the current node and replaces it with the unboxed node.
func (g *stickyGrouper) unbox() {
	boxed, ok := g.node.(node.Boxed)
	if !ok {
		panic("transformer: unboxing node that does not implement Boxed")
	}

	unboxed := boxed.Unbox()

	g.node = unboxed
	g.nodes[g.pos] = unboxed
}

func (g *stickyGrouper) stickyByElement(element string) (config.Sticky, bool) {
	for _, sticky := range g.stickies {
		if sticky.Element == element {
			return sticky, true
		}
	}
	return config.Sticky{}, false
}
