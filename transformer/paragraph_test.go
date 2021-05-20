package transformer_test

import (
	"github.com/touchmarine/to/node"
	"github.com/touchmarine/to/stringifier"
	"github.com/touchmarine/to/transformer"
	"testing"
)

func TestParagraph(t *testing.T) {
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
			[]node.Node{
				&node.Line{"Line", nil},
			},
		},
		{
			"filled line",
			[]node.Node{
				&node.Line{"Line", []node.Inline{node.Text("a")}},
			},
			[]node.Node{
				&node.Line{"Line", []node.Inline{node.Text("a")}},
			},
		},
		{
			"filled and blank line",
			[]node.Node{
				&node.Line{"Line", []node.Inline{node.Text("a")}},
				&node.Line{"Line", nil},
			},
			[]node.Node{
				&node.Line{"Line", []node.Inline{node.Text("a")}},
				&node.Line{"Line", nil},
			},
		},

		{
			"one filled",
			[]node.Node{
				&node.Line{"Line", []node.Inline{node.Text("a")}},
				&node.Line{"Line", nil},
				&node.Line{"Line", []node.Inline{node.Text("b")}},
			},
			[]node.Node{
				&node.Group{"Paragraph", []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
				}},
				&node.Line{"Line", nil},
				&node.Group{"Paragraph", []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("b")}},
				}},
			},
		},
		{
			"two filled",
			[]node.Node{
				&node.Line{"Line", []node.Inline{node.Text("a")}},
				&node.Line{"Line", []node.Inline{node.Text("b")}},
				&node.Line{"Line", nil},
				&node.Line{"Line", []node.Inline{node.Text("c")}},
			},
			[]node.Node{
				&node.Group{"Paragraph", []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
					&node.Line{"Line", []node.Inline{node.Text("b")}},
				}},
				&node.Line{"Line", nil},
				&node.Group{"Paragraph", []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("c")}},
				}},
			},
		},

		{
			"multiple blank lines",
			[]node.Node{
				&node.Line{"Line", []node.Inline{node.Text("a")}},
				&node.Line{"Line", nil},
				&node.Line{"Line", nil},
				&node.Line{"Line", []node.Inline{node.Text("b")}},
			},
			[]node.Node{
				&node.Group{"Paragraph", []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
				}},
				&node.Line{"Line", nil},
				&node.Line{"Line", nil},
				&node.Group{"Paragraph", []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("b")}},
				}},
			},
		},

		{
			"alternating",
			[]node.Node{
				&node.Line{"Line", []node.Inline{node.Text("a")}},
				&node.Line{"Line", nil},
				&node.Line{"Line", []node.Inline{node.Text("b")}},
				&node.Line{"Line", nil},
				&node.Line{"Line", []node.Inline{node.Text("c")}},
			},
			[]node.Node{
				&node.Group{"Paragraph", []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
				}},
				&node.Line{"Line", nil},
				&node.Group{"Paragraph", []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("b")}},
				}},
				&node.Line{"Line", nil},
				&node.Group{"Paragraph", []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("c")}},
				}},
			},
		},
		{
			"alternating 1",
			[]node.Node{
				&node.Line{"Line", []node.Inline{node.Text("a")}},
				&node.Line{"Line", nil},
				&node.Line{"Line", []node.Inline{node.Text("b")}},
				&node.Line{"Line", nil},
			},
			[]node.Node{
				&node.Group{"Paragraph", []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
				}},
				&node.Line{"Line", nil},
				&node.Group{"Paragraph", []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("b")}},
				}},
				&node.Line{"Line", nil},
			},
		},

		{
			"before walled",
			[]node.Node{
				&node.Line{"Line", []node.Inline{node.Text("a")}},
				&node.Line{"Line", nil},
				&node.Walled{"Blockquote", nil},
			},
			[]node.Node{
				&node.Group{"Paragraph", []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
				}},
				&node.Line{"Line", nil},
				&node.Walled{"Blockquote", nil},
			},
		},
		{
			"before walled 1",
			[]node.Node{
				&node.Line{"Line", []node.Inline{node.Text("a")}},
				&node.Walled{"Blockquote", nil},
			},
			[]node.Node{
				&node.Group{"Paragraph", []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
				}},
				&node.Walled{"Blockquote", nil},
			},
		},
		{
			"before walled 2",
			[]node.Node{
				&node.Line{"Line", []node.Inline{node.Text("a")}},
				&node.Walled{"Blockquote", nil},
				&node.Line{"Line", nil},
			},
			[]node.Node{
				&node.Group{"Paragraph", []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
				}},
				&node.Walled{"Blockquote", nil},
				&node.Line{"Line", nil},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			paraNodes := make([]node.Node, len(c.in))
			copy(paraNodes, c.in)
			paraNodes = transformer.Paragraph(paraNodes)

			got := stringifier.Stringify(paraNodes...)
			want := stringifier.Stringify(c.out...)

			if got != want {
				t.Errorf("\ngot\n%s\nwant\n%s", got, want)
			}
		})
	}
}
