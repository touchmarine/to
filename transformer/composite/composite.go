package composite

import (
	"log"

	"github.com/touchmarine/to/node"
)

const trace = false

// Map maps Composites to PrimaryElements.
type Map map[string]Composite

type Composite struct {
	Name             string
	PrimaryElement   string
	SecondaryElement string
}

type Transformer struct {
	CompositeMap Map
}

// Transform recognizes composite patterns and creates Composite nodes. Note
// that it mutates the given nodes.
//
// Transform recognizes only one form of patterns: PrimaryElement followed
// immediately by the SecondaryElement.
func (t Transformer) Transform(n *node.Node) *node.Node {
	targets := t.search(n)
	if trace {
		log.Printf("targets = %+v\n", targets)
	}
	for _, target := range targets {
		composite := &node.Node{
			Element: target.name,
			Type:    node.TypeContainer,
		}

		target.primary.Parent.InsertBefore(composite, target.primary)

		target.primary.Parent.RemoveChild(target.primary)
		composite.AppendChild(target.primary)
		target.secondary.Parent.RemoveChild(target.secondary)
		composite.AppendChild(target.secondary)
	}

	return n
}

type target struct {
	name               string     // composite name
	primary, secondary *node.Node // primary and secondary element
}

// search walks breadth-first and searches for a primary element followed
// immediately by the secondary element
func (t Transformer) search(n *node.Node) []target {
	var targets []target

	for s := n; s != nil; s = s.NextSibling {
		if trace {
			log.Printf("s = %+v\n", s)
		}
		if s.TypeCategory() == node.CategoryInline {
			if s.NextSibling != nil && s.TypeCategory() == node.CategoryInline {
				comp, ok := t.CompositeMap[s.Element]
				if ok && s.NextSibling.Element == comp.SecondaryElement {
					if trace {
						log.Printf("comp.Name = %+v\n", comp.Name)
					}

					targets = append(targets, target{
						name:      comp.Name,
						primary:   s,
						secondary: s.NextSibling,
					})

					// skip the secondary element
					s = s.NextSibling
				}
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
