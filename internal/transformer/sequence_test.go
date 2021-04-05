package transformer_test

import (
	"testing"
	"to/internal/config"
	"to/internal/node"
	"to/internal/stringifier"
	"to/internal/transformer"
)

func TestSequence(t *testing.T) {
	cases := []struct {
		name string
		in   []node.Node
		out  []node.Node
	}{
		{
			"rank 0",
			[]node.Node{
				&node.Hanging{"NumberedHeading", 0, nil},
			},
			[]node.Node{
				&node.SeqNumBox{
					&node.Hanging{"NumberedHeading", 0, nil},
					[]uint{},
				},
			},
		},
		{
			"# (minRank 2)",
			[]node.Node{
				&node.Hanging{"NumberedHeading", 1, nil},
			},
			[]node.Node{
				&node.SeqNumBox{
					&node.Hanging{"NumberedHeading", 1, nil},
					[]uint{},
				},
			},
		},
		{
			"##",
			[]node.Node{
				&node.Hanging{"NumberedHeading", 2, nil},
			},
			[]node.Node{
				&node.SeqNumBox{
					&node.Hanging{"NumberedHeading", 2, nil},
					[]uint{1},
				},
			},
		},

		{
			"##\n##",
			[]node.Node{
				&node.Hanging{"NumberedHeading", 2, nil},
				&node.Hanging{"NumberedHeading", 2, nil},
			},
			[]node.Node{
				&node.SeqNumBox{
					&node.Hanging{"NumberedHeading", 2, nil},
					[]uint{1},
				},
				&node.SeqNumBox{
					&node.Hanging{"NumberedHeading", 2, nil},
					[]uint{2},
				},
			},
		},
		{
			"##\n###",
			[]node.Node{
				&node.Hanging{"NumberedHeading", 2, nil},
				&node.Hanging{"NumberedHeading", 3, nil},
			},
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
		},
		{
			"###",
			[]node.Node{
				&node.Hanging{"NumberedHeading", 3, nil},
			},
			[]node.Node{
				&node.SeqNumBox{
					&node.Hanging{"NumberedHeading", 3, nil},
					[]uint{0, 1},
				},
			},
		},

		{
			"reset ##\n###\n##\n###",
			[]node.Node{
				&node.Hanging{"NumberedHeading", 2, nil},
				&node.Hanging{"NumberedHeading", 3, nil},
				&node.Hanging{"NumberedHeading", 2, nil},
				&node.Hanging{"NumberedHeading", 3, nil},
			},
			[]node.Node{
				&node.SeqNumBox{
					&node.Hanging{"NumberedHeading", 2, nil},
					[]uint{1},
				},
				&node.SeqNumBox{
					&node.Hanging{"NumberedHeading", 3, nil},
					[]uint{1, 1},
				},
				&node.SeqNumBox{
					&node.Hanging{"NumberedHeading", 2, nil},
					[]uint{2},
				},
				&node.SeqNumBox{
					&node.Hanging{"NumberedHeading", 3, nil},
					[]uint{2, 1},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			a := make([]node.Node, len(c.in))
			copy(a, c.in)
			a = transformer.Sequence(config.Default.Elements, a)

			got := stringifier.Stringify(a...)
			want := stringifier.Stringify(c.out...)

			if got != want {
				t.Errorf("\ngot\n%s\nwant\n%s", got, want)
			}
		})
	}
}
