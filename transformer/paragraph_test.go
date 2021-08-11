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
			"one textblock",
			[]node.Node{
				&node.BasicBlock{"TextBlock", []node.Inline{node.Text("a")}},
			},
			[]node.Node{
				&node.BasicBlock{"TextBlock", []node.Inline{node.Text("a")}},
			},
		},
		{
			"two textblocks",
			[]node.Node{
				&node.BasicBlock{"TextBlock", []node.Inline{node.Text("a")}},
				&node.BasicBlock{"TextBlock", []node.Inline{node.Text("b")}},
			},
			[]node.Node{
				&node.Group{"Paragraph", []node.Block{
					&node.BasicBlock{"TextBlock", []node.Inline{node.Text("a")}},
				}},
				&node.Group{"Paragraph", []node.Block{
					&node.BasicBlock{"TextBlock", []node.Inline{node.Text("b")}},
				}},
			},
		},
		{
			"before walled",
			[]node.Node{
				&node.BasicBlock{"TextBlock", []node.Inline{node.Text("a")}},
				&node.Walled{"Blockquote", nil},
			},
			[]node.Node{
				&node.Group{"Paragraph", []node.Block{
					&node.BasicBlock{"TextBlock", []node.Inline{node.Text("a")}},
				}},
				&node.Walled{"Blockquote", nil},
			},
		},
		{
			"after walled",
			[]node.Node{
				&node.Walled{"Blockquote", nil},
				&node.BasicBlock{"TextBlock", []node.Inline{node.Text("a")}},
			},
			[]node.Node{
				&node.Walled{"Blockquote", nil},
				&node.Group{"Paragraph", []node.Block{
					&node.BasicBlock{"TextBlock", []node.Inline{node.Text("a")}},
				}},
			},
		},

		// nested
		{
			"one nested textblock",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.BasicBlock{"TextBlock", []node.Inline{node.Text("a")}},
				}},
			},
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.BasicBlock{"TextBlock", []node.Inline{node.Text("a")}},
				}},
			},
		},
		{
			"two nested textblocks",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.BasicBlock{"TextBlock", []node.Inline{node.Text("a")}},
					&node.BasicBlock{"TextBlock", []node.Inline{node.Text("b")}},
				}},
			},
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Group{"Paragraph", []node.Block{
						&node.BasicBlock{"TextBlock", []node.Inline{node.Text("a")}},
					}},
					&node.Group{"Paragraph", []node.Block{
						&node.BasicBlock{"TextBlock", []node.Inline{node.Text("b")}},
					}},
				}},
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
