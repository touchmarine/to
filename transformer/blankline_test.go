package transformer_test

import (
	"github.com/touchmarine/to/node"
	"github.com/touchmarine/to/stringifier"
	"github.com/touchmarine/to/transformer"
	"testing"
)

func TestBlankLine(t *testing.T) {
	cases := []struct {
		name string
		in   []node.Node
		out  []node.Node
	}{
		{
			"blank line",
			[]node.Node{
				&node.Line{"Line", nil},
			},
			[]node.Node{},
		},
		{
			"blank lines",
			[]node.Node{
				&node.Line{"Line", nil},
				&node.Line{"Line", nil},
			},
			[]node.Node{},
		},
		{
			"filled+blank",
			[]node.Node{
				&node.Line{"Line", []node.Inline{node.Text("a")}},
				&node.Line{"Line", nil},
			},
			[]node.Node{
				&node.Line{"Line", []node.Inline{node.Text("a")}},
			},
		},
		{
			"blank+filled",
			[]node.Node{
				&node.Line{"Line", nil},
				&node.Line{"Line", []node.Inline{node.Text("a")}},
			},
			[]node.Node{
				&node.Line{"Line", []node.Inline{node.Text("a")}},
			},
		},

		{
			"nested",
			[]node.Node{
				&node.Group{"Paragraph", []node.Block{
					&node.Line{"Line", nil},
					&node.Line{"Line", []node.Inline{node.Text("a")}},
				}},
			},
			[]node.Node{
				&node.Group{"Paragraph", []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
				}},
			},
		},
		{
			"nested 1",
			[]node.Node{
				&node.Group{"Paragraph", []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
				}},
				&node.Line{"Line", nil},
				&node.Group{"Paragraph", []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("b")}},
				}},
			},
			[]node.Node{
				&node.Group{"Paragraph", []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
				}},
				&node.Group{"Paragraph", []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("b")}},
				}},
			},
		},

		{
			"lines not settable",
			[]node.Node{
				&node.Fenced{"CodeBlock", [][]byte{nil, []byte("a")}, nil},
			},
			[]node.Node{
				&node.Fenced{"CodeBlock", [][]byte{nil, []byte("a")}, nil},
			},
		},

		{
			"nested",
			[]node.Node{
				&node.Hanging{"DescriptionList", []node.Block{
					&node.Line{"Line", nil},
					&node.Line{"Line", []node.Inline{node.Text("a")}},
				}},
			},
			[]node.Node{
				&node.Hanging{"DescriptionList", []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
				}},
			},
		},
		{
			"nested+boxed",
			[]node.Node{
				&node.SeqNumBox{
					&node.Hanging{"DescriptionList", []node.Block{
						&node.Line{"Line", nil},
						&node.Line{"Line", []node.Inline{node.Text("a")}},
					}},
					[]int{1},
				},
			},
			[]node.Node{
				&node.SeqNumBox{
					&node.Hanging{"DescriptionList", []node.Block{
						&node.Line{"Line", []node.Inline{node.Text("a")}},
					}},
					[]int{1},
				},
			},
		},

		{
			"hat",
			[]node.Node{
				&node.Hat{
					[][]byte{[]byte("a")},
					&node.Line{"Line", nil},
				},
			},
			[]node.Node{
				&node.Hat{
					[][]byte{[]byte("a")},
					nil,
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			n := make([]node.Node, len(c.in))
			copy(n, c.in)
			n = transformer.BlankLine(n)

			got := stringifier.Stringify(n...)
			want := stringifier.Stringify(c.out...)

			if got != want {
				t.Errorf("\ngot\n%s\nwant\n%s", got, want)
			}
		})
	}
}
