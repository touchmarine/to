package paragraph_test

import (
	"github.com/touchmarine/to/node"
	"github.com/touchmarine/to/stringifier"
	"github.com/touchmarine/to/transformer/paragraph"
	"testing"
)

func TestTransform(t *testing.T) {
	cases := []struct {
		name string
		in   []node.Node
		out  []node.Node
	}{
		{
			"one textblock",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
			},
			[]node.Node{
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
			},
		},
		{
			"two textblocks",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("b")}}},
			},
			[]node.Node{
				&node.Group{"GP", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
				}},
				&node.Group{"GP", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("b")}}},
				}},
			},
		},
		{
			"before walled",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
				&node.Walled{"B", nil},
			},
			[]node.Node{
				&node.Group{"GP", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
				}},
				&node.Walled{"B", nil},
			},
		},
		{
			"after walled",
			[]node.Node{
				&node.Walled{"B", nil},
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
			},
			[]node.Node{
				&node.Walled{"B", nil},
				&node.Group{"GP", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
				}},
			},
		},

		// nested
		{
			"one nested T",
			[]node.Node{
				&node.Walled{"B", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
				}},
			},
			[]node.Node{
				&node.Walled{"B", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
				}},
			},
		},
		{
			"two nested textblocks",
			[]node.Node{
				&node.Walled{"B", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("b")}}},
				}},
			},
			[]node.Node{
				&node.Walled{"B", []node.Block{
					&node.Group{"GP", []node.Block{
						&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
					}},
					&node.Group{"GP", []node.Block{
						&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("b")}}},
					}},
				}},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			paraNodes := make([]node.Node, len(c.in))
			copy(paraNodes, c.in)
			paragrapher := paragraph.Transformer{"GP"}
			paraNodes = paragrapher.Transform(paraNodes)

			got := stringifier.Stringify(paraNodes...)
			want := stringifier.Stringify(c.out...)

			if got != want {
				t.Errorf("\ngot\n%s\nwant\n%s", got, want)
			}
		})
	}
}
