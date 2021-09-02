package aggregator

import (
	"fmt"
	"github.com/touchmarine/to/config"
	"github.com/touchmarine/to/node"
)

type Item struct {
	Element string
	ID      string
	Text    string
	SeqNums []int
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

		switch m := n.(type) {
		case *node.Sticky:
			a.aggregate(node.BlocksToNodes(m.BlockChildren()))
		case node.Boxed:
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
			txt := node.ExtractText(n)
			item.ID = txt
			item.Text = txt

			for _, ag := range a.aggs {
				a.m[ag.Name] = append(a.m[ag.Name], item)
			}
		}
	}

	return a.m
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
