package transformer_test

import (
	"github.com/touchmarine/to/config"
	"github.com/touchmarine/to/node"
	"github.com/touchmarine/to/stringifier"
	"github.com/touchmarine/to/transformer"
	"testing"
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
				&node.Hanging{"NumberedListItemDot", nil},
			},
			[]node.Node{
				&node.Group{"NumberedListDot", []node.Block{
					&node.Hanging{"NumberedListItemDot", nil},
				}},
			},
		},
		{
			"two items",
			[]node.Node{
				&node.Hanging{"NumberedListItemDot", nil},
				&node.Hanging{"NumberedListItemDot", nil},
			},
			[]node.Node{
				&node.Group{"NumberedListDot", []node.Block{
					&node.Hanging{"NumberedListItemDot", nil},
					&node.Hanging{"NumberedListItemDot", nil},
				}},
			},
		},
		{
			"three items",
			[]node.Node{
				&node.Hanging{"NumberedListItemDot", nil},
				&node.Hanging{"NumberedListItemDot", nil},
				&node.Hanging{"NumberedListItemDot", nil},
			},
			[]node.Node{
				&node.Group{"NumberedListDot", []node.Block{
					&node.Hanging{"NumberedListItemDot", nil},
					&node.Hanging{"NumberedListItemDot", nil},
					&node.Hanging{"NumberedListItemDot", nil},
				}},
			},
		},
		{
			"two single groups",
			[]node.Node{
				&node.Hanging{"NumberedListItemDot", nil},
				&node.Walled{"Blockquote", nil},
				&node.Hanging{"NumberedListItemDot", nil},
			},
			[]node.Node{
				&node.Group{"NumberedListDot", []node.Block{
					&node.Hanging{"NumberedListItemDot", nil},
				}},
				&node.Walled{"Blockquote", nil},
				&node.Group{"NumberedListDot", []node.Block{
					&node.Hanging{"NumberedListItemDot", nil},
				}},
			},
		},
		{
			"double and single group",
			[]node.Node{
				&node.Hanging{"NumberedListItemDot", nil},
				&node.Hanging{"NumberedListItemDot", nil},
				&node.Walled{"Blockquote", nil},
				&node.Hanging{"NumberedListItemDot", nil},
			},
			[]node.Node{
				&node.Group{"NumberedListDot", []node.Block{
					&node.Hanging{"NumberedListItemDot", nil},
					&node.Hanging{"NumberedListItemDot", nil},
				}},
				&node.Walled{"Blockquote", nil},
				&node.Group{"NumberedListDot", []node.Block{
					&node.Hanging{"NumberedListItemDot", nil},
				}},
			},
		},
		{
			"single and double group",
			[]node.Node{
				&node.Hanging{"NumberedListItemDot", nil},
				&node.Walled{"Blockquote", nil},
				&node.Hanging{"NumberedListItemDot", nil},
				&node.Hanging{"NumberedListItemDot", nil},
			},
			[]node.Node{
				&node.Group{"NumberedListDot", []node.Block{
					&node.Hanging{"NumberedListItemDot", nil},
				}},
				&node.Walled{"Blockquote", nil},
				&node.Group{"NumberedListDot", []node.Block{
					&node.Hanging{"NumberedListItemDot", nil},
					&node.Hanging{"NumberedListItemDot", nil},
				}},
			},
		},

		{
			"after another",
			[]node.Node{
				&node.Walled{"Blockquote", nil},
				&node.Hanging{"NumberedListItemDot", nil},
			},
			[]node.Node{
				&node.Walled{"Blockquote", nil},
				&node.Group{"NumberedListDot", []node.Block{
					&node.Hanging{"NumberedListItemDot", nil},
				}},
			},
		},
		{
			"before another",
			[]node.Node{
				&node.Hanging{"NumberedListItemDot", nil},
				&node.Walled{"Blockquote", nil},
			},
			[]node.Node{
				&node.Group{"NumberedListDot", []node.Block{
					&node.Hanging{"NumberedListItemDot", nil},
				}},
				&node.Walled{"Blockquote", nil},
			},
		},

		// nested
		{
			"nested single item",
			[]node.Node{
				&node.Hanging{"NumberedListItemDot", []node.Block{
					&node.Hanging{"NumberedListItemDot", nil},
				}},
			},
			[]node.Node{
				&node.Group{"NumberedListDot", []node.Block{
					&node.Hanging{"NumberedListItemDot", []node.Block{
						&node.Group{"NumberedListDot", []node.Block{
							&node.Hanging{"NumberedListItemDot", nil},
						}},
					}},
				}},
			},
		},
		{
			"nested two items",
			[]node.Node{
				&node.Hanging{"NumberedListItemDot", []node.Block{
					&node.Hanging{"NumberedListItemDot", nil},
				}},
				&node.Hanging{"NumberedListItemDot", []node.Block{
					&node.Hanging{"NumberedListItemDot", nil},
				}},
			},
			[]node.Node{
				&node.Group{"NumberedListDot", []node.Block{
					&node.Hanging{"NumberedListItemDot", []node.Block{
						&node.Group{"NumberedListDot", []node.Block{
							&node.Hanging{"NumberedListItemDot", nil},
						}},
					}},
					&node.Hanging{"NumberedListItemDot", []node.Block{
						&node.Group{"NumberedListDot", []node.Block{
							&node.Hanging{"NumberedListItemDot", nil},
						}},
					}},
				}},
			},
		},
		{
			"nested two single groups",
			[]node.Node{
				&node.Hanging{"NumberedListItemDot", []node.Block{
					&node.Hanging{"NumberedListItemDot", nil},
				}},
				&node.Walled{"Blockquote", nil},
				&node.Hanging{"NumberedListItemDot", []node.Block{
					&node.Hanging{"NumberedListItemDot", nil},
				}},
			},
			[]node.Node{
				&node.Group{"NumberedListDot", []node.Block{
					&node.Hanging{"NumberedListItemDot", []node.Block{
						&node.Group{"NumberedListDot", []node.Block{
							&node.Hanging{"NumberedListItemDot", nil},
						}},
					}},
				}},
				&node.Walled{"Blockquote", nil},
				&node.Group{"NumberedListDot", []node.Block{
					&node.Hanging{"NumberedListItemDot", []node.Block{
						&node.Group{"NumberedListDot", []node.Block{
							&node.Hanging{"NumberedListItemDot", nil},
						}},
					}},
				}},
			},
		},
		{
			"nested in another",
			[]node.Node{
				&node.Hanging{"NumberedListItemDot", []node.Block{
					&node.Walled{"Paragraph", []node.Block{
						&node.Hanging{"ListItemDot", nil},
					}},
				}},
			},
			[]node.Node{
				&node.Group{"NumberedListDot", []node.Block{
					&node.Hanging{"NumberedListItemDot", []node.Block{
						&node.Walled{"Paragraph", []node.Block{
							&node.Group{"ListDot", []node.Block{
								&node.Hanging{"ListItemDot", nil},
							}},
						}},
					}},
				}},
			},
		},

		{
			"nested non groupable",
			[]node.Node{
				&node.Hanging{"NumberedListItemDot", []node.Block{
					&node.Walled{"Blockquote", nil},
					&node.Walled{"Blockquote", nil},
				}},
			},
			[]node.Node{
				&node.Group{"NumberedListDot", []node.Block{
					&node.Hanging{"NumberedListItemDot", []node.Block{
						&node.Walled{"Blockquote", nil},
						&node.Walled{"Blockquote", nil},
					}},
				}},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			groupedNodes := make([]node.Node, len(c.in))
			copy(groupedNodes, c.in)
			groupedNodes = transformer.Group(config.Default.Groups, groupedNodes)

			got := stringifier.Stringify(groupedNodes...)
			want := stringifier.Stringify(c.out...)

			if got != want {
				t.Errorf("\ngot\n%s\nwant\n%s", got, want)
			}
		})
	}
}
