package sequentialnumber

import (
	"strings"

	"github.com/touchmarine/to/aggregator"
	"github.com/touchmarine/to/node"
	"github.com/touchmarine/to/transformer/sequentialnumber"
)

type Aggregator struct {
	Elements []string
}

func (ar Aggregator) Aggregate(n *node.Node) aggregator.Aggregate {
	var ae aggregate
	walk(n, func(n *node.Node) bool {
		if ar.isTargetElement(n.Element) {
			if v, ok := n.Data[sequentialnumber.Key]; ok {
				seqnum := v.(string)
				ae = append(ae, particle{
					Element:          n.Element,
					ID:               n.TextContent(),
					Text:             n.TextContent(),
					SequentialNumber: seqnum,
				})
			}
		}
		return true
	})
	return aggregator.Aggregate(ae)
}

func (a Aggregator) isTargetElement(s string) bool {
	for _, e := range a.Elements {
		if e == s {
			return true
		}
	}
	return false
}

type aggregate []particle

func (aggregate) AnAggregate() {}

type particle struct {
	Element          string
	ID               string
	Text             string
	SequentialNumber string
}

func (p particle) depth() int {
	return len(strings.Split(p.SequentialNumber, "."))
}

func walk(n *node.Node, fn func(n *node.Node) bool) {
	if fn(n) {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c, fn)
		}
	}
}
