package sequentialnumber_test

import (
	"encoding/json"
	"testing"

	"github.com/touchmarine/to/aggregator/sequentialnumber"
	"github.com/touchmarine/to/node"
	"github.com/touchmarine/to/parser"
	sequentialnumbertr "github.com/touchmarine/to/transformer/sequentialnumber"
)

func TestAggregate(t *testing.T) {
	cases := []struct {
		name string
		in   *node.Node
		out  *sequentialnumber.Aggregate
	}{
		{
			"no sequential number",
			&node.Node{Element: "A", Type: node.TypeRankedHanging, Data: node.Data{
				parser.KeyRank: 2,
			}},
			nil,
		},
		{
			"1 sequential number",
			&node.Node{Element: "A", Type: node.TypeRankedHanging, Data: node.Data{
				parser.KeyRank:         2,
				sequentialnumbertr.Key: "1",
			}},
			&sequentialnumber.Aggregate{
				{
					Element:          "A",
					SequentialNumber: "1",
				},
			},
		},
		{
			"1 sequential number with text",
			appendChildren(
				&node.Node{Element: "A", Type: node.TypeRankedHanging, Data: node.Data{
					parser.KeyRank:         2,
					sequentialnumbertr.Key: "1",
				}},
				[]*node.Node{
					appendChildren(
						&node.Node{Element: "T", Type: node.TypeLeaf},
						[]*node.Node{
							&node.Node{Element: "MT", Type: node.TypeText, Value: "a"},
						},
					),
				},
			),
			&sequentialnumber.Aggregate{
				{
					Element:          "A",
					ID:               "a",
					Text:             "a",
					SequentialNumber: "1",
				},
			},
		},
		{
			"1 sequential number with multiple inlines",
			appendChildren(
				&node.Node{Element: "A", Type: node.TypeRankedHanging, Data: node.Data{
					parser.KeyRank:         2,
					sequentialnumbertr.Key: "1",
				}},
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
			&sequentialnumber.Aggregate{
				{
					Element:          "A",
					ID:               "a b c",
					Text:             "a b c",
					SequentialNumber: "1",
				},
			},
		},
		{
			"1 sequential number with multiple blocks",
			appendChildren(
				&node.Node{Element: "A", Type: node.TypeRankedHanging, Data: node.Data{
					parser.KeyRank:         2,
					sequentialnumbertr.Key: "1",
				}},
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
			&sequentialnumber.Aggregate{
				{
					Element:          "A",
					ID:               "a\nb",
					Text:             "a\nb",
					SequentialNumber: "1",
				},
			},
		},
		{
			"2 sequences",
			appendChildren(
				&node.Node{Type: node.TypeContainer},
				[]*node.Node{
					&node.Node{Element: "A", Type: node.TypeRankedHanging, Data: node.Data{
						parser.KeyRank:         2,
						sequentialnumbertr.Key: "1",
					}},
					&node.Node{Element: "A", Type: node.TypeRankedHanging, Data: node.Data{
						parser.KeyRank:         3,
						sequentialnumbertr.Key: "1.1",
					}},
				},
			),
			&sequentialnumber.Aggregate{
				{
					Element:          "A",
					SequentialNumber: "1",
				},
				{
					Element:          "A",
					SequentialNumber: "1.1",
				},
			},
		},
		{
			"nested",
			appendChildren(
				&node.Node{Type: node.TypeContainer},
				[]*node.Node{
					&node.Node{Element: "A", Type: node.TypeRankedHanging, Data: node.Data{
						parser.KeyRank:         2,
						sequentialnumbertr.Key: "1",
					}},
				},
			),
			&sequentialnumber.Aggregate{
				{
					Element:          "A",
					SequentialNumber: "1",
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			aggregate := sequentialnumber.Aggregator{[]string{"A"}}.Aggregate(c.in)
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
