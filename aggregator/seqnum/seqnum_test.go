package seqnum_test

import (
	"encoding/json"
	"testing"

	"github.com/touchmarine/to/aggregator/seqnum"
	"github.com/touchmarine/to/node"
	"github.com/touchmarine/to/transformer/sequence"
)

func TestAggregate(t *testing.T) {
	cases := []struct {
		name string
		in   *node.Node
		out  *seqnum.Aggregate
	}{
		{
			"no sequence",
			&node.Node{Element: "A", Type: node.TypeRankedHanging, Data: 2},
			nil,
		},
		{
			"1 sequence",
			appendChildren(
				&node.Node{Element: sequence.Element, Type: node.TypeContainer, Data: []int{1}},
				[]*node.Node{
					&node.Node{Element: "A", Type: node.TypeRankedHanging, Data: 2},
				},
			),
			&seqnum.Aggregate{
				{
					Element:           "A",
					SequentialNumbers: []int{1},
				},
			},
		},
		{
			"1 sequence with text",
			appendChildren(
				&node.Node{Element: sequence.Element, Type: node.TypeContainer, Data: []int{1}},
				[]*node.Node{
					appendChildren(
						&node.Node{Element: "A", Type: node.TypeRankedHanging, Data: 2},
						[]*node.Node{
							appendChildren(
								&node.Node{Element: "T", Type: node.TypeLeaf},
								[]*node.Node{
									&node.Node{Element: "MT", Type: node.TypeText, Value: "a"},
								},
							),
						},
					),
				},
			),
			&seqnum.Aggregate{
				{
					Element:           "A",
					ID:                "a",
					Text:              "a",
					SequentialNumbers: []int{1},
				},
			},
		},
		{
			"1 sequence with multiple inlines",
			appendChildren(
				&node.Node{Element: sequence.Element, Type: node.TypeContainer, Data: []int{1}},
				[]*node.Node{
					appendChildren(
						&node.Node{Element: "A", Type: node.TypeRankedHanging, Data: 2},
						[]*node.Node{
							appendChildren(
								&node.Node{Element: "T", Type: node.TypeLeaf},
								[]*node.Node{
									&node.Node{Element: "MT", Type: node.TypeText, Value: "a "},
									&node.Node{Element: "ME", Type: node.TypeUniform, Value: "b"},
									&node.Node{Element: "MT", Type: node.TypeText, Value: " c"},
								},
							),
						},
					),
				},
			),
			&seqnum.Aggregate{
				{
					Element:           "A",
					ID:                "a b c",
					Text:              "a b c",
					SequentialNumbers: []int{1},
				},
			},
		},
		{
			"1 sequence with multiple blocks",
			appendChildren(
				&node.Node{Element: sequence.Element, Type: node.TypeContainer, Data: []int{1}},
				[]*node.Node{
					appendChildren(
						&node.Node{Element: "A", Type: node.TypeRankedHanging, Data: 2},
						[]*node.Node{
							appendChildren(
								&node.Node{Element: "T", Type: node.TypeLeaf},
								[]*node.Node{
									&node.Node{Element: "MT", Type: node.TypeText, Value: "a"},
								},
							),
							appendChildren(
								&node.Node{Element: "T", Type: node.TypeLeaf},
								[]*node.Node{
									&node.Node{Element: "MT", Type: node.TypeText, Value: "b"},
								},
							),
						},
					),
				},
			),
			&seqnum.Aggregate{
				{
					Element:           "A",
					ID:                "a\nb",
					Text:              "a\nb",
					SequentialNumbers: []int{1},
				},
			},
		},
		{
			"2 sequences",
			appendChildren(
				&node.Node{Type: node.TypeContainer},
				[]*node.Node{
					appendChildren(
						&node.Node{Element: sequence.Element, Type: node.TypeContainer, Data: []int{1}},
						[]*node.Node{
							&node.Node{Element: "A", Type: node.TypeRankedHanging, Data: 2},
						},
					),
					appendChildren(
						&node.Node{Element: sequence.Element, Type: node.TypeContainer, Data: []int{1, 1}},
						[]*node.Node{
							&node.Node{Element: "A", Type: node.TypeRankedHanging, Data: 3},
						},
					),
				},
			),
			&seqnum.Aggregate{
				{
					Element:           "A",
					SequentialNumbers: []int{1},
				},
				{
					Element:           "A",
					SequentialNumbers: []int{1, 1},
				},
			},
		},
		{
			"nested",
			appendChildren(
				&node.Node{Type: node.TypeContainer},
				[]*node.Node{
					appendChildren(
						&node.Node{Element: sequence.Element, Type: node.TypeContainer, Data: []int{1}},
						[]*node.Node{
							&node.Node{Element: "A", Type: node.TypeRankedHanging, Data: 2},
						},
					),
				},
			),
			&seqnum.Aggregate{
				{
					Element:           "A",
					SequentialNumbers: []int{1},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			aggregate := seqnum.Aggregator{[]string{"A"}}.Aggregate(c.in)
			got := jsonMarshal(t, aggregate)
			want := jsonMarshal(t, c.out)
			if got != want {
				t.Errorf("\ngot\n%s\nwant\n%s", got, want)
			}
		})
	}
}

func appendChildren(n *node.Node, children []*node.Node) *node.Node {
	for _, child := range children {
		n.AppendChild(child)
	}
	return n
}

func jsonMarshal(t *testing.T, v interface{}) string {
	t.Helper()

	json, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		t.Fatal(err)
	}

	return string(json)
}
