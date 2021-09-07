package group_test

import (
	"github.com/touchmarine/to/node"
	"github.com/touchmarine/to/stringifier"
	"github.com/touchmarine/to/transformer/group"
	"testing"
)

func TestTransform(t *testing.T) {
	cases := []struct {
		name string
		in   []node.Node
		out  []node.Node
	}{
		{
			"single item",
			[]node.Node{
				&node.Hanging{"A", nil},
			},
			[]node.Node{
				&node.Group{"GA", []node.Block{
					&node.Hanging{"A", nil},
				}},
			},
		},
		{
			"two items",
			[]node.Node{
				&node.Hanging{"A", nil},
				&node.Hanging{"A", nil},
			},
			[]node.Node{
				&node.Group{"GA", []node.Block{
					&node.Hanging{"A", nil},
					&node.Hanging{"A", nil},
				}},
			},
		},
		{
			"three items",
			[]node.Node{
				&node.Hanging{"A", nil},
				&node.Hanging{"A", nil},
				&node.Hanging{"A", nil},
			},
			[]node.Node{
				&node.Group{"GA", []node.Block{
					&node.Hanging{"A", nil},
					&node.Hanging{"A", nil},
					&node.Hanging{"A", nil},
				}},
			},
		},
		{
			"two single groups",
			[]node.Node{
				&node.Hanging{"A", nil},
				&node.Walled{"B", nil},
				&node.Hanging{"A", nil},
			},
			[]node.Node{
				&node.Group{"GA", []node.Block{
					&node.Hanging{"A", nil},
				}},
				&node.Walled{"B", nil},
				&node.Group{"GA", []node.Block{
					&node.Hanging{"A", nil},
				}},
			},
		},
		{
			"double and single group",
			[]node.Node{
				&node.Hanging{"A", nil},
				&node.Hanging{"A", nil},
				&node.Walled{"B", nil},
				&node.Hanging{"A", nil},
			},
			[]node.Node{
				&node.Group{"GA", []node.Block{
					&node.Hanging{"A", nil},
					&node.Hanging{"A", nil},
				}},
				&node.Walled{"B", nil},
				&node.Group{"GA", []node.Block{
					&node.Hanging{"A", nil},
				}},
			},
		},
		{
			"single and double group",
			[]node.Node{
				&node.Hanging{"A", nil},
				&node.Walled{"B", nil},
				&node.Hanging{"A", nil},
				&node.Hanging{"A", nil},
			},
			[]node.Node{
				&node.Group{"GA", []node.Block{
					&node.Hanging{"A", nil},
				}},
				&node.Walled{"B", nil},
				&node.Group{"GA", []node.Block{
					&node.Hanging{"A", nil},
					&node.Hanging{"A", nil},
				}},
			},
		},

		{
			"after another",
			[]node.Node{
				&node.Walled{"B", nil},
				&node.Hanging{"A", nil},
			},
			[]node.Node{
				&node.Walled{"B", nil},
				&node.Group{"GA", []node.Block{
					&node.Hanging{"A", nil},
				}},
			},
		},
		{
			"before another",
			[]node.Node{
				&node.Hanging{"A", nil},
				&node.Walled{"B", nil},
			},
			[]node.Node{
				&node.Group{"GA", []node.Block{
					&node.Hanging{"A", nil},
				}},
				&node.Walled{"B", nil},
			},
		},

		// nested
		{
			"nested single item",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Hanging{"A", nil},
				}},
			},
			[]node.Node{
				&node.Group{"GA", []node.Block{
					&node.Hanging{"A", []node.Block{
						&node.Group{"GA", []node.Block{
							&node.Hanging{"A", nil},
						}},
					}},
				}},
			},
		},
		{
			"nested two items",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Hanging{"A", nil},
				}},
				&node.Hanging{"A", []node.Block{
					&node.Hanging{"A", nil},
				}},
			},
			[]node.Node{
				&node.Group{"GA", []node.Block{
					&node.Hanging{"A", []node.Block{
						&node.Group{"GA", []node.Block{
							&node.Hanging{"A", nil},
						}},
					}},
					&node.Hanging{"A", []node.Block{
						&node.Group{"GA", []node.Block{
							&node.Hanging{"A", nil},
						}},
					}},
				}},
			},
		},
		{
			"nested two single groups",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Hanging{"A", nil},
				}},
				&node.Walled{"B", nil},
				&node.Hanging{"A", []node.Block{
					&node.Hanging{"A", nil},
				}},
			},
			[]node.Node{
				&node.Group{"GA", []node.Block{
					&node.Hanging{"A", []node.Block{
						&node.Group{"GA", []node.Block{
							&node.Hanging{"A", nil},
						}},
					}},
				}},
				&node.Walled{"B", nil},
				&node.Group{"GA", []node.Block{
					&node.Hanging{"A", []node.Block{
						&node.Group{"GA", []node.Block{
							&node.Hanging{"A", nil},
						}},
					}},
				}},
			},
		},
		{
			"nested in another",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Walled{"C", []node.Block{
						&node.Hanging{"D", nil},
					}},
				}},
			},
			[]node.Node{
				&node.Group{"GA", []node.Block{
					&node.Hanging{"A", []node.Block{
						&node.Walled{"C", []node.Block{
							&node.Group{"GB", []node.Block{
								&node.Hanging{"D", nil},
							}},
						}},
					}},
				}},
			},
		},

		{
			"nested non groupable",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Walled{"B", nil},
					&node.Walled{"B", nil},
				}},
			},
			[]node.Node{
				&node.Group{"GA", []node.Block{
					&node.Hanging{"A", []node.Block{
						&node.Walled{"B", nil},
						&node.Walled{"B", nil},
					}},
				}},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			groupedNodes := make([]node.Node, len(c.in))
			copy(groupedNodes, c.in)
			grouper := group.Transformer{group.Map{
				"A": {
					Name:    "GA",
					Element: "A",
				},
				"D": {
					Name:    "GB",
					Element: "D",
				},
			}}
			groupedNodes = grouper.Transform(groupedNodes)

			got := stringifier.Stringify(groupedNodes...)
			want := stringifier.Stringify(c.out...)

			if got != want {
				t.Errorf("\ngot\n%s\nwant\n%s", got, want)
			}
		})
	}
}
