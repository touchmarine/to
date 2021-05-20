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
				&node.Line{"Line", nil},
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
				&node.Line{"Line", nil},
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
			"nested in another",
			[]node.Node{
				&node.Hanging{"NumberedListItemDot", 0, []node.Block{
					&node.Walled{"Paragraph", []node.Block{
						&node.Hanging{"ListItemDot", 0, nil},
					}},
				}},
			},
			[]node.Node{
				&node.Group{"NumberedListDot", []node.Block{
					&node.Hanging{"NumberedListItemDot", 0, []node.Block{
						&node.Walled{"Paragraph", []node.Block{
							&node.Group{"ListDot", []node.Block{
								&node.Hanging{"ListItemDot", 0, nil},
							}},
						}},
					}},
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

		// boxed
		{
			"hat",
			[]node.Node{
				&node.Hat{
					[][]byte{[]byte("a")},
					&node.Hanging{"NumberedListItemDot", 0, nil},
				},
			},
			[]node.Node{
				&node.Hat{
					[][]byte{[]byte("a")},
					&node.Group{"NumberedListDot", []node.Block{
						&node.Hat{
							[][]byte{[]byte("a")},
							&node.Hanging{"NumberedListItemDot", 0, nil},
						},
					}},
				},
			},
		},
		{
			"two hats",
			[]node.Node{
				&node.Hat{
					[][]byte{[]byte("a")},
					&node.Hanging{"NumberedListItemDot", 0, nil},
				},
				&node.Hat{
					[][]byte{[]byte("b")},
					&node.Hanging{"NumberedListItemDot", 0, nil},
				},
			},
			[]node.Node{
				&node.Hat{
					[][]byte{[]byte("a")},
					&node.Group{"NumberedListDot", []node.Block{
						&node.Hat{
							[][]byte{[]byte("a")},
							&node.Hanging{"NumberedListItemDot", 0, nil},
						},
					}},
				},
				&node.Hat{
					[][]byte{[]byte("b")},
					&node.Group{"NumberedListDot", []node.Block{
						&node.Hat{
							[][]byte{[]byte("b")},
							&node.Hanging{"NumberedListItemDot", 0, nil},
						},
					}},
				},
			},
		},
		{
			"two hats 2",
			[]node.Node{
				&node.Hat{
					[][]byte{[]byte("a")},
					&node.Hanging{"NumberedListItemDot", 0, nil},
				},
				&node.Line{"Line", nil},
				&node.Hat{
					[][]byte{[]byte("b")},
					&node.Hanging{"NumberedListItemDot", 0, nil},
				},
			},
			[]node.Node{
				&node.Hat{
					[][]byte{[]byte("a")},
					&node.Group{"NumberedListDot", []node.Block{
						&node.Hat{
							[][]byte{[]byte("a")},
							&node.Hanging{"NumberedListItemDot", 0, nil},
						},
					}},
				},
				&node.Line{"Line", nil},
				&node.Hat{
					[][]byte{[]byte("b")},
					&node.Group{"NumberedListDot", []node.Block{
						&node.Hat{
							[][]byte{[]byte("b")},
							&node.Hanging{"NumberedListItemDot", 0, nil},
						},
					}},
				},
			},
		},
		{
			"mixed hat",
			[]node.Node{
				&node.Hat{
					[][]byte{},
					&node.Hanging{"NumberedListItemDot", 0, nil},
				},
				&node.Hanging{"NumberedListItemDot", 0, nil},
			},
			[]node.Node{
				&node.Hat{
					[][]byte{},
					&node.Group{"NumberedListDot", []node.Block{
						&node.Hat{
							[][]byte{},
							&node.Hanging{"NumberedListItemDot", 0, nil},
						},
						&node.Hanging{"NumberedListItemDot", 0, nil},
					}},
				},
			},
		},
		{
			"mixed hat reverse",
			[]node.Node{
				&node.Hanging{"NumberedListItemDot", 0, nil},
				&node.Hat{
					[][]byte{},
					&node.Hanging{"NumberedListItemDot", 0, nil},
				},
			},
			[]node.Node{
				&node.Hat{
					[][]byte{},
					&node.Group{"NumberedListDot", []node.Block{
						&node.Hanging{"NumberedListItemDot", 0, nil},
						&node.Hat{
							[][]byte{},
							&node.Hanging{"NumberedListItemDot", 0, nil},
						},
					}},
				},
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
