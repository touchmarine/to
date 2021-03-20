package transformer_test

import (
	"encoding/json"
	"testing"
	"to/internal/node"
	"to/internal/transformer"
)

func TestGroup(t *testing.T) {
	cases := []struct {
		name string
		in   []node.Node
		out  []node.Node
	}{
		{
			"single item",
			[]node.Node{
				&node.Hanging{"NumberedListItemDot", 0, nil},
			},
			[]node.Node{
				&node.Group{"NumberedListDot", []node.Block{
					&node.Hanging{"NumberedListItemDot", 0, nil},
				}},
			},
		},
		{
			"two items",
			[]node.Node{
				&node.Hanging{"NumberedListItemDot", 0, nil},
				&node.Hanging{"NumberedListItemDot", 0, nil},
			},
			[]node.Node{
				&node.Group{"NumberedListDot", []node.Block{
					&node.Hanging{"NumberedListItemDot", 0, nil},
					&node.Hanging{"NumberedListItemDot", 0, nil},
				}},
			},
		},
		{
			"three items",
			[]node.Node{
				&node.Hanging{"NumberedListItemDot", 0, nil},
				&node.Hanging{"NumberedListItemDot", 0, nil},
				&node.Hanging{"NumberedListItemDot", 0, nil},
			},
			[]node.Node{
				&node.Group{"NumberedListDot", []node.Block{
					&node.Hanging{"NumberedListItemDot", 0, nil},
					&node.Hanging{"NumberedListItemDot", 0, nil},
					&node.Hanging{"NumberedListItemDot", 0, nil},
				}},
			},
		},
		{
			"two single groups",
			[]node.Node{
				&node.Hanging{"NumberedListItemDot", 0, nil},
				&node.Line{"Line", nil},
				&node.Hanging{"NumberedListItemDot", 0, nil},
			},
			[]node.Node{
				&node.Group{"NumberedListDot", []node.Block{
					&node.Hanging{"NumberedListItemDot", 0, nil},
				}},
				&node.Line{"Line", nil},
				&node.Group{"NumberedListDot", []node.Block{
					&node.Hanging{"NumberedListItemDot", 0, nil},
				}},
			},
		},
		{
			"double and single group",
			[]node.Node{
				&node.Hanging{"NumberedListItemDot", 0, nil},
				&node.Hanging{"NumberedListItemDot", 0, nil},
				&node.Line{"Line", nil},
				&node.Hanging{"NumberedListItemDot", 0, nil},
			},
			[]node.Node{
				&node.Group{"NumberedListDot", []node.Block{
					&node.Hanging{"NumberedListItemDot", 0, nil},
					&node.Hanging{"NumberedListItemDot", 0, nil},
				}},
				&node.Line{"Line", nil},
				&node.Group{"NumberedListDot", []node.Block{
					&node.Hanging{"NumberedListItemDot", 0, nil},
				}},
			},
		},
		{
			"single and double group",
			[]node.Node{
				&node.Hanging{"NumberedListItemDot", 0, nil},
				&node.Line{"Line", nil},
				&node.Hanging{"NumberedListItemDot", 0, nil},
				&node.Hanging{"NumberedListItemDot", 0, nil},
			},
			[]node.Node{
				&node.Group{"NumberedListDot", []node.Block{
					&node.Hanging{"NumberedListItemDot", 0, nil},
				}},
				&node.Line{"Line", nil},
				&node.Group{"NumberedListDot", []node.Block{
					&node.Hanging{"NumberedListItemDot", 0, nil},
					&node.Hanging{"NumberedListItemDot", 0, nil},
				}},
			},
		},

		{
			"after another",
			[]node.Node{
				&node.Line{"Line", nil},
				&node.Hanging{"NumberedListItemDot", 0, nil},
			},
			[]node.Node{
				&node.Line{"Line", nil},
				&node.Group{"NumberedListDot", []node.Block{
					&node.Hanging{"NumberedListItemDot", 0, nil},
				}},
			},
		},
		{
			"before another",
			[]node.Node{
				&node.Hanging{"NumberedListItemDot", 0, nil},
				&node.Line{"Line", nil},
			},
			[]node.Node{
				&node.Group{"NumberedListDot", []node.Block{
					&node.Hanging{"NumberedListItemDot", 0, nil},
				}},
				&node.Line{"Line", nil},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			groupedNodes := make([]node.Node, len(c.in))
			copy(groupedNodes, c.in)
			groupedNodes = transformer.Group(groupedNodes)

			groupedJSON, err := json.MarshalIndent(groupedNodes, "", "\t")
			if err != nil {
				t.Fatal(err)
			}

			outJSON, err := json.MarshalIndent(c.out, "", "\t")
			if err != nil {
				t.Fatal(err)
			}

			if got, want := string(groupedJSON), string(outJSON); got != want {
				t.Errorf("\ngot\n%s\nwant\n%s", got, want)
			}
		})
	}
}
