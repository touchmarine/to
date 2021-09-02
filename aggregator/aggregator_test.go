package aggregator_test

import (
	"encoding/json"
	"github.com/touchmarine/to/aggregator"
	"github.com/touchmarine/to/config"
	"github.com/touchmarine/to/node"
	"testing"
)

func TestAggregate(t *testing.T) {
	cases := []struct {
		name string
		in   []node.Node
		out  map[string][]aggregator.Item
	}{
		{
			"1 heading",
			[]node.Node{
				&node.SeqNumBox{
					&node.RankedHanging{"NumberedHeading", 2, nil},
					[]int{1},
				},
			},
			map[string][]aggregator.Item{
				"Headings": []aggregator.Item{
					{
						Element: "NumberedHeading",
						SeqNums: []int{1},
						SeqNum:  "1",
					},
				},
			},
		},
		{
			"no seqbox",
			[]node.Node{
				&node.RankedHanging{"NumberedHeading", 2, nil},
			},
			map[string][]aggregator.Item{
				"Headings": []aggregator.Item{
					{
						Element: "NumberedHeading",
					},
				},
			},
		},

		{
			"1 heading with text",
			[]node.Node{
				&node.SeqNumBox{
					&node.RankedHanging{"NumberedHeading", 2, []node.Block{
						&node.BasicBlock{"TextBlock", []node.Inline{
							node.Text("a"),
						}},
					}},
					[]int{1},
				},
			},
			map[string][]aggregator.Item{
				"Headings": []aggregator.Item{
					{
						Element: "NumberedHeading",
						ID:      "a",
						Text:    "a",
						SeqNums: []int{1},
						SeqNum:  "1",
					},
				},
			},
		},
		{
			"1 heading with multiple inlines",
			[]node.Node{
				&node.SeqNumBox{
					&node.RankedHanging{"NumberedHeading", 2, []node.Block{
						&node.BasicBlock{"TextBlock", []node.Inline{
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
			map[string][]aggregator.Item{
				"Headings": []aggregator.Item{
					{
						Element: "NumberedHeading",
						ID:      "abc",
						Text:    "abc",
						SeqNums: []int{1},
						SeqNum:  "1",
					},
				},
			},
		},
		{
			"1 heading with multiple blocks",
			[]node.Node{
				&node.SeqNumBox{
					&node.RankedHanging{"NumberedHeading", 2, []node.Block{
						&node.BasicBlock{"TextBlock", []node.Inline{
							node.Text("a"),
						}},
						&node.BasicBlock{"TextBlock", []node.Inline{
							node.Text("b"),
						}},
					}},
					[]int{1},
				},
			},
			map[string][]aggregator.Item{
				"Headings": []aggregator.Item{
					{
						Element: "NumberedHeading",
						ID:      "a\nb",
						Text:    "a\nb",
						SeqNums: []int{1},
						SeqNum:  "1",
					},
				},
			},
		},

		{
			"2 headings",
			[]node.Node{
				&node.SeqNumBox{
					&node.RankedHanging{"NumberedHeading", 2, nil},
					[]int{1},
				},
				&node.SeqNumBox{
					&node.RankedHanging{"NumberedHeading", 3, nil},
					[]int{1, 1},
				},
			},
			map[string][]aggregator.Item{
				"Headings": []aggregator.Item{
					{
						Element: "NumberedHeading",
						SeqNums: []int{1},
						SeqNum:  "1",
					},
					{
						Element: "NumberedHeading",
						SeqNums: []int{1, 1},
						SeqNum:  "1.1",
					},
				},
			},
		},

		{
			"in sticky",
			[]node.Node{
				&node.Sticky{"SA", false, []node.Block{
					&node.VerbatimWalled{"A", [][]byte{[]byte("a")}},
					&node.SeqNumBox{
						&node.RankedHanging{"NumberedHeading", 2, nil},
						[]int{1},
					},
				}},
			},
			map[string][]aggregator.Item{
				"Headings": []aggregator.Item{
					{
						Element: "NumberedHeading",
						SeqNums: []int{1},
						SeqNum:  "1",
					},
				},
			},
		},
		{
			"in double sticky",
			[]node.Node{
				&node.Sticky{"SB", true, []node.Block{
					&node.Sticky{"SA", false, []node.Block{
						&node.VerbatimWalled{"A", [][]byte{[]byte("a")}},
						&node.SeqNumBox{
							&node.RankedHanging{"NumberedHeading", 2, nil},
							[]int{1},
						},
					}},
					&node.VerbatimWalled{"B", [][]byte{[]byte("c")}},
				}},
			},
			map[string][]aggregator.Item{
				"Headings": []aggregator.Item{
					{
						Element: "NumberedHeading",
						SeqNums: []int{1},
						SeqNum:  "1",
					},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			a := make([]node.Node, len(c.in))
			copy(a, c.in)
			b := aggregator.Aggregate(config.Default.Aggregates, a)

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
