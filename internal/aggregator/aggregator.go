package aggregator

import (
	"fmt"
	"strings"
	"to/internal/config"
	"to/internal/node"
)

type Item struct {
	Element string
	ID      string
	Text    string
	SeqNums []uint
	SeqNum  string
}

type aggregator struct {
	aggs []config.Aggregate
	m    map[string][]Item
}

func Aggregate(aggregates []config.Aggregate, nodes []node.Node) map[string][]Item {
	a := aggregator{
		aggs: aggregates,
		m:    make(map[string][]Item),
	}
	return a.aggregate(nodes)
}

func (a *aggregator) aggregate(nodes []node.Node) map[string][]Item {
	for _, n := range nodes {
		var item Item

		if m, ok := n.(node.Boxed); ok {
			switch k := n.(type) {
			case *node.SeqNumBox:
				item.SeqNums = k.SeqNums
				item.SeqNum = k.SeqNum()
			default:
				panic(fmt.Sprintf("aggregate: unexpected Boxed node %T", n))
			}

			n = m.Unbox()
		}

		name := n.Node()

		if a.isAggregate(name) {
			item.Element = name
			txt := a.text(n)
			item.ID = txt
			item.Text = txt

			for _, ag := range a.aggs {
				a.m[ag.Name] = append(a.m[ag.Name], item)
			}
		}
	}

	return a.m
}

func (a *aggregator) text(n node.Node) string {
	var b strings.Builder

	switch n.(type) {
	case node.BlockChildren, node.InlineChildren, node.Content:
	default:
		panic(fmt.Sprintf("text: unexpected node type %T", n))
	}

	if m, ok := n.(node.BlockChildren); ok {
		for i, c := range m.BlockChildren() {
			if i > 0 {
				b.WriteString("\n")
			}
			b.WriteString(a.text(c))
		}
	}

	if m, ok := n.(node.InlineChildren); ok {
		for _, c := range m.InlineChildren() {
			b.WriteString(a.text(c))
		}
	}

	if m, ok := n.(node.Content); ok {
		b.Write(m.Content())
	}

	return b.String()
}

func (a *aggregator) isAggregate(name string) bool {
	for _, ag := range a.aggs {
		for _, el := range ag.Elements {
			if el == name {
				return true
			}
		}
	}
	return false
}
