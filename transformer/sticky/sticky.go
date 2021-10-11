package sticky

import (
	"log"

	"github.com/touchmarine/to/node"
)

const trace = false

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
func (t Transformer) Transform(n *node.Node) *node.Node {
	targets := t.search(n)
	if trace {
		log.Printf("targets = %+v\n", targets)
	}
	for _, target := range targets {
		sticky := &node.Node{
			Element: target.name,
			Type:    node.TypeContainer,
			Data:    target.position,
		}

		target.child1.Parent.InsertBefore(sticky, target.child1)

		target.child1.Parent.RemoveChild(target.child1)
		sticky.AppendChild(target.child1)
		target.child2.Parent.RemoveChild(target.child2)
		sticky.AppendChild(target.child2)
	}

	return n
}

type target struct {
	name           string // sticky name
	position       string // "before" | "after"
	child1, child2 *node.Node
}

// search walks breadth-first and searches for a before or after sticky.
func (t Transformer) search(n *node.Node) []target {
	var targets []target

	for s := n; s != nil; s = s.NextSibling {
		if trace {
			log.Printf("s = %+v\n", s)
		}
		if s.TypeCategory() == node.CategoryBlock && s.NextSibling != nil && s.TypeCategory() == node.CategoryBlock {
			thisSticky, isThisSticky := t.StickyMap[s.Element]
			nextSticky, isNextSticky := t.StickyMap[s.NextSibling.Element]

			var a target
			if isThisSticky && !thisSticky.After && !isNextSticky {
				// sticky before
				a.name = thisSticky.Name
				a.position = "before"
			} else if !isThisSticky && isNextSticky && nextSticky.After {
				// sticky after
				a.name = nextSticky.Name
				a.position = "after"
			}

			if a.name != "" {
				// found sticky
				a.child1 = s
				a.child2 = s.NextSibling

				targets = append(targets, a)
			}
		}
	}

	// walk children
	for s := n; s != nil; s = s.NextSibling {
		if s.FirstChild != nil {
			newTargets := t.search(s.FirstChild)
			targets = append(targets, newTargets...)
		}
	}

	return targets
}
