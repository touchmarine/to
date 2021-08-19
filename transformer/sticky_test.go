package transformer_test

import (
	"github.com/touchmarine/to/config"
	"github.com/touchmarine/to/node"
	"github.com/touchmarine/to/stringifier"
	"github.com/touchmarine/to/transformer"
	"testing"
)

func TestGroupStickies(t *testing.T) {
	cases := []struct {
		name string
		in   []node.Node
		out  []node.Node
	}{
		// before
		{
			"before nothing",
			[]node.Node{
				&node.VerbatimWalled{"A", [][]byte{[]byte("a")}},
			},
			[]node.Node{
				&node.VerbatimWalled{"A", [][]byte{[]byte("a")}},
			},
		},
		{
			"before",
			[]node.Node{
				&node.VerbatimWalled{"A", [][]byte{[]byte("b")}},
				&node.BasicBlock{"TextBlock", []node.Inline{node.Text("a")}},
			},
			[]node.Node{
				&node.Sticky{"SA", false, []node.Block{
					&node.VerbatimWalled{"A", [][]byte{[]byte("b")}},
					&node.BasicBlock{"TextBlock", []node.Inline{node.Text("a")}},
				}},
			},
		},
		{
			"before sticky",
			[]node.Node{
				&node.VerbatimWalled{"A", [][]byte{[]byte("a")}},
				&node.VerbatimWalled{"A", [][]byte{[]byte("b")}},
			},
			[]node.Node{
				&node.VerbatimWalled{"A", [][]byte{[]byte("a")}},
				&node.VerbatimWalled{"A", [][]byte{[]byte("b")}},
			},
		},
		{
			"before sticky 1",
			[]node.Node{
				&node.VerbatimWalled{"A", [][]byte{[]byte("a")}},
				&node.VerbatimWalled{"A", [][]byte{[]byte("b")}},
				&node.BasicBlock{"TextBlock", []node.Inline{node.Text("c")}},
			},
			[]node.Node{
				&node.VerbatimWalled{"A", [][]byte{[]byte("a")}},
				&node.Sticky{"SA", false, []node.Block{
					&node.VerbatimWalled{"A", [][]byte{[]byte("b")}},
					&node.BasicBlock{"TextBlock", []node.Inline{node.Text("c")}},
				}},
			},
		},
		{
			"before in after position",
			[]node.Node{
				&node.BasicBlock{"TextBlock", []node.Inline{node.Text("a")}},
				&node.VerbatimWalled{"A", [][]byte{[]byte("b")}},
			},
			[]node.Node{
				&node.BasicBlock{"TextBlock", []node.Inline{node.Text("a")}},
				&node.VerbatimWalled{"A", [][]byte{[]byte("b")}},
			},
		},

		// after
		{
			"after nothing",
			[]node.Node{
				&node.VerbatimWalled{"B", [][]byte{[]byte("a")}},
			},
			[]node.Node{
				&node.VerbatimWalled{"B", [][]byte{[]byte("a")}},
			},
		},
		{
			"after",
			[]node.Node{
				&node.BasicBlock{"TextBlock", []node.Inline{node.Text("a")}},
				&node.VerbatimWalled{"B", [][]byte{[]byte("b")}},
			},
			[]node.Node{
				&node.Sticky{"SB", true, []node.Block{
					&node.BasicBlock{"TextBlock", []node.Inline{node.Text("a")}},
					&node.VerbatimWalled{"B", [][]byte{[]byte("b")}},
				}},
			},
		},
		{
			"after sticky",
			[]node.Node{
				&node.VerbatimWalled{"B", [][]byte{[]byte("a")}},
				&node.VerbatimWalled{"B", [][]byte{[]byte("b")}},
			},
			[]node.Node{
				&node.VerbatimWalled{"B", [][]byte{[]byte("a")}},
				&node.VerbatimWalled{"B", [][]byte{[]byte("b")}},
			},
		},
		{
			"after sticky 1",
			[]node.Node{
				&node.BasicBlock{"TextBlock", []node.Inline{node.Text("a")}},
				&node.VerbatimWalled{"B", [][]byte{[]byte("b")}},
				&node.VerbatimWalled{"B", [][]byte{[]byte("c")}},
			},
			[]node.Node{
				&node.Sticky{"SB", true, []node.Block{
					&node.BasicBlock{"TextBlock", []node.Inline{node.Text("a")}},
					&node.VerbatimWalled{"B", [][]byte{[]byte("b")}},
				}},
				&node.VerbatimWalled{"B", [][]byte{[]byte("c")}},
			},
		},
		{
			"after in before position",
			[]node.Node{
				&node.VerbatimWalled{"B", [][]byte{[]byte("b")}},
				&node.BasicBlock{"TextBlock", []node.Inline{node.Text("a")}},
			},
			[]node.Node{
				&node.VerbatimWalled{"B", [][]byte{[]byte("b")}},
				&node.BasicBlock{"TextBlock", []node.Inline{node.Text("a")}},
			},
		},

		// before and after
		{
			"before and after",
			[]node.Node{
				&node.VerbatimWalled{"A", [][]byte{[]byte("a")}},
				&node.BasicBlock{"TextBlock", []node.Inline{node.Text("b")}},
				&node.VerbatimWalled{"B", [][]byte{[]byte("c")}},
			},
			[]node.Node{
				&node.Sticky{"SB", true, []node.Block{
					&node.Sticky{"SA", false, []node.Block{
						&node.VerbatimWalled{"A", [][]byte{[]byte("a")}},
						&node.BasicBlock{"TextBlock", []node.Inline{node.Text("b")}},
					}},
					&node.VerbatimWalled{"B", [][]byte{[]byte("c")}},
				}},
			},
		},
		{
			"before and after 2",
			[]node.Node{
				&node.VerbatimWalled{"A", [][]byte{[]byte("a")}},
				&node.VerbatimWalled{"A", [][]byte{[]byte("b")}},
				&node.BasicBlock{"TextBlock", []node.Inline{node.Text("c")}},
				&node.VerbatimWalled{"B", [][]byte{[]byte("d")}},
				&node.VerbatimWalled{"B", [][]byte{[]byte("e")}},
			},
			[]node.Node{
				&node.VerbatimWalled{"A", [][]byte{[]byte("a")}},
				&node.Sticky{"SB", true, []node.Block{
					&node.Sticky{"SA", false, []node.Block{
						&node.VerbatimWalled{"A", [][]byte{[]byte("b")}},
						&node.BasicBlock{"TextBlock", []node.Inline{node.Text("c")}},
					}},
					&node.VerbatimWalled{"B", [][]byte{[]byte("d")}},
				}},
				&node.VerbatimWalled{"B", [][]byte{[]byte("e")}},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			stickied := make([]node.Node, len(c.in))
			copy(stickied, c.in)
			stickied = transformer.GroupStickies([]config.Sticky{
				{
					Name:    "SA",
					Element: "A",
				},
				{
					Name:    "SB",
					Element: "B",
					After:   true,
				},
			}, stickied)

			got := stringifier.Stringify(stickied...)
			want := stringifier.Stringify(c.out...)

			if got != want {
				t.Errorf("\ngot\n%s\nwant\n%s", got, want)
			}
		})
	}
}
