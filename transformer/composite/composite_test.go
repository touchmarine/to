package composite_test

import (
	"github.com/touchmarine/to/node"
	"github.com/touchmarine/to/stringifier"
	"github.com/touchmarine/to/transformer/composite"
	"testing"
)

func TestTransform(t *testing.T) {
	cases := []struct {
		name string
		in   []node.Node
		out  []node.Node
	}{
		{
			"basic",
			[]node.Node{
				&node.Leaf{"A", []node.Inline{
					&node.Uniform{"B", nil},
					&node.Escaped{"C", nil},
				}},
			},
			[]node.Node{
				&node.Leaf{"A", []node.Inline{
					&node.Composite{
						"CA",
						&node.Uniform{"B", nil},
						&node.Escaped{"C", nil},
					},
				}},
			},
		},
		{
			"filled",
			[]node.Node{
				&node.Leaf{"A", []node.Inline{
					&node.Uniform{"B", []node.Inline{
						&node.Text{"MT", []byte("a")},
					}},
					&node.Escaped{"C", []byte("b")},
				}},
			},
			[]node.Node{
				&node.Leaf{"A", []node.Inline{
					&node.Composite{
						"CA",
						&node.Uniform{"B", []node.Inline{
							&node.Text{"MT", []byte("a")},
						}},
						&node.Escaped{"C", []byte("b")},
					},
				}},
			},
		},
		{
			"nested",
			[]node.Node{
				&node.Leaf{"A", []node.Inline{
					&node.Uniform{"B", []node.Inline{
						&node.Uniform{"B", nil},
						&node.Escaped{"C", nil},
					}},
					&node.Escaped{"C", nil},
				}},
			},
			[]node.Node{
				&node.Leaf{"A", []node.Inline{
					&node.Composite{
						"CA",
						&node.Uniform{"B", []node.Inline{
							&node.Uniform{"B", nil},
							&node.Escaped{"C", nil},
						}},
						&node.Escaped{"C", nil},
					},
				}},
			},
		},
		{
			"consecutive",
			[]node.Node{
				&node.Leaf{"A", []node.Inline{
					&node.Uniform{"B", nil},
					&node.Escaped{"C", nil},
					&node.Uniform{"B", nil},
					&node.Escaped{"C", nil},
				}},
			},
			[]node.Node{
				&node.Leaf{"A", []node.Inline{
					&node.Composite{
						"CA",
						&node.Uniform{"B", nil},
						&node.Escaped{"C", nil},
					},
					&node.Composite{
						"CA",
						&node.Uniform{"B", nil},
						&node.Escaped{"C", nil},
					},
				}},
			},
		},
		{
			"two lines",
			[]node.Node{
				&node.Leaf{"A", []node.Inline{
					&node.Uniform{"B", nil},
					&node.Escaped{"C", nil},
				}},
				&node.Leaf{"A", []node.Inline{
					&node.Uniform{"B", nil},
					&node.Escaped{"C", nil},
				}},
			},
			[]node.Node{
				&node.Leaf{"A", []node.Inline{
					&node.Composite{
						"CA",
						&node.Uniform{"B", nil},
						&node.Escaped{"C", nil},
					},
				}},
				&node.Leaf{"A", []node.Inline{
					&node.Composite{
						"CA",
						&node.Uniform{"B", nil},
						&node.Escaped{"C", nil},
					},
				}},
			},
		},
		{
			"nested line",
			[]node.Node{
				&node.Walled{
					"Blockquote",
					[]node.Block{
						&node.Leaf{"A", []node.Inline{
							&node.Uniform{"B", nil},
							&node.Escaped{"C", nil},
						}},
					},
				},
			},
			[]node.Node{
				&node.Walled{
					"Blockquote",
					[]node.Block{
						&node.Leaf{"A", []node.Inline{
							&node.Composite{
								"CA",
								&node.Uniform{"B", nil},
								&node.Escaped{"C", nil},
							},
						}},
					},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			composited := make([]node.Node, len(c.in))
			copy(composited, c.in)
			compositer := composite.Transformer{composite.Map{
				"B": {
					Name:             "CA",
					PrimaryElement:   "B",
					SecondaryElement: "C",
				},
			}}
			composited = compositer.Transform(composited)

			got := stringifier.Stringify(composited...)
			want := stringifier.Stringify(c.out...)

			if got != want {
				t.Errorf("\ngot\n%s\nwant\n%s", got, want)
			}
		})
	}
}
