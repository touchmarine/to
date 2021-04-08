package transformer_test

import (
	"testing"
	"to/internal/config"
	"to/internal/node"
	"to/internal/stringifier"
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
				&node.Line{"BlankLine", nil},
				&node.Hanging{"NumberedListItemDot", 0, nil},
			},
			[]node.Node{
				&node.Group{"NumberedListDot", []node.Block{
					&node.Hanging{"NumberedListItemDot", 0, nil},
				}},
				&node.Line{"BlankLine", nil},
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
				&node.Line{"BlankLine", nil},
				&node.Hanging{"NumberedListItemDot", 0, nil},
			},
			[]node.Node{
				&node.Group{"NumberedListDot", []node.Block{
					&node.Hanging{"NumberedListItemDot", 0, nil},
					&node.Hanging{"NumberedListItemDot", 0, nil},
				}},
				&node.Line{"BlankLine", nil},
				&node.Group{"NumberedListDot", []node.Block{
					&node.Hanging{"NumberedListItemDot", 0, nil},
				}},
			},
		},
		{
			"single and double group",
			[]node.Node{
				&node.Hanging{"NumberedListItemDot", 0, nil},
				&node.Line{"BlankLine", nil},
				&node.Hanging{"NumberedListItemDot", 0, nil},
				&node.Hanging{"NumberedListItemDot", 0, nil},
			},
			[]node.Node{
				&node.Group{"NumberedListDot", []node.Block{
					&node.Hanging{"NumberedListItemDot", 0, nil},
				}},
				&node.Line{"BlankLine", nil},
				&node.Group{"NumberedListDot", []node.Block{
					&node.Hanging{"NumberedListItemDot", 0, nil},
					&node.Hanging{"NumberedListItemDot", 0, nil},
				}},
			},
		},

		{
			"after another",
			[]node.Node{
				&node.Line{"BlankLine", nil},
				&node.Hanging{"NumberedListItemDot", 0, nil},
			},
			[]node.Node{
				&node.Line{"BlankLine", nil},
				&node.Group{"NumberedListDot", []node.Block{
					&node.Hanging{"NumberedListItemDot", 0, nil},
				}},
			},
		},
		{
			"before another",
			[]node.Node{
				&node.Hanging{"NumberedListItemDot", 0, nil},
				&node.Line{"BlankLine", nil},
			},
			[]node.Node{
				&node.Group{"NumberedListDot", []node.Block{
					&node.Hanging{"NumberedListItemDot", 0, nil},
				}},
				&node.Line{"BlankLine", nil},
			},
		},

		// nested
		{
			"nested single item",
			[]node.Node{
				&node.Hanging{"NumberedListItemDot", 0, []node.Block{
					&node.Hanging{"NumberedListItemDot", 0, nil},
				}},
			},
			[]node.Node{
				&node.Group{"NumberedListDot", []node.Block{
					&node.Hanging{"NumberedListItemDot", 0, []node.Block{
						&node.Group{"NumberedListDot", []node.Block{
							&node.Hanging{"NumberedListItemDot", 0, nil},
						}},
					}},
				}},
			},
		},
		{
			"nested two items",
			[]node.Node{
				&node.Hanging{"NumberedListItemDot", 0, []node.Block{
					&node.Hanging{"NumberedListItemDot", 0, nil},
				}},
				&node.Hanging{"NumberedListItemDot", 0, []node.Block{
					&node.Hanging{"NumberedListItemDot", 0, nil},
				}},
			},
			[]node.Node{
				&node.Group{"NumberedListDot", []node.Block{
					&node.Hanging{"NumberedListItemDot", 0, []node.Block{
						&node.Group{"NumberedListDot", []node.Block{
							&node.Hanging{"NumberedListItemDot", 0, nil},
						}},
					}},
					&node.Hanging{"NumberedListItemDot", 0, []node.Block{
						&node.Group{"NumberedListDot", []node.Block{
							&node.Hanging{"NumberedListItemDot", 0, nil},
						}},
					}},
				}},
			},
		},
		{
			"nested two single groups",
			[]node.Node{
				&node.Hanging{"NumberedListItemDot", 0, []node.Block{
					&node.Hanging{"NumberedListItemDot", 0, nil},
				}},
				&node.Line{"BlankLine", nil},
				&node.Hanging{"NumberedListItemDot", 0, []node.Block{
					&node.Hanging{"NumberedListItemDot", 0, nil},
				}},
			},
			[]node.Node{
				&node.Group{"NumberedListDot", []node.Block{
					&node.Hanging{"NumberedListItemDot", 0, []node.Block{
						&node.Group{"NumberedListDot", []node.Block{
							&node.Hanging{"NumberedListItemDot", 0, nil},
						}},
					}},
				}},
				&node.Line{"BlankLine", nil},
				&node.Group{"NumberedListDot", []node.Block{
					&node.Hanging{"NumberedListItemDot", 0, []node.Block{
						&node.Group{"NumberedListDot", []node.Block{
							&node.Hanging{"NumberedListItemDot", 0, nil},
						}},
					}},
				}},
			},
		},

		// group nested
		{
			"top level",
			[]node.Node{
				&node.Line{"Line", nil},
				&node.Line{"Line", nil},
			},
			[]node.Node{
				&node.Group{"Paragraph", []node.Block{
					&node.Line{"Line", nil},
					&node.Line{"Line", nil},
				}},
			},
		},
		{
			"nested non groupable",
			[]node.Node{
				&node.Hanging{"NumberedListItemDot", 0, []node.Block{
					&node.Line{"Line", nil},
					&node.Line{"Line", nil},
				}},
			},
			[]node.Node{
				&node.Group{"NumberedListDot", []node.Block{
					&node.Hanging{"NumberedListItemDot", 0, []node.Block{
						&node.Line{"Line", nil},
						&node.Line{"Line", nil},
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
