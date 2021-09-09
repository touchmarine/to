package seqnum_test

import (
	"encoding/json"
	"github.com/touchmarine/to/aggregator/seqnum"
	"github.com/touchmarine/to/node"
	"testing"
)

func TestAggregate(t *testing.T) {
	cases := []struct {
		name string
		in   []node.Node
		out  seqnum.Aggregate
	}{
		{
			"1 heading",
			[]node.Node{
				&node.SequentialNumberBox{
					&node.RankedHanging{"A", 2, nil},
					[]int{1},
				},
			},
			seqnum.Aggregate{
				{
					Element:           "A",
					SequentialNumbers: []int{1},
					SequentialNumber:  "1",
				},
			},
		},
		{
			"no seqbox",
			[]node.Node{
				&node.RankedHanging{"A", 2, nil},
			},
			seqnum.Aggregate{
				{
					Element: "A",
				},
			},
		},

		{
			"1 heading with text",
			[]node.Node{
				&node.SequentialNumberBox{
					&node.RankedHanging{"A", 2, []node.Block{
						&node.Leaf{"D", []node.Inline{
							node.Text("a"),
						}},
					}},
					[]int{1},
				},
			},
			seqnum.Aggregate{
				{
					Element:           "A",
					ID:                "a",
					Text:              "a",
					SequentialNumbers: []int{1},
					SequentialNumber:  "1",
				},
			},
		},
		{
			"1 heading with multiple inlines",
			[]node.Node{
				&node.SequentialNumberBox{
					&node.RankedHanging{"A", 2, []node.Block{
						&node.Leaf{"D", []node.Inline{
							node.Text("a"),
							&node.Uniform{"Emphasis", []node.Inline{
								node.Text("b"),
							}},
							node.Text("c"),
						}},
					}},
					[]int{1},
				},
			},
			seqnum.Aggregate{
				{
					Element:           "A",
					ID:                "abc",
					Text:              "abc",
					SequentialNumbers: []int{1},
					SequentialNumber:  "1",
				},
			},
		},
		{
			"1 heading with multiple blocks",
			[]node.Node{
				&node.SequentialNumberBox{
					&node.RankedHanging{"A", 2, []node.Block{
						&node.Leaf{"D", []node.Inline{
							node.Text("a"),
						}},
						&node.Leaf{"D", []node.Inline{
							node.Text("b"),
						}},
					}},
					[]int{1},
				},
			},
			seqnum.Aggregate{
				{
					Element:           "A",
					ID:                "a\nb",
					Text:              "a\nb",
					SequentialNumbers: []int{1},
					SequentialNumber:  "1",
				},
			},
		},

		{
			"2 headings",
			[]node.Node{
				&node.SequentialNumberBox{
					&node.RankedHanging{"A", 2, nil},
					[]int{1},
				},
				&node.SequentialNumberBox{
					&node.RankedHanging{"A", 3, nil},
					[]int{1, 1},
				},
			},
			seqnum.Aggregate{
				{
					Element:           "A",
					SequentialNumbers: []int{1},
					SequentialNumber:  "1",
				},
				{
					Element:           "A",
					SequentialNumbers: []int{1, 1},
					SequentialNumber:  "1.1",
				},
			},
		},

		{
			"in sticky",
			[]node.Node{
				&node.Sticky{"SA", false, []node.Block{
					&node.VerbatimWalled{"B", [][]byte{[]byte("a")}},
					&node.SequentialNumberBox{
						&node.RankedHanging{"A", 2, nil},
						[]int{1},
					},
				}},
			},
			seqnum.Aggregate{
				{
					Element:           "A",
					SequentialNumbers: []int{1},
					SequentialNumber:  "1",
				},
			},
		},
		{
			"in double sticky",
			[]node.Node{
				&node.Sticky{"SB", true, []node.Block{
					&node.Sticky{"SA", false, []node.Block{
						&node.VerbatimWalled{"B", [][]byte{[]byte("a")}},
						&node.SequentialNumberBox{
							&node.RankedHanging{"A", 2, nil},
							[]int{1},
						},
					}},
					&node.VerbatimWalled{"C", [][]byte{[]byte("c")}},
				}},
			},
			seqnum.Aggregate{
				{
					Element:           "A",
					SequentialNumbers: []int{1},
					SequentialNumber:  "1",
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			a := make([]node.Node, len(c.in))
			copy(a, c.in)
			sequencer := seqnum.Aggregator{[]string{"A"}}
			b := sequencer.Aggregate(a)

			got := jsonMarshal(t, b)
			want := jsonMarshal(t, c.out)

			if got != want {
				t.Errorf("\ngot\n%s\nwant\n%s", got, want)
			}
		})
	}
}

func jsonMarshal(t *testing.T, v interface{}) string {
	t.Helper()

	json, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		t.Fatal(err)
	}

	return string(json)
}
