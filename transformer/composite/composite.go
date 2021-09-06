package composite

import (
	"fmt"
	"github.com/touchmarine/to/config"
	"github.com/touchmarine/to/node"
)

type Transformer struct {
	Composites []config.Composite
}

// Transform recognizes composite patterns and creates Composite nodes. Note
// that it mutates the given nodes.
//
// Transform recognizes only one form of patterns: PrimaryElement followed
// immediately by the SecondaryElement.
func (t Transformer) Transform(nodes []node.Node) []node.Node {
	c := compositer{
		transformer: &t,
		composites:  t.Composites,
		nodes:       nodes,
		pos:         -1,
	}

	c.composite()
	return c.nodes
}

type compositer struct {
	transformer *Transformer
	composites  []config.Composite
	nodes       []node.Node

	node node.Node
	pos  int
}

func (c *compositer) next() bool {
	if c.pos+1 < len(c.nodes) {
		c.pos++
		c.node = c.nodes[c.pos]
		return true
	}

	return false
}

func (c *compositer) peek() node.Node {
	if c.pos+1 < len(c.nodes) {
		return c.nodes[c.pos+1]
	}

	return nil
}

func (c *compositer) composite() {
	for c.next() {
	beginning:
		switch m := c.node.(type) {
		case node.Boxed:
			c.unbox()

			if c.node == nil {
				continue
			}

			goto beginning

		case node.Inline:
			peek := c.peek()

			if peek != nil {
				inlinePeek, isInline := peek.(node.Inline)
				if !isInline {
					panic("transformer: mixed node types, expected Inline")
				}

				comp, ok := c.compositeByPrimaryElement(c.node.Node())
				if ok && peek.Node() == comp.SecondaryElement {
					n := &node.Composite{comp.Name, m, inlinePeek}

					// replace current node by Composite and remove peek
					c.nodes[c.pos] = n
					c.nodes = append(c.nodes[:c.pos+1], c.nodes[c.pos+2:]...)
				}
			}

		case node.SettableInlineChildren:
			composited := c.transformer.Transform(node.InlinesToNodes(m.InlineChildren()))
			m.SetInlineChildren(node.NodesToInlines(composited))

		case node.InlineChildren:
			panic(fmt.Sprintf("transformer: node %T does not implement SettableInlineChildren", c.node))
		}

		if m, ok := c.node.(node.SettableBlockChildren); ok {
			composited := c.transformer.Transform(node.BlocksToNodes(m.BlockChildren()))
			m.SetBlockChildren(node.NodesToBlocks(composited))
		} else {
			_, isBlockChildren := c.node.(node.BlockChildren)
			_, isGroup := c.node.(*node.Group)
			if isBlockChildren && !(isGroup && c.node.Node() == "Paragraph") {
				panic(fmt.Sprintf("transformer: node %T does not implement SettableBlockChildren", c.node))
			}
		}
	}
}

func (c *compositer) unbox() {
	boxed, ok := c.node.(node.Boxed)
	if !ok {
		panic("transformer: unboxing node that does not implement Boxed")
	}

	c.node = boxed.Unbox()
}

func (c *compositer) compositeByPrimaryElement(element string) (config.Composite, bool) {
	for _, comp := range c.composites {
		if comp.PrimaryElement == element {
			return comp, true
		}
	}
	return config.Composite{}, false
}
