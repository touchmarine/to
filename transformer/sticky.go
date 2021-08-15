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
		switch m := g.node.(type) {
		case node.Boxed:
			g.unbox()

			// go back to process the unboxed node
			g.pos--

		case node.Block:
			peek := g.peek()

			if _, isBlock := peek.(node.Block); isBlock {
				sticky, isSticky := g.stickyByElement(g.node.Node())
				stickyPeek, isPeekSticky := g.stickyByElement(peek.Node())

				if isSticky && !sticky.After && !isPeekSticky {
					// sticky before

					g.sticky(sticky.Name, false)

					// go back so a possible before and
					// after sticky can be detected
					g.pos--
				} else if !isSticky && isPeekSticky && stickyPeek.After {
					// sticky after

					g.sticky(stickyPeek.Name, true)
				}
			}

		case node.SettableBlockChildren:
			stickied := GroupStickies(g.stickies, node.BlocksToNodes(m.BlockChildren()))
			m.SetBlockChildren(node.NodesToBlocks(stickied))

		case node.BlockChildren:
			if _, isGroup := g.node.(*node.Group); !(isGroup && g.node.Node() == "Paragraph") {
				panic(fmt.Sprintf("transformer: node %T does not implement SettableBlockChildren", g.node))
			}
		}
	}
}

// sticky groups the current and the following node into a Sticky node.
func (g *stickyGrouper) sticky(name string, after bool) {
	children := g.nodes[g.pos : g.pos+2]
	n := &node.Sticky{name, after, node.NodesToBlocks(children)}

	// set sticky to current node and remove the next node
	g.node = n
	g.nodes[g.pos] = n // insert sticky
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
