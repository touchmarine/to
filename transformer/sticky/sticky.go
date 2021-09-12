package sticky

import (
	"fmt"
	"github.com/touchmarine/to/node"
)

// Map maps Stickies to Elements.
type Map map[string]Sticky

type Sticky struct {
	Name    string `json:"name"`
	Element string `json:"element"`
	After   bool   `json:"after"`
}

type Transformer struct {
	StickyMap Map
}

// Transform recognizes sticky patterns and creates Sticky nodes. Note that
// it mutates the given nodes.
//
// Transform recognizes only immediately placed sticky elements. Out of
// multiple consecutive sticky elements, only the one closest to a non-sticky
// element is grouped. One element can have one sticky element before it and one
// after it.
func (t Transformer) Transform(nodes []node.Node) []node.Node {
	s := stickyGrouper{
		transformer: &t,
		stickyMap:   t.StickyMap,
		nodes:       nodes,
		pos:         -1,
	}

	s.groupStickies()
	return s.nodes
}

type stickyGrouper struct {
	transformer *Transformer
	stickyMap   Map
	nodes       []node.Node

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
				sticky, isSticky := g.stickyMap[g.node.Node()]
				stickyPeek, isPeekSticky := g.stickyMap[peek.Node()]

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
		stickied := g.transformer.Transform(node.BlocksToNodes(m.BlockChildren()))
		m.SetBlockChildren(node.NodesToBlocks(stickied))
	} else {
		_, isBlockChildren := n.(node.BlockChildren)
		_, isGroup := n.(*node.Group)
		if isBlockChildren && !isGroup {
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
