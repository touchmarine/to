package transformer_test

import (
	"github.com/touchmarine/to/node"
	"github.com/touchmarine/to/stringifier"
	"github.com/touchmarine/to/transformer"
	"testing"
)

func TestSequence(t *testing.T) {
	cases := []struct {
		name string
		in   []node.Node
		out  []node.Node
	}{
		{
			"##",
			[]node.Node{
				&node.RankedHanging{"NumberedHeading", 2, nil},
			},
			[]node.Node{
				&node.SeqNumBox{
					&node.RankedHanging{"NumberedHeading", 2, nil},
					[]int{1},
				},
			},
		},

		{
			"##\n##",
			[]node.Node{
				&node.RankedHanging{"NumberedHeading", 2, nil},
				&node.RankedHanging{"NumberedHeading", 2, nil},
			},
			[]node.Node{
				&node.SeqNumBox{
					&node.RankedHanging{"NumberedHeading", 2, nil},
					[]int{1},
				},
				&node.SeqNumBox{
					&node.RankedHanging{"NumberedHeading", 2, nil},
					[]int{2},
				},
			},
		},
		{
			"##\n###",
			[]node.Node{
				&node.RankedHanging{"NumberedHeading", 2, nil},
				&node.RankedHanging{"NumberedHeading", 3, nil},
			},
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
		},
		{
			"###",
			[]node.Node{
				&node.RankedHanging{"NumberedHeading", 3, nil},
			},
			[]node.Node{
				&node.SeqNumBox{
					&node.RankedHanging{"NumberedHeading", 3, nil},
					[]int{0, 1},
				},
			},
		},

		{
			"reset ##\n###\n##\n###",
			[]node.Node{
				&node.RankedHanging{"NumberedHeading", 2, nil},
				&node.RankedHanging{"NumberedHeading", 3, nil},
				&node.RankedHanging{"NumberedHeading", 2, nil},
				&node.RankedHanging{"NumberedHeading", 3, nil},
			},
			[]node.Node{
				&node.SeqNumBox{
					&node.RankedHanging{"NumberedHeading", 2, nil},
					[]int{1},
				},
				&node.SeqNumBox{
					&node.RankedHanging{"NumberedHeading", 3, nil},
					[]int{1, 1},
				},
				&node.SeqNumBox{
					&node.RankedHanging{"NumberedHeading", 2, nil},
					[]int{2},
				},
				&node.SeqNumBox{
					&node.RankedHanging{"NumberedHeading", 3, nil},
					[]int{2, 1},
				},
			},
		},

		{
			"sticky",
			[]node.Node{
				&node.Sticky{"SA", false, []node.Block{
					&node.VerbatimWalled{"A", [][]byte{[]byte("a")}},
					&node.RankedHanging{"NumberedHeading", 2, nil},
				}},
			},
			[]node.Node{
				&node.Sticky{"SA", false, []node.Block{
					&node.VerbatimWalled{"A", [][]byte{[]byte("a")}},
					&node.SeqNumBox{
						&node.RankedHanging{"NumberedHeading", 2, nil},
						[]int{1},
					},
				}},
			},
		},
		{
			"double sticky",
			[]node.Node{
				&node.Sticky{"SB", true, []node.Block{
					&node.Sticky{"SA", false, []node.Block{
						&node.VerbatimWalled{"A", [][]byte{[]byte("a")}},
						&node.RankedHanging{"NumberedHeading", 2, nil},
					}},
					&node.VerbatimWalled{"B", [][]byte{[]byte("c")}},
				}},
			},
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
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			a := make([]node.Node, len(c.in))
			copy(a, c.in)
			a = transformer.Sequence(a)

			got := stringifier.Stringify(a...)
			want := stringifier.Stringify(c.out...)

			if got != want {
				t.Errorf("\ngot\n%s\nwant\n%s", got, want)
			}
		})
	}
}
