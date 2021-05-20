package aggregator_test

import (
	"encoding/json"
	"github.com/touchmarine/to/internal/aggregator"
	"github.com/touchmarine/to/internal/config"
	"github.com/touchmarine/to/internal/node"
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
					&node.Hanging{"NumberedHeading", 2, nil},
					[]uint{1},
				},
			},
			map[string][]aggregator.Item{
				"Headings": []aggregator.Item{
					{
						Element: "NumberedHeading",
						SeqNums: []uint{1},
						SeqNum:  "1",
					},
				},
			},
		},
		{
			"no seqbox",
			[]node.Node{
				&node.Hanging{"NumberedHeading", 2, nil},
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
					&node.Hanging{"NumberedHeading", 2, []node.Block{
						&node.Line{"Line", []node.Inline{
							node.Text("a"),
						}},
					}},
					[]uint{1},
				},
			},
			map[string][]aggregator.Item{
				"Headings": []aggregator.Item{
					{
						Element: "NumberedHeading",
						ID:      "a",
						Text:    "a",
						SeqNums: []uint{1},
						SeqNum:  "1",
					},
				},
			},
		},
		{
			"1 heading with multiple inlines",
			[]node.Node{
				&node.SeqNumBox{
					&node.Hanging{"NumberedHeading", 2, []node.Block{
						&node.Line{"Line", []node.Inline{
							node.Text("a"),
							&node.Uniform{"Emphasis", []node.Inline{
								node.Text("b"),
							}},
							node.Text("c"),
						}},
					}},
					[]uint{1},
				},
			},
			map[string][]aggregator.Item{
				"Headings": []aggregator.Item{
					{
						Element: "NumberedHeading",
						ID:      "abc",
						Text:    "abc",
						SeqNums: []uint{1},
						SeqNum:  "1",
					},
				},
			},
		},
		{
			"1 heading with multiple blocks",
			[]node.Node{
				&node.SeqNumBox{
					&node.Hanging{"NumberedHeading", 2, []node.Block{
						&node.Line{"Line", []node.Inline{
							node.Text("a"),
						}},
						&node.Line{"Line", []node.Inline{
							node.Text("b"),
						}},
					}},
					[]uint{1},
				},
			},
			map[string][]aggregator.Item{
				"Headings": []aggregator.Item{
					{
						Element: "NumberedHeading",
						ID:      "a\nb",
						Text:    "a\nb",
						SeqNums: []uint{1},
						SeqNum:  "1",
					},
				},
			},
		},

		{
			"2 headings",
			[]node.Node{
				&node.SeqNumBox{
					&node.Hanging{"NumberedHeading", 2, nil},
					[]uint{1},
				},
				&node.SeqNumBox{
					&node.Hanging{"NumberedHeading", 3, nil},
					[]uint{1, 1},
				},
			},
			map[string][]aggregator.Item{
				"Headings": []aggregator.Item{
					{
						Element: "NumberedHeading",
						SeqNums: []uint{1},
						SeqNum:  "1",
					},
					{
						Element: "NumberedHeading",
						SeqNums: []uint{1, 1},
						SeqNum:  "1.1",
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
