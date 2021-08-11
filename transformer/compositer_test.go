package transformer_test

import (
	"github.com/touchmarine/to/config"
	"github.com/touchmarine/to/node"
	"github.com/touchmarine/to/stringifier"
	"github.com/touchmarine/to/transformer"
	"testing"
)

func TestCompositer(t *testing.T) {
	cases := []struct {
		name string
		in   []node.Node
		out  []node.Node
	}{
		{
			"basic",
			[]node.Node{
				&node.BasicBlock{"TextBlock", []node.Inline{
					&node.Uniform{"Group", nil},
					&node.Escaped{"Link", nil},
				}},
			},
			[]node.Node{
				&node.BasicBlock{"TextBlock", []node.Inline{
					&node.Composite{
						"NamedLink",
						&node.Uniform{"Group", nil},
						&node.Escaped{"Link", nil},
					},
				}},
			},
		},
		{
			"filled",
			[]node.Node{
				&node.BasicBlock{"TextBlock", []node.Inline{
					&node.Uniform{"Group", []node.Inline{
						node.Text("a"),
					}},
					&node.Escaped{"Link", []byte("b")},
				}},
			},
			[]node.Node{
				&node.BasicBlock{"TextBlock", []node.Inline{
					&node.Composite{
						"NamedLink",
						&node.Uniform{"Group", []node.Inline{
							node.Text("a"),
						}},
						&node.Escaped{"Link", []byte("b")},
					},
				}},
			},
		},
		{
			"nested",
			[]node.Node{
				&node.BasicBlock{"TextBlock", []node.Inline{
					&node.Uniform{"Group", []node.Inline{
						&node.Uniform{"Group", nil},
						&node.Escaped{"Link", nil},
					}},
					&node.Escaped{"Link", nil},
				}},
			},
			[]node.Node{
				&node.BasicBlock{"TextBlock", []node.Inline{
					&node.Composite{
						"NamedLink",
						&node.Uniform{"Group", []node.Inline{
							&node.Uniform{"Group", nil},
							&node.Escaped{"Link", nil},
						}},
						&node.Escaped{"Link", nil},
					},
				}},
			},
		},
		{
			"consecutive",
			[]node.Node{
				&node.BasicBlock{"TextBlock", []node.Inline{
					&node.Uniform{"Group", nil},
					&node.Escaped{"Link", nil},
					&node.Uniform{"Group", nil},
					&node.Escaped{"Link", nil},
				}},
			},
			[]node.Node{
				&node.BasicBlock{"TextBlock", []node.Inline{
					&node.Composite{
						"NamedLink",
						&node.Uniform{"Group", nil},
						&node.Escaped{"Link", nil},
					},
					&node.Composite{
						"NamedLink",
						&node.Uniform{"Group", nil},
						&node.Escaped{"Link", nil},
					},
				}},
			},
		},
		{
			"two lines",
			[]node.Node{
				&node.BasicBlock{"TextBlock", []node.Inline{
					&node.Uniform{"Group", nil},
					&node.Escaped{"Link", nil},
				}},
				&node.BasicBlock{"TextBlock", []node.Inline{
					&node.Uniform{"Group", nil},
					&node.Escaped{"Link", nil},
				}},
			},
			[]node.Node{
				&node.BasicBlock{"TextBlock", []node.Inline{
					&node.Composite{
						"NamedLink",
						&node.Uniform{"Group", nil},
						&node.Escaped{"Link", nil},
					},
				}},
				&node.BasicBlock{"TextBlock", []node.Inline{
					&node.Composite{
						"NamedLink",
						&node.Uniform{"Group", nil},
						&node.Escaped{"Link", nil},
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
						&node.BasicBlock{"TextBlock", []node.Inline{
							&node.Uniform{"Group", nil},
							&node.Escaped{"Link", nil},
						}},
					},
				},
			},
			[]node.Node{
				&node.Walled{
					"Blockquote",
					[]node.Block{
						&node.BasicBlock{"TextBlock", []node.Inline{
							&node.Composite{
								"NamedLink",
								&node.Uniform{"Group", nil},
								&node.Escaped{"Link", nil},
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
			composited = transformer.Composite(config.Default.Composites, composited)

			got := stringifier.Stringify(composited...)
			want := stringifier.Stringify(c.out...)

			if got != want {
				t.Errorf("\ngot\n%s\nwant\n%s", got, want)
			}
		})
	}
}
