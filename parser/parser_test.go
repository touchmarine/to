package parser_test

import (
	"github.com/touchmarine/to/node"
	"github.com/touchmarine/to/parser"
	"github.com/touchmarine/to/stringifier"
	"strings"
	"testing"
	"unicode"
)

func TestLine(t *testing.T) {
	cases := []struct {
		in  string
		out []node.Node
	}{
		{
			" ",
			[]node.Node{
				&node.Line{"Line", nil},
			},
		},
		{
			" \n ",
			[]node.Node{
				&node.Line{"Line", nil},
				&node.Line{"Line", nil},
			},
		},
		{
			"a",
			[]node.Node{
				&node.Line{"Line", []node.Inline{node.Text("a")}},
			},
		},
		{
			"a_*",
			[]node.Node{
				&node.Line{"Line", []node.Inline{node.Text("a_*")}},
			},
		},

		{
			"\na",
			[]node.Node{
				&node.Line{"Line", nil},
				&node.Line{"Line", []node.Inline{node.Text("a")}},
			},
		},
		{
			"\n\na",
			[]node.Node{
				&node.Line{"Line", nil},
				&node.Line{"Line", nil},
				&node.Line{"Line", []node.Inline{node.Text("a")}},
			},
		},
		{
			" \na",
			[]node.Node{
				&node.Line{"Line", nil},
				&node.Line{"Line", []node.Inline{node.Text("a")}},
			},
		},
		{
			"\t\na",
			[]node.Node{
				&node.Line{"Line", nil},
				&node.Line{"Line", []node.Inline{node.Text("a")}},
			},
		},
		{
			"a\n",
			[]node.Node{
				&node.Line{"Line", []node.Inline{node.Text("a")}},
			},
		},
		{
			"a\nb",
			[]node.Node{
				&node.Line{"Line", []node.Inline{node.Text("a")}},
				&node.Line{"Line", []node.Inline{node.Text("b")}},
			},
		},
		{
			"a\n\nb",
			[]node.Node{
				&node.Line{"Line", []node.Inline{node.Text("a")}},
				&node.Line{"Line", nil},
				&node.Line{"Line", []node.Inline{node.Text("b")}},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			test(t, c.in, c.out, nil)
		})
	}
}

func TestWalled(t *testing.T) {
	cases := []struct {
		in  string
		out []node.Node
	}{
		{
			">",
			[]node.Node{&node.Walled{"Blockquote", []node.Block{
				&node.Line{"Line", nil},
			}}},
		},
		{
			">\na",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Line{"Line", nil},
				}},
				&node.Line{"Line", []node.Inline{node.Text("a")}},
			},
		},
		{
			">>",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Walled{"Blockquote", []node.Block{
						&node.Line{"Line", nil},
					}},
				}},
			},
		},
		{
			">a",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
				}},
			},
		},
		{
			">\n>",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Line{"Line", nil},
					&node.Line{"Line", nil},
				}},
			},
		},
		{
			">a\n>b",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
					&node.Line{"Line", []node.Inline{node.Text("b")}},
				}},
			},
		},
		{
			">\n>>",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Line{"Line", nil},
					&node.Walled{"Blockquote", []node.Block{
						&node.Line{"Line", nil},
					}},
				}},
			},
		},
		{
			">a\n>>b",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
					&node.Walled{"Blockquote", []node.Block{
						&node.Line{"Line", []node.Inline{node.Text("b")}},
					}},
				}},
			},
		},
		{
			">>\n>>",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Walled{"Blockquote", []node.Block{
						&node.Line{"Line", nil},
						&node.Line{"Line", nil},
					}},
				}},
			},
		},
		{
			">>a\n>>b",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Walled{"Blockquote", []node.Block{
						&node.Line{"Line", []node.Inline{node.Text("a")}},
						&node.Line{"Line", []node.Inline{node.Text("b")}},
					}},
				}},
			},
		},
		{
			">>\n>",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Walled{"Blockquote", []node.Block{
						&node.Line{"Line", nil},
					}},
					&node.Line{"Line", nil},
				}},
			},
		},
		{
			">>a\n>b",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Walled{"Blockquote", []node.Block{
						&node.Line{"Line", []node.Inline{node.Text("a")}},
					}},
					&node.Line{"Line", []node.Inline{node.Text("b")}},
				}},
			},
		},
		{
			">\n\n>",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Line{"Line", nil},
				}},
				&node.Line{"Line", nil},
				&node.Walled{"Blockquote", []node.Block{
					&node.Line{"Line", nil},
				}},
			},
		},
		{
			">a\n\n>b",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
				}},
				&node.Line{"Line", nil},
				&node.Walled{"Blockquote", []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("b")}},
				}},
			},
		},

		// spacing
		{
			" >",
			[]node.Node{&node.Walled{"Blockquote", []node.Block{
				&node.Line{"Line", nil},
			}}},
		},

		// regression
		{
			"> >",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Walled{"Blockquote", []node.Block{
						&node.Line{"Line", nil},
					}},
				}},
			},
		},
		{
			">\t>",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Walled{"Blockquote", []node.Block{
						&node.Line{"Line", nil},
					}},
				}},
			},
		},
		{
			"> > >",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Walled{"Blockquote", []node.Block{
						&node.Walled{"Blockquote", []node.Block{
							&node.Line{"Line", nil},
						}},
					}},
				}},
			},
		},
		{
			">\n >",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Line{"Line", nil},
					&node.Line{"Line", nil},
				}},
			},
		},
		{
			" >\n>",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Line{"Line", nil},
					&node.Line{"Line", nil},
				}},
			},
		},
		{
			">a\n >b",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Line{"Line", []node.Inline{
						node.Text("a"),
					}},
					&node.Line{"Line", []node.Inline{
						node.Text("b"),
					}},
				}},
			},
		},

		{
			">\na",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Line{"Line", nil},
				}},
				&node.Line{"Line", []node.Inline{node.Text("a")}},
			},
		},
		{
			">\n>\na",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Line{"Line", nil},
					&node.Line{"Line", nil},
				}},
				&node.Line{"Line", []node.Inline{node.Text("a")}},
			},
		},
		{
			">\n>\n>\na",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Line{"Line", nil},
					&node.Line{"Line", nil},
					&node.Line{"Line", nil},
				}},
				&node.Line{"Line", []node.Inline{node.Text("a")}},
			},
		},
		{
			">\n>\n>\n>\na",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Line{"Line", nil},
					&node.Line{"Line", nil},
					&node.Line{"Line", nil},
					&node.Line{"Line", nil},
				}},
				&node.Line{"Line", []node.Inline{node.Text("a")}},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			test(t, c.in, c.out, nil)
		})
	}
}

func TestHanging(t *testing.T) {
	cases := []struct {
		in  string
		out []node.Node
	}{
		{
			"*",
			[]node.Node{&node.Hanging{"DescriptionList", 0, []node.Block{
				&node.Line{"Line", nil},
			}}},
		},
		{
			"**",
			[]node.Node{
				&node.Hanging{"DescriptionList", 0, []node.Block{
					&node.Hanging{"DescriptionList", 0, []node.Block{
						&node.Line{"Line", nil},
					}},
				}},
			},
		},
		{
			"*a",
			[]node.Node{
				&node.Hanging{"DescriptionList", 0, []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
				}},
			},
		},
		{
			"*\n*",
			[]node.Node{
				&node.Hanging{"DescriptionList", 0, []node.Block{
					&node.Line{"Line", nil},
				}},
				&node.Hanging{"DescriptionList", 0, []node.Block{
					&node.Line{"Line", nil},
				}},
			},
		},
		{
			"*\n\n*",
			[]node.Node{
				&node.Hanging{"DescriptionList", 0, []node.Block{
					&node.Line{"Line", nil},
				}},
				&node.Line{"Line", nil},
				&node.Hanging{"DescriptionList", 0, []node.Block{
					&node.Line{"Line", nil},
				}},
			},
		},
		{
			"*\n\n\n*",
			[]node.Node{
				&node.Hanging{"DescriptionList", 0, []node.Block{
					&node.Line{"Line", nil},
				}},
				&node.Line{"Line", nil},
				&node.Line{"Line", nil},
				&node.Hanging{"DescriptionList", 0, []node.Block{
					&node.Line{"Line", nil},
				}},
			},
		},
		{
			"*\n\t\n*",
			[]node.Node{
				&node.Hanging{"DescriptionList", 0, []node.Block{
					&node.Line{"Line", nil},
				}},
				&node.Line{"Line", nil},
				&node.Hanging{"DescriptionList", 0, []node.Block{
					&node.Line{"Line", nil},
				}},
			},
		},
		{
			"*a\n\n*b",
			[]node.Node{
				&node.Hanging{"DescriptionList", 0, []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
				}},
				&node.Line{"Line", nil},
				&node.Hanging{"DescriptionList", 0, []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("b")}},
				}},
			},
		},
		{
			"*a\nb",
			[]node.Node{
				&node.Hanging{"DescriptionList", 0, []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
				}},
				&node.Line{"Line", []node.Inline{node.Text("b")}},
			},
		},
		{
			"*a\n b",
			[]node.Node{
				&node.Hanging{"DescriptionList", 0, []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
					&node.Line{"Line", []node.Inline{node.Text("b")}},
				}},
			},
		},
		{
			"*a\n  b",
			[]node.Node{
				&node.Hanging{"DescriptionList", 0, []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
					&node.Line{"Line", []node.Inline{node.Text("b")}},
				}},
			},
		},

		// spacing
		{
			" *a\n b",
			[]node.Node{
				&node.Hanging{"DescriptionList", 0, []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
				}},
				&node.Line{"Line", []node.Inline{node.Text("b")}},
			},
		},
		{
			" *a\n  b",
			[]node.Node{
				&node.Hanging{"DescriptionList", 0, []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
					&node.Line{"Line", []node.Inline{node.Text("b")}},
				}},
			},
		},

		// nested
		{
			"*\n *",
			[]node.Node{
				&node.Hanging{"DescriptionList", 0, []node.Block{
					&node.Line{"Line", nil},
					&node.Hanging{"DescriptionList", 0, []node.Block{
						&node.Line{"Line", nil},
					}},
				}},
			},
		},
		{
			"*a\n *b",
			[]node.Node{
				&node.Hanging{"DescriptionList", 0, []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
					&node.Hanging{"DescriptionList", 0, []node.Block{
						&node.Line{"Line", []node.Inline{node.Text("b")}},
					}},
				}},
			},
		},
		{
			"**a\n b",
			[]node.Node{
				&node.Hanging{"DescriptionList", 0, []node.Block{
					&node.Hanging{"DescriptionList", 0, []node.Block{
						&node.Line{"Line", []node.Inline{node.Text("a")}},
					}},
					&node.Line{"Line", []node.Inline{node.Text("b")}},
				}},
			},
		},
		{
			"**a\n  b",
			[]node.Node{
				&node.Hanging{"DescriptionList", 0, []node.Block{
					&node.Hanging{"DescriptionList", 0, []node.Block{
						&node.Line{"Line", []node.Inline{node.Text("a")}},
						&node.Line{"Line", []node.Inline{node.Text("b")}},
					}},
				}},
			},
		},
		{
			"**a\n   b",
			[]node.Node{
				&node.Hanging{"DescriptionList", 0, []node.Block{
					&node.Hanging{"DescriptionList", 0, []node.Block{
						&node.Line{"Line", []node.Inline{node.Text("a")}},
						&node.Line{"Line", []node.Inline{node.Text("b")}},
					}},
				}},
			},
		},

		{
			">*",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Hanging{"DescriptionList", 0, []node.Block{
						&node.Line{"Line", nil},
					}},
				}},
			},
		},
		{
			">*\n*",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Hanging{"DescriptionList", 0, []node.Block{
						&node.Line{"Line", nil},
					}},
				}},
				&node.Hanging{"DescriptionList", 0, []node.Block{
					&node.Line{"Line", nil},
				}},
			},
		},
		{
			">*\n>*",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Hanging{"DescriptionList", 0, []node.Block{
						&node.Line{"Line", nil},
					}},
					&node.Hanging{"DescriptionList", 0, []node.Block{
						&node.Line{"Line", nil},
					}},
				}},
			},
		},
		{
			">*\n> *",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Hanging{"DescriptionList", 0, []node.Block{
						&node.Line{"Line", nil},
						&node.Hanging{"DescriptionList", 0, []node.Block{
							&node.Line{"Line", nil},
						}},
					}},
				}},
			},
		},
		{
			">*\n> a",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Hanging{"DescriptionList", 0, []node.Block{
						&node.Line{"Line", nil},
						&node.Line{"Line", []node.Inline{node.Text("a")}},
					}},
				}},
			},
		},
		{
			"> *\n>  a",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Hanging{"DescriptionList", 0, []node.Block{
						&node.Line{"Line", nil},
						&node.Line{"Line", []node.Inline{node.Text("a")}},
					}},
				}},
			},
		},

		// nested+spacing
		{
			" *\n *",
			[]node.Node{
				&node.Hanging{"DescriptionList", 0, []node.Block{
					&node.Line{"Line", nil},
				}},
				&node.Hanging{"DescriptionList", 0, []node.Block{
					&node.Line{"Line", nil},
				}},
			},
		},
		{
			" *\n  *",
			[]node.Node{
				&node.Hanging{"DescriptionList", 0, []node.Block{
					&node.Line{"Line", nil},
					&node.Hanging{"DescriptionList", 0, []node.Block{
						&node.Line{"Line", nil},
					}},
				}},
			},
		},

		// tab (equals 8 spaces in this regard)
		{
			"*a\n\tb",
			[]node.Node{
				&node.Hanging{"DescriptionList", 0, []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
					&node.Line{"Line", []node.Inline{node.Text("b")}},
				}},
			},
		},
		{
			"\t*a\n\tb",
			[]node.Node{
				&node.Hanging{"DescriptionList", 0, []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
				}},
				&node.Line{"Line", []node.Inline{node.Text("b")}},
			},
		},
		{
			"\t*a\n\t b",
			[]node.Node{
				&node.Hanging{"DescriptionList", 0, []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
					&node.Line{"Line", []node.Inline{node.Text("b")}},
				}},
			},
		},
		{
			"\t*a\n \tb",
			[]node.Node{
				&node.Hanging{"DescriptionList", 0, []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
					&node.Line{"Line", []node.Inline{node.Text("b")}},
				}},
			},
		},
		{
			"\t*a\n  \tb",
			[]node.Node{
				&node.Hanging{"DescriptionList", 0, []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
					&node.Line{"Line", []node.Inline{node.Text("b")}},
				}},
			},
		},

		{
			"\t\t*a\n                b",
			[]node.Node{
				&node.Hanging{"DescriptionList", 0, []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
				}},
				&node.Line{"Line", []node.Inline{node.Text("b")}},
			},
		},
		{
			"\t\t*a\n                 b",
			[]node.Node{
				&node.Hanging{"DescriptionList", 0, []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
					&node.Line{"Line", []node.Inline{node.Text("b")}},
				}},
			},
		},
		{
			"                *a\n\t\tb",
			[]node.Node{
				&node.Hanging{"DescriptionList", 0, []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
				}},
				&node.Line{"Line", []node.Inline{node.Text("b")}},
			},
		},
		{
			"               *a\n\t\tb",
			[]node.Node{
				&node.Hanging{"DescriptionList", 0, []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
					&node.Line{"Line", []node.Inline{node.Text("b")}},
				}},
			},
		},

		// nested+blank lines
		{
			"*\n *",
			[]node.Node{
				&node.Hanging{"DescriptionList", 0, []node.Block{
					&node.Line{"Line", nil},
					&node.Hanging{"DescriptionList", 0, []node.Block{
						&node.Line{"Line", nil},
					}},
				}},
			},
		},
		//*
		//
		// *
		{
			"*\n\n *",
			[]node.Node{
				&node.Hanging{"DescriptionList", 0, []node.Block{
					&node.Line{"Line", nil},
					&node.Line{"Line", nil},
					&node.Hanging{"DescriptionList", 0, []node.Block{
						&node.Line{"Line", nil},
					}},
				}},
			},
		},
		{
			"*\n \t\n *",
			[]node.Node{
				&node.Hanging{"DescriptionList", 0, []node.Block{
					&node.Line{"Line", nil},
					&node.Line{"Line", nil},
					&node.Hanging{"DescriptionList", 0, []node.Block{
						&node.Line{"Line", nil},
					}},
				}},
			},
		},

		{
			"**\n\n*",
			[]node.Node{
				&node.Hanging{"DescriptionList", 0, []node.Block{
					&node.Hanging{"DescriptionList", 0, []node.Block{
						&node.Line{"Line", nil},
					}},
				}},
				&node.Line{"Line", nil},
				&node.Hanging{"DescriptionList", 0, []node.Block{
					&node.Line{"Line", nil},
				}},
			},
		},
		{
			"a\n\n**\n\n*",
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					node.Text("a"),
				}},
				&node.Line{"Line", nil},
				&node.Hanging{"DescriptionList", 0, []node.Block{
					&node.Hanging{"DescriptionList", 0, []node.Block{
						&node.Line{"Line", nil},
					}},
				}},
				&node.Line{"Line", nil},
				&node.Hanging{"DescriptionList", 0, []node.Block{
					&node.Line{"Line", nil},
				}},
			},
		},

		{
			"**\n\na",
			[]node.Node{
				&node.Hanging{"DescriptionList", 0, []node.Block{
					&node.Hanging{"DescriptionList", 0, []node.Block{
						&node.Line{"Line", nil},
					}},
				}},
				&node.Line{"Line", nil},
				&node.Line{"Line", []node.Inline{
					node.Text("a"),
				}},
			},
		},
		{
			"**\n\n a",
			[]node.Node{
				&node.Hanging{"DescriptionList", 0, []node.Block{
					&node.Hanging{"DescriptionList", 0, []node.Block{
						&node.Line{"Line", nil},
					}},
					&node.Line{"Line", nil},
					&node.Line{"Line", []node.Inline{
						node.Text("a"),
					}},
				}},
			},
		},
		{
			"**\n\n  a",
			[]node.Node{
				&node.Hanging{"DescriptionList", 0, []node.Block{
					&node.Hanging{"DescriptionList", 0, []node.Block{
						&node.Line{"Line", nil},
						&node.Line{"Line", nil},
						&node.Line{"Line", []node.Inline{
							node.Text("a"),
						}},
					}},
				}},
			},
		},
		{
			"**\n\n  a\nb",
			[]node.Node{
				&node.Hanging{"DescriptionList", 0, []node.Block{
					&node.Hanging{"DescriptionList", 0, []node.Block{
						&node.Line{"Line", nil},
						&node.Line{"Line", nil},
						&node.Line{"Line", []node.Inline{
							node.Text("a"),
						}},
					}},
				}},
				&node.Line{"Line", []node.Inline{
					node.Text("b"),
				}},
			},
		},

		// regression
		{
			"*\n >b",
			[]node.Node{
				&node.Hanging{"DescriptionList", 0, []node.Block{
					&node.Line{"Line", nil},
					&node.Walled{"Blockquote", []node.Block{
						&node.Line{"Line", []node.Inline{
							node.Text("b"),
						}},
					}},
				}},
			},
		},
		{
			"*>a\n >b",
			[]node.Node{
				&node.Hanging{"DescriptionList", 0, []node.Block{
					&node.Walled{"Blockquote", []node.Block{
						&node.Line{"Line", []node.Inline{
							node.Text("a"),
						}},
						&node.Line{"Line", []node.Inline{
							node.Text("b"),
						}},
					}},
				}},
			},
		},

		//*
		//	*a
		//	 *b
		//	c
		{
			"*\n\t*a\n\t *b\n\tc",
			[]node.Node{&node.Hanging{"DescriptionList", 0, []node.Block{
				&node.Line{"Line", nil},
				&node.Hanging{"DescriptionList", 0, []node.Block{
					&node.Line{"Line", []node.Inline{
						node.Text("a"),
					}},
					&node.Hanging{"DescriptionList", 0, []node.Block{
						&node.Line{"Line", []node.Inline{
							node.Text("b"),
						}},
					}},
				}},
				&node.Line{"Line", []node.Inline{
					node.Text("c"),
				}},
			}}},
		},

		{
			"*\n  >*a",
			[]node.Node{&node.Hanging{"DescriptionList", 0, []node.Block{
				&node.Line{"Line", nil},
				&node.Walled{"Blockquote", []node.Block{
					&node.Hanging{"DescriptionList", 0, []node.Block{
						&node.Line{"Line", []node.Inline{
							node.Text("a"),
						}},
					}},
				}},
			}}},
		},
		{
			"*\n\t>*a",
			[]node.Node{&node.Hanging{"DescriptionList", 0, []node.Block{
				&node.Line{"Line", nil},
				&node.Walled{"Blockquote", []node.Block{
					&node.Hanging{"DescriptionList", 0, []node.Block{
						&node.Line{"Line", []node.Inline{
							node.Text("a"),
						}},
					}},
				}},
			}}},
		},
		//*
		//	>	*
		//	>
		{
			"*\n\t>\t*\n\t>",
			[]node.Node{&node.Hanging{"DescriptionList", 0, []node.Block{
				&node.Line{"Line", nil},
				&node.Walled{"Blockquote", []node.Block{
					&node.Hanging{"DescriptionList", 0, []node.Block{
						&node.Line{"Line", nil},
						&node.Line{"Line", nil},
					}},
				}},
			}}},
		},
		{
			"  >*a\n > *b",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Hanging{"DescriptionList", 0, []node.Block{
						&node.Line{"Line", []node.Inline{node.Text("a")}},
						&node.Hanging{"DescriptionList", 0, []node.Block{
							&node.Line{"Line", []node.Inline{node.Text("b")}},
						}},
					}},
				}},
			},
		},
		{
			"  > *a\n >  *b",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Hanging{"DescriptionList", 0, []node.Block{
						&node.Line{"Line", []node.Inline{node.Text("a")}},
						&node.Hanging{"DescriptionList", 0, []node.Block{
							&node.Line{"Line", []node.Inline{node.Text("b")}},
						}},
					}},
				}},
			},
		},

		//>*
		//>
		//> *
		{
			">*\n>\n> *",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Hanging{"DescriptionList", 0, []node.Block{
						&node.Line{"Line", nil},
						&node.Line{"Line", nil},
						&node.Hanging{"DescriptionList", 0, []node.Block{
							&node.Line{"Line", nil},
						}},
					}},
				}},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			test(t, c.in, c.out, nil)
		})
	}
}

func TestHangingRanked(t *testing.T) {
	cases := []struct {
		in  string
		out []node.Node
	}{
		{
			"=",
			[]node.Node{&node.Hanging{"Heading", 1, []node.Block{
				&node.Line{"Line", nil},
			}}},
		},
		{
			"==",
			[]node.Node{&node.Hanging{"Heading", 2, []node.Block{
				&node.Line{"Line", nil},
			}}},
		},
		{
			"= =",
			[]node.Node{
				&node.Hanging{"Heading", 1, []node.Block{
					&node.Hanging{"Heading", 1, []node.Block{
						&node.Line{"Line", nil},
					}},
				}},
			},
		},
		{
			"== ==",
			[]node.Node{
				&node.Hanging{"Heading", 2, []node.Block{
					&node.Hanging{"Heading", 2, []node.Block{
						&node.Line{"Line", nil},
					}},
				}},
			},
		},
		{
			"==a",
			[]node.Node{&node.Hanging{"Heading", 2, []node.Block{
				&node.Line{"Line", []node.Inline{
					node.Text("a"),
				}},
			}}},
		},
		{
			"==\n==",
			[]node.Node{
				&node.Hanging{"Heading", 2, []node.Block{
					&node.Line{"Line", nil},
				}},
				&node.Hanging{"Heading", 2, []node.Block{
					&node.Line{"Line", nil},
				}},
			},
		},

		{
			"==a\nb",
			[]node.Node{
				&node.Hanging{"Heading", 2, []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
				}},
				&node.Line{"Line", []node.Inline{node.Text("b")}},
			},
		},
		{
			"==a\n b",
			[]node.Node{
				&node.Hanging{"Heading", 2, []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
				}},
				&node.Line{"Line", []node.Inline{node.Text("b")}},
			},
		},
		{
			"==a\n  b",
			[]node.Node{
				&node.Hanging{"Heading", 2, []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
					&node.Line{"Line", []node.Inline{node.Text("b")}},
				}},
			},
		},
		{
			"==a\n   b",
			[]node.Node{
				&node.Hanging{"Heading", 2, []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
					&node.Line{"Line", []node.Inline{node.Text("b")}},
				}},
			},
		},

		// spacing
		{
			" ==a\n  b",
			[]node.Node{
				&node.Hanging{"Heading", 2, []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
				}},
				&node.Line{"Line", []node.Inline{node.Text("b")}},
			},
		},
		{
			" ==a\n   b",
			[]node.Node{
				&node.Hanging{"Heading", 2, []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
					&node.Line{"Line", []node.Inline{node.Text("b")}},
				}},
			},
		},

		// nested
		{
			"==\n  ==",
			[]node.Node{
				&node.Hanging{"Heading", 2, []node.Block{
					&node.Line{"Line", nil},
					&node.Hanging{"Heading", 2, []node.Block{
						&node.Line{"Line", nil},
					}},
				}},
			},
		},
		{
			"==a\n  ==b",
			[]node.Node{
				&node.Hanging{"Heading", 2, []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
					&node.Hanging{"Heading", 2, []node.Block{
						&node.Line{"Line", []node.Inline{node.Text("b")}},
					}},
				}},
			},
		},
		{
			"== ==a\n  b",
			[]node.Node{
				&node.Hanging{"Heading", 2, []node.Block{
					&node.Hanging{"Heading", 2, []node.Block{
						&node.Line{"Line", []node.Inline{node.Text("a")}},
					}},
					&node.Line{"Line", []node.Inline{node.Text("b")}},
				}},
			},
		},
		{
			"== ==a\n     b",
			[]node.Node{
				&node.Hanging{"Heading", 2, []node.Block{
					&node.Hanging{"Heading", 2, []node.Block{
						&node.Line{"Line", []node.Inline{node.Text("a")}},
						&node.Line{"Line", []node.Inline{node.Text("b")}},
					}},
				}},
			},
		},
		{
			"== ==a\n      b",
			[]node.Node{
				&node.Hanging{"Heading", 2, []node.Block{
					&node.Hanging{"Heading", 2, []node.Block{
						&node.Line{"Line", []node.Inline{node.Text("a")}},
						&node.Line{"Line", []node.Inline{node.Text("b")}},
					}},
				}},
			},
		},

		{
			">==",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Hanging{"Heading", 2, []node.Block{
						&node.Line{"Line", nil},
					}},
				}},
			},
		},
		{
			">==\n==",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Hanging{"Heading", 2, []node.Block{
						&node.Line{"Line", nil},
					}},
				}},
				&node.Hanging{"Heading", 2, []node.Block{
					&node.Line{"Line", nil},
				}},
			},
		},
		{
			">==\n>==",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Hanging{"Heading", 2, []node.Block{
						&node.Line{"Line", nil},
					}},
					&node.Hanging{"Heading", 2, []node.Block{
						&node.Line{"Line", nil},
					}},
				}},
			},
		},
		{
			">==\n>  ==",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Hanging{"Heading", 2, []node.Block{
						&node.Line{"Line", nil},
						&node.Hanging{"Heading", 2, []node.Block{
							&node.Line{"Line", nil},
						}},
					}},
				}},
			},
		},
		{
			">==\n>  a",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Hanging{"Heading", 2, []node.Block{
						&node.Line{"Line", nil},
						&node.Line{"Line", []node.Inline{node.Text("a")}},
					}},
				}},
			},
		},
		{
			"> ==\n>   a",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Hanging{"Heading", 2, []node.Block{
						&node.Line{"Line", nil},
						&node.Line{"Line", []node.Inline{node.Text("a")}},
					}},
				}},
			},
		},

		// nested+spacing
		{
			" ==\n ==",
			[]node.Node{
				&node.Hanging{"Heading", 2, []node.Block{
					&node.Line{"Line", nil},
				}},
				&node.Hanging{"Heading", 2, []node.Block{
					&node.Line{"Line", nil},
				}},
			},
		},
		{
			" ==\n   ==",
			[]node.Node{
				&node.Hanging{"Heading", 2, []node.Block{
					&node.Line{"Line", nil},
					&node.Hanging{"Heading", 2, []node.Block{
						&node.Line{"Line", nil},
					}},
				}},
			},
		},

		// tab (equals 8 spaces in this regard)
		{
			"==a\n\tb",
			[]node.Node{
				&node.Hanging{"Heading", 2, []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
					&node.Line{"Line", []node.Inline{node.Text("b")}},
				}},
			},
		},
		{
			"\t==a\n\tb",
			[]node.Node{
				&node.Hanging{"Heading", 2, []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
				}},
				&node.Line{"Line", []node.Inline{node.Text("b")}},
			},
		},
		{
			"\t==a\n\t  b",
			[]node.Node{
				&node.Hanging{"Heading", 2, []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
					&node.Line{"Line", []node.Inline{node.Text("b")}},
				}},
			},
		},
		{
			"\t==a\n  \tb",
			[]node.Node{
				&node.Hanging{"Heading", 2, []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
					&node.Line{"Line", []node.Inline{node.Text("b")}},
				}},
			},
		},
		{
			"\t==a\n   \tb",
			[]node.Node{
				&node.Hanging{"Heading", 2, []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
					&node.Line{"Line", []node.Inline{node.Text("b")}},
				}},
			},
		},

		{
			"\t\t==a\n                 b",
			[]node.Node{
				&node.Hanging{"Heading", 2, []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
				}},
				&node.Line{"Line", []node.Inline{node.Text("b")}},
			},
		},
		{
			"\t\t==a\n                  b",
			[]node.Node{
				&node.Hanging{"Heading", 2, []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
					&node.Line{"Line", []node.Inline{node.Text("b")}},
				}},
			},
		},
		{
			"                 ==a\n\t\tb",
			[]node.Node{
				&node.Hanging{"Heading", 2, []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
				}},
				&node.Line{"Line", []node.Inline{node.Text("b")}},
			},
		},
		{
			"              ==a\n\t\tb",
			[]node.Node{
				&node.Hanging{"Heading", 2, []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
					&node.Line{"Line", []node.Inline{node.Text("b")}},
				}},
			},
		},

		// regression
		{
			"==\n  >b",
			[]node.Node{
				&node.Hanging{"Heading", 2, []node.Block{
					&node.Line{"Line", nil},
					&node.Walled{"Blockquote", []node.Block{
						&node.Line{"Line", []node.Inline{
							node.Text("b"),
						}},
					}},
				}},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			test(t, c.in, c.out, nil)
		})
	}
}

func TestHangingMinRank(t *testing.T) {
	cases := []struct {
		in  string
		out []node.Node
	}{
		{
			"#",
			[]node.Node{&node.Line{"Line", []node.Inline{
				node.Text("#"),
			}}},
		},
		{
			"##",
			[]node.Node{&node.Hanging{"NumberedHeading", 2, []node.Block{
				&node.Line{"Line", nil},
			}}},
		},
		{
			"# #",
			[]node.Node{&node.Line{"Line", []node.Inline{
				node.Text("# #"),
			}}},
		},
		{
			"## ##",
			[]node.Node{
				&node.Hanging{"NumberedHeading", 2, []node.Block{
					&node.Hanging{"NumberedHeading", 2, []node.Block{
						&node.Line{"Line", nil},
					}},
				}},
			},
		},
		{
			"##a",
			[]node.Node{&node.Hanging{"NumberedHeading", 2, []node.Block{
				&node.Line{"Line", []node.Inline{
					node.Text("a"),
				}},
			}}},
		},
		{
			"##\n##",
			[]node.Node{
				&node.Hanging{"NumberedHeading", 2, []node.Block{
					&node.Line{"Line", nil},
				}},
				&node.Hanging{"NumberedHeading", 2, []node.Block{
					&node.Line{"Line", nil},
				}},
			},
		},

		{
			"##a\nb",
			[]node.Node{
				&node.Hanging{"NumberedHeading", 2, []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
				}},
				&node.Line{"Line", []node.Inline{node.Text("b")}},
			},
		},

		// nested
		{
			"##\n  ##",
			[]node.Node{
				&node.Hanging{"NumberedHeading", 2, []node.Block{
					&node.Line{"Line", nil},
					&node.Hanging{"NumberedHeading", 2, []node.Block{
						&node.Line{"Line", nil},
					}},
				}},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			test(t, c.in, c.out, nil)
		})
	}
}

func TestHangingVerbatim(t *testing.T) {
	cases := []struct {
		in  string
		out []node.Node
	}{
		{
			".image",
			[]node.Node{&node.HangingVerbatim{"Image", 0, [][]byte{
				[]byte(""),
			}}},
		},
		{
			".image.image",
			[]node.Node{
				&node.HangingVerbatim{"Image", 0, [][]byte{
					[]byte(".image"),
				}},
			},
		},
		{
			".imagea",
			[]node.Node{
				&node.HangingVerbatim{"Image", 0, [][]byte{
					[]byte("a"),
				}},
			},
		},
		{
			".image\n.image",
			[]node.Node{
				&node.HangingVerbatim{"Image", 0, nil},
				&node.HangingVerbatim{"Image", 0, nil},
			},
		},
		{
			".image\n\n.image",
			[]node.Node{
				&node.HangingVerbatim{"Image", 0, [][]byte{
					[]byte(""),
				}},
				&node.Line{"Line", nil},
				&node.HangingVerbatim{"Image", 0, [][]byte{
					[]byte(""),
				}},
			},
		},
		{
			".image\n\n\n.image",
			[]node.Node{
				&node.HangingVerbatim{"Image", 0, [][]byte{
					[]byte(""),
				}},
				&node.Line{"Line", nil},
				&node.Line{"Line", nil},
				&node.HangingVerbatim{"Image", 0, [][]byte{
					[]byte(""),
				}},
			},
		},
		{
			".imagea\n\n.imageb",
			[]node.Node{
				&node.HangingVerbatim{"Image", 0, [][]byte{
					[]byte("a"),
				}},
				&node.Line{"Line", nil},
				&node.HangingVerbatim{"Image", 0, [][]byte{
					[]byte("b"),
				}},
			},
		},
		{
			".imagea\nb",
			[]node.Node{
				&node.HangingVerbatim{"Image", 0, [][]byte{
					[]byte("a"),
				}},
				&node.Line{"Line", []node.Inline{node.Text("b")}},
			},
		},
		{
			".imagea\n      b",
			[]node.Node{
				&node.HangingVerbatim{"Image", 0, [][]byte{
					[]byte("a"),
					[]byte("b"),
				}},
			},
		},
		{
			".imagea\n       b",
			[]node.Node{
				&node.HangingVerbatim{"Image", 0, [][]byte{
					[]byte("a"),
					[]byte("b"),
				}},
			},
		},
		{
			".image a",
			[]node.Node{
				&node.HangingVerbatim{"Image", 0, [][]byte{
					[]byte(" a"),
				}},
			},
		},
		{
			".imagea ",
			[]node.Node{
				&node.HangingVerbatim{"Image", 0, [][]byte{
					[]byte("a "),
				}},
			},
		},

		// spacing
		{
			" .imagea\n b",
			[]node.Node{
				&node.HangingVerbatim{"Image", 0, [][]byte{
					[]byte("a"),
				}},
				&node.Line{"Line", []node.Inline{node.Text("b")}},
			},
		},

		// nested
		{
			".image\n      .image",
			[]node.Node{
				&node.HangingVerbatim{"Image", 0, [][]byte{
					nil,
					[]byte(".image"),
				}},
			},
		},
		{
			".imagea\n      .imageb",
			[]node.Node{
				&node.HangingVerbatim{"Image", 0, [][]byte{
					[]byte("a"),
					[]byte(".imageb"),
				}},
			},
		},

		{
			">.image",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.HangingVerbatim{"Image", 0, nil},
				}},
			},
		},
		{
			">.image\n.image",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.HangingVerbatim{"Image", 0, nil},
				}},
				&node.HangingVerbatim{"Image", 0, nil},
			},
		},
		{
			">.image\n>.image",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.HangingVerbatim{"Image", 0, nil},
					&node.HangingVerbatim{"Image", 0, nil},
				}},
			},
		},
		{
			">.image\n>      .image",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.HangingVerbatim{"Image", 0, [][]byte{
						nil,
						[]byte(".image"),
					}},
				}},
			},
		},
		{
			">.image\n>      a",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.HangingVerbatim{"Image", 0, [][]byte{
						nil,
						[]byte("a"),
					}},
				}},
			},
		},
		{
			">.image\n>       a",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.HangingVerbatim{"Image", 0, [][]byte{
						nil,
						[]byte("a"),
					}},
				}},
			},
		},
		{
			"> .image\n>       a",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.HangingVerbatim{"Image", 0, [][]byte{
						nil,
						[]byte("a"),
					}},
				}},
			},
		},

		// nested+spacing
		{
			" .image\n .image",
			[]node.Node{
				&node.HangingVerbatim{"Image", 0, nil},
				&node.HangingVerbatim{"Image", 0, nil},
			},
		},
		{
			" .image\n       .image",
			[]node.Node{
				&node.HangingVerbatim{"Image", 0, [][]byte{
					nil,
					[]byte(".image"),
				}},
			},
		},

		// nested+blank lines
		{
			".image\n      .image",
			[]node.Node{
				&node.HangingVerbatim{"Image", 0, [][]byte{
					nil,
					[]byte(".image"),
				}},
			},
		},
		{
			".image\n\n      .image",
			[]node.Node{
				&node.HangingVerbatim{"Image", 0, [][]byte{
					nil,
					[]byte(""),
					[]byte(".image"),
				}},
			},
		},
		{
			".image\n \t\n      .image",
			[]node.Node{
				&node.HangingVerbatim{"Image", 0, [][]byte{
					nil,
					[]byte(""),
					[]byte(".image"),
				}},
			},
		},
		{
			".imagea\n\tb\n\n      c\n\td",
			[]node.Node{
				&node.HangingVerbatim{"Image", 0, [][]byte{
					[]byte("a"),
					[]byte("b"),
					[]byte(""),
					[]byte("c"),
					[]byte("d"),
				}},
			},
		},

		{
			".image.image\n\na",
			[]node.Node{
				&node.HangingVerbatim{"Image", 0, [][]byte{
					[]byte(".image"),
				}},
				&node.Line{"Line", nil},
				&node.Line{"Line", []node.Inline{
					node.Text("a"),
				}},
			},
		},
		{
			".image.image\n\n      a",
			[]node.Node{
				&node.HangingVerbatim{"Image", 0, [][]byte{
					[]byte(".image"),
					[]byte(""),
					[]byte("a"),
				}},
			},
		},
		{
			".image.image\n\n      a\nb",
			[]node.Node{
				&node.HangingVerbatim{"Image", 0, [][]byte{
					[]byte(".image"),
					[]byte(""),
					[]byte("a"),
				}},
				&node.Line{"Line", []node.Inline{
					node.Text("b"),
				}},
			},
		},

		//>.image
		//>
		//>.image
		{
			">.image\n>\n>.image",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.HangingVerbatim{"Image", 0, [][]byte{nil}},
					&node.Line{"Line", nil},
					&node.HangingVerbatim{"Image", 0, [][]byte{nil}},
				}},
			},
		},
		//>.image
		//>
		//>      .image
		{
			">.image\n>\n>      .image",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.HangingVerbatim{"Image", 0, [][]byte{
						nil,
						nil,
						[]byte(".image"),
					}},
				}},
			},
		},

		// regression
		{
			".image\n      >b",
			[]node.Node{
				&node.HangingVerbatim{"Image", 0, [][]byte{
					nil,
					[]byte(">b"),
				}},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			test(t, c.in, c.out, nil)
		})
	}
}

func TestHangingMulti(t *testing.T) {
	cases := []struct {
		in  string
		out []node.Node
	}{
		{
			"1",
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					node.Text("1"),
				}},
			},
		},
		{
			"1.",
			[]node.Node{&node.Hanging{"NumberedListItemDot", 0, []node.Block{
				&node.Line{"Line", nil},
			}}},
		},
		{
			"1.1.",
			[]node.Node{
				&node.Hanging{"NumberedListItemDot", 0, []node.Block{
					&node.Hanging{"NumberedListItemDot", 0, []node.Block{
						&node.Line{"Line", nil},
					}},
				}},
			},
		},
		{
			"1.a",
			[]node.Node{
				&node.Hanging{"NumberedListItemDot", 0, []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
				}},
			},
		},
		{
			"1.\n1.",
			[]node.Node{
				&node.Hanging{"NumberedListItemDot", 0, []node.Block{
					&node.Line{"Line", nil},
				}},
				&node.Hanging{"NumberedListItemDot", 0, []node.Block{
					&node.Line{"Line", nil},
				}},
			},
		},
		{
			"1.a\nb",
			[]node.Node{
				&node.Hanging{"NumberedListItemDot", 0, []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
				}},
				&node.Line{"Line", []node.Inline{node.Text("b")}},
			},
		},
		{
			"1.a\n b",
			[]node.Node{
				&node.Hanging{"NumberedListItemDot", 0, []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
				}},
				&node.Line{"Line", []node.Inline{node.Text("b")}},
			},
		},
		{
			"1.a\n  b",
			[]node.Node{
				&node.Hanging{"NumberedListItemDot", 0, []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
					&node.Line{"Line", []node.Inline{node.Text("b")}},
				}},
			},
		},

		// spacing
		{
			" 1.a\n b",
			[]node.Node{
				&node.Hanging{"NumberedListItemDot", 0, []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
				}},
				&node.Line{"Line", []node.Inline{node.Text("b")}},
			},
		},
		{
			" 1.a\n  b",
			[]node.Node{
				&node.Hanging{"NumberedListItemDot", 0, []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
				}},
				&node.Line{"Line", []node.Inline{node.Text("b")}},
			},
		},
		{
			" 1.a\n   b",
			[]node.Node{
				&node.Hanging{"NumberedListItemDot", 0, []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
					&node.Line{"Line", []node.Inline{node.Text("b")}},
				}},
			},
		},

		// nested
		{
			"1.\n  1.",
			[]node.Node{
				&node.Hanging{"NumberedListItemDot", 0, []node.Block{
					&node.Line{"Line", nil},
					&node.Hanging{"NumberedListItemDot", 0, []node.Block{
						&node.Line{"Line", nil},
					}},
				}},
			},
		},
		{
			"1.a\n  1.b",
			[]node.Node{
				&node.Hanging{"NumberedListItemDot", 0, []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
					&node.Hanging{"NumberedListItemDot", 0, []node.Block{
						&node.Line{"Line", []node.Inline{node.Text("b")}},
					}},
				}},
			},
		},

		{
			">1.",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Hanging{"NumberedListItemDot", 0, []node.Block{
						&node.Line{"Line", nil},
					}},
				}},
			},
		},
		{
			">1.\n1.",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Hanging{"NumberedListItemDot", 0, []node.Block{
						&node.Line{"Line", nil},
					}},
				}},
				&node.Hanging{"NumberedListItemDot", 0, []node.Block{
					&node.Line{"Line", nil},
				}},
			},
		},
		{
			">1.\n>1.",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Hanging{"NumberedListItemDot", 0, []node.Block{
						&node.Line{"Line", nil},
					}},
					&node.Hanging{"NumberedListItemDot", 0, []node.Block{
						&node.Line{"Line", nil},
					}},
				}},
			},
		},
		{
			">1.\n>  1.",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Hanging{"NumberedListItemDot", 0, []node.Block{
						&node.Line{"Line", nil},
						&node.Hanging{"NumberedListItemDot", 0, []node.Block{
							&node.Line{"Line", nil},
						}},
					}},
				}},
			},
		},
		{
			">1.\n>  a",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Hanging{"NumberedListItemDot", 0, []node.Block{
						&node.Line{"Line", nil},
						&node.Line{"Line", []node.Inline{node.Text("a")}},
					}},
				}},
			},
		},
		{
			"> 1.\n>   a",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Hanging{"NumberedListItemDot", 0, []node.Block{
						&node.Line{"Line", nil},
						&node.Line{"Line", []node.Inline{node.Text("a")}},
					}},
				}},
			},
		},

		{
			"1.-\n\n1.",
			[]node.Node{
				&node.Hanging{"NumberedListItemDot", 0, []node.Block{
					&node.Hanging{"ListItemDot", 0, []node.Block{
						&node.Line{"Line", nil},
					}},
				}},
				&node.Line{"Line", nil},
				&node.Hanging{"NumberedListItemDot", 0, []node.Block{
					&node.Line{"Line", nil},
				}},
			},
		},

		// regression
		{
			"1.\n  >b",
			[]node.Node{
				&node.Hanging{"NumberedListItemDot", 0, []node.Block{
					&node.Line{"Line", nil},
					&node.Walled{"Blockquote", []node.Block{
						&node.Line{"Line", []node.Inline{
							node.Text("b"),
						}},
					}},
				}},
			},
		},
		{
			"1.>a\n  >b",
			[]node.Node{
				&node.Hanging{"NumberedListItemDot", 0, []node.Block{
					&node.Walled{"Blockquote", []node.Block{
						&node.Line{"Line", []node.Inline{
							node.Text("a"),
						}},
						&node.Line{"Line", []node.Inline{
							node.Text("b"),
						}},
					}},
				}},
			},
		},
		{
			"1. >a\n   >b",
			[]node.Node{
				&node.Hanging{"NumberedListItemDot", 0, []node.Block{
					&node.Walled{"Blockquote", []node.Block{
						&node.Line{"Line", []node.Inline{
							node.Text("a"),
						}},
						&node.Line{"Line", []node.Inline{
							node.Text("b"),
						}},
					}},
				}},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			test(t, c.in, c.out, nil)
		})
	}
}

func TestFenced(t *testing.T) {
	cases := []struct {
		in  string
		out []node.Node
	}{
		{
			"``",
			[]node.Node{&node.Fenced{"CodeBlock", nil, nil}},
		},
		{
			"``a",
			[]node.Node{
				&node.Fenced{"CodeBlock", [][]byte{[]byte("a")}, nil},
			},
		},
		{
			"``a``",
			[]node.Node{
				&node.Fenced{"CodeBlock", [][]byte{[]byte("a``")}, nil},
			},
		},
		{
			"``\na",
			[]node.Node{
				&node.Fenced{"CodeBlock", [][]byte{nil, []byte("a")}, nil},
			},
		},
		{
			"``\n a",
			[]node.Node{
				&node.Fenced{"CodeBlock", [][]byte{nil, []byte(" a")}, nil},
			},
		},
		{
			"``\n\na",
			[]node.Node{
				&node.Fenced{
					"CodeBlock",
					[][]byte{
						nil,
						nil,
						[]byte("a")},
					nil,
				},
			},
		},
		{
			"````",
			[]node.Node{
				&node.Fenced{"CodeBlock", nil, nil},
			},
		},
		{
			"``\n``",
			[]node.Node{
				&node.Fenced{"CodeBlock", nil, nil},
			},
		},
		{
			"```\n```",
			[]node.Node{
				&node.Fenced{"CodeBlock", nil, nil},
			},
		},
		{
			"```\n``\n```",
			[]node.Node{
				&node.Fenced{"CodeBlock", [][]byte{nil, []byte("``")}, nil},
			},
		},
		{
			"```\n`````",
			[]node.Node{
				&node.Fenced{"CodeBlock", nil, []byte("``")},
			},
		},
		{
			"``\n``a",
			[]node.Node{
				&node.Fenced{"CodeBlock", nil, []byte("a")},
			},
		},

		// nesting
		{
			"``\n>",
			[]node.Node{
				&node.Fenced{"CodeBlock", [][]byte{nil, []byte(">")}, nil},
			},
		},
		{
			">``\na",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Fenced{"CodeBlock", nil, nil},
				}},
				&node.Line{"Line", []node.Inline{node.Text("a")}},
			},
		},
		{
			">``\n>a",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Fenced{"CodeBlock", [][]byte{nil, []byte("a")}, nil},
				}},
			},
		},
		{
			">``\n>a\n>``",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Fenced{"CodeBlock", [][]byte{nil, []byte("a")}, nil},
				}},
			},
		},
		{
			">``\n>a\n>``b",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Fenced{"CodeBlock", [][]byte{nil, []byte("a")}, []byte("b")},
				}},
			},
		},
		{
			">``\n>a\n>``\nb",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Fenced{"CodeBlock", [][]byte{nil, []byte("a")}, nil},
				}},
				&node.Line{"Line", []node.Inline{node.Text("b")}},
			},
		},

		// nesting+spacing
		{
			"> ``\n>a",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Fenced{"CodeBlock", [][]byte{nil, []byte("a")}, nil},
				}},
			},
		},
		{
			"> ``\n> a",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Fenced{"CodeBlock", [][]byte{nil, []byte("a")}, nil},
				}},
			},
		},
		{
			"> ``\n>  a",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Fenced{"CodeBlock", [][]byte{nil, []byte(" a")}, nil},
				}},
			},
		},
		{
			">  ``\n>  a",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Fenced{"CodeBlock", [][]byte{nil, []byte("a")}, nil},
				}},
			},
		},

		// tab
		{
			">\t``\n>a",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Fenced{"CodeBlock", [][]byte{nil, []byte("a")}, nil},
				}},
			},
		},
		{
			">\t``\n>        a",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Fenced{"CodeBlock", [][]byte{nil, []byte("a")}, nil},
				}},
			},
		},
		{
			">\t``\n>         a",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Fenced{"CodeBlock", [][]byte{nil, []byte(" a")}, nil},
				}},
			},
		},
		{
			">\t``\n>            a",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Fenced{"CodeBlock", [][]byte{nil, []byte("    a")}, nil},
				}},
			},
		},
		{
			"> ``\n>\ta",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Fenced{"CodeBlock", [][]byte{nil, []byte("       a")}, nil},
				}},
			},
		},
		{
			"> \t``\n>\t a",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Fenced{"CodeBlock", [][]byte{nil, []byte("a")}, nil},
				}},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			test(t, c.in, c.out, nil)
		})
	}
}

func TestSpacing(t *testing.T) {
	cases := []struct {
		in  string
		out []node.Node
	}{
		{
			"\n>",
			[]node.Node{
				&node.Line{"Line", nil},
				&node.Walled{"Blockquote", []node.Block{
					&node.Line{"Line", nil},
				}},
			},
		},

		// space
		{
			" >",
			[]node.Node{&node.Walled{"Blockquote", []node.Block{
				&node.Line{"Line", nil},
			}}},
		},
		{
			"> >",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Walled{"Blockquote", []node.Block{
						&node.Line{"Line", nil},
					}},
				}},
			},
		},
		{
			">  >",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Walled{"Blockquote", []node.Block{
						&node.Line{"Line", nil},
					}},
				}},
			},
		},
		{
			"> ",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Line{"Line", nil},
				}},
			},
		},
		{
			"> a",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
				}},
			},
		},

		// tab
		{
			"\t>",
			[]node.Node{&node.Walled{"Blockquote", []node.Block{
				&node.Line{"Line", nil},
			}}},
		},
		{
			">\t>",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Walled{"Blockquote", []node.Block{
						&node.Line{"Line", nil},
					}},
				}},
			},
		},
		{
			">\t\t>",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Walled{"Blockquote", []node.Block{
						&node.Line{"Line", nil},
					}},
				}},
			},
		},
		{
			">\t",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Line{"Line", nil},
				}},
			},
		},
		{
			">\ta",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
				}},
			},
		},

		// space+tab
		{
			" \t>",
			[]node.Node{&node.Walled{"Blockquote", []node.Block{
				&node.Line{"Line", nil},
			}}},
		},
		{
			"> \t>",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Walled{"Blockquote", []node.Block{
						&node.Line{"Line", nil},
					}},
				}},
			},
		},
		{
			">  \t>",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Walled{"Blockquote", []node.Block{
						&node.Line{"Line", nil},
					}},
				}},
			},
		},
		{
			"> \t",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Line{"Line", nil},
				}},
			},
		},
		{
			"> \ta",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
				}},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			test(t, c.in, c.out, nil)
		})
	}
}

func TestUniform(t *testing.T) {
	cases := []struct {
		in  string
		out []node.Node
	}{
		{
			"__",
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					&node.Uniform{"Emphasis", nil},
				}},
			},
		},
		{
			"____",
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					&node.Uniform{"Emphasis", nil},
				}},
			},
		},
		{
			"__a",
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					&node.Uniform{"Emphasis", []node.Inline{
						node.Text("a"),
					}},
				}},
			},
		},
		{
			"__a__",
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					&node.Uniform{"Emphasis", []node.Inline{
						node.Text("a"),
					}},
				}},
			},
		},
		{
			"__a__b",
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					&node.Uniform{"Emphasis", []node.Inline{
						node.Text("a"),
					}},
					node.Text("b"),
				}},
			},
		},

		// nested
		{
			"__**",
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					&node.Uniform{"Emphasis", []node.Inline{
						&node.Uniform{"Strong", nil},
					}},
				}},
			},
		},
		{
			"__**a",
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					&node.Uniform{"Emphasis", []node.Inline{
						&node.Uniform{"Strong", []node.Inline{
							node.Text("a"),
						}},
					}},
				}},
			},
		},
		{
			"__**a**b",
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					&node.Uniform{"Emphasis", []node.Inline{
						&node.Uniform{"Strong", []node.Inline{
							node.Text("a"),
						}},
						node.Text("b"),
					}},
				}},
			},
		},
		{
			"__**a__b",
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					&node.Uniform{"Emphasis", []node.Inline{
						&node.Uniform{"Strong", []node.Inline{
							node.Text("a"),
						}},
					}},
					node.Text("b"),
				}},
			},
		},
		{
			"__**a**b__c",
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					&node.Uniform{"Emphasis", []node.Inline{
						&node.Uniform{"Strong", []node.Inline{
							node.Text("a"),
						}},
						node.Text("b"),
					}},
					node.Text("c"),
				}},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			test(t, c.in, c.out, nil)
		})
	}
}

func TestEscaped(t *testing.T) {
	cases := []struct {
		in  string
		out []node.Node
	}{
		{
			"a``",
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					node.Text("a"),
					&node.Escaped{"Code", nil},
				}},
			},
		},
		{
			"a````",
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					node.Text("a"),
					&node.Escaped{"Code", nil},
				}},
			},
		},
		{
			"a`()`",
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					node.Text("a"),
					&node.Escaped{"Code", nil},
				}},
			},
		},
		{
			"a``b",
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					node.Text("a"),
					&node.Escaped{"Code", []byte("b")},
				}},
			},
		},
		{
			"a``b``",
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					node.Text("a"),
					&node.Escaped{"Code", []byte("b")},
				}},
			},
		},

		{
			"a`(",
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					node.Text("a"),
					&node.Escaped{"Code", nil},
				}},
			},
		},
		{
			"a`(b)`",
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					node.Text("a"),
					&node.Escaped{"Code", []byte("b")},
				}},
			},
		},
		{
			"a`)",
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					node.Text("a"),
					&node.Escaped{"Code", nil},
				}},
			},
		},
		{
			"a`)b(`",
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					node.Text("a"),
					&node.Escaped{"Code", []byte("b")},
				}},
			},
		},
		{
			"a`<",
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					node.Text("a"),
					&node.Escaped{"Code", nil},
				}},
			},
		},
		{
			"a`[",
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					node.Text("a"),
					&node.Escaped{"Code", nil},
				}},
			},
		},

		// nested elements are not allowed
		{
			"a`(``)`",
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					node.Text("a"),
					&node.Escaped{"Code", []byte("``")},
				}},
			},
		},
		{
			"a``__``",
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					node.Text("a"),
					&node.Escaped{"Code", []byte("__")},
				}},
			},
		},
		{
			"a__``__b``c",
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					node.Text("a"),
					&node.Uniform{"Emphasis", []node.Inline{
						&node.Escaped{"Code", []byte("__b")},
						node.Text("c"),
					}},
				}},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			test(t, c.in, c.out, nil)
		})
	}
}

func TestHat(t *testing.T) {
	cases := []struct {
		in  string
		out []node.Node
	}{
		{
			"%a",
			[]node.Node{
				&node.Hat{
					[][]byte{[]byte("a")},
					nil,
				},
			},
		},

		{
			"%a\nb",
			[]node.Node{
				&node.Hat{
					[][]byte{[]byte("a")},
					&node.Line{"Line", []node.Inline{node.Text("b")}},
				},
			},
		},
		{
			"%a\nb\n%c",
			[]node.Node{
				&node.Hat{
					[][]byte{[]byte("a")},
					&node.Line{"Line", []node.Inline{node.Text("b")}},
				},
				&node.Hat{
					[][]byte{[]byte("c")},
					nil,
				},
			},
		},
		{
			"%a\n%b\nc",
			[]node.Node{
				&node.Hat{
					[][]byte{[]byte("a"), []byte("b")},
					&node.Line{"Line", []node.Inline{node.Text("c")}},
				},
			},
		},
		{
			"%a\n%\nc",
			[]node.Node{
				&node.Hat{
					[][]byte{[]byte("a"), []byte("")},
					&node.Line{"Line", []node.Inline{node.Text("c")}},
				},
			},
		},

		{
			"%a\n>",
			[]node.Node{
				&node.Hat{
					[][]byte{[]byte("a")},
					&node.Walled{"Blockquote", []node.Block{
						&node.Line{"Line", nil},
					}},
				},
			},
		},
		{
			"%a\n*",
			[]node.Node{
				&node.Hat{
					[][]byte{[]byte("a")},
					&node.Hanging{"DescriptionList", 0, []node.Block{
						&node.Line{"Line", nil},
					}},
				},
			},
		},

		{
			">%a\n>b",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Hat{
						[][]byte{[]byte("a")},
						&node.Line{"Line", []node.Inline{node.Text("b")}},
					},
				}},
			},
		},
		{
			"*%a\n b",
			[]node.Node{
				&node.Hanging{"DescriptionList", 0, []node.Block{
					&node.Hat{
						[][]byte{[]byte("a")},
						&node.Line{"Line", []node.Inline{node.Text("b")}},
					},
				}},
			},
		},

		{
			">%a\nb",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Hat{
						[][]byte{[]byte("a")},
						nil,
					},
				}},
				&node.Line{"Line", []node.Inline{node.Text("b")}},
			},
		},
		{
			"*%a\nb",
			[]node.Node{
				&node.Hanging{"DescriptionList", 0, []node.Block{
					&node.Hat{
						[][]byte{[]byte("a")},
						nil,
					},
				}},
				&node.Line{"Line", []node.Inline{node.Text("b")}},
			},
		},
		//*%a
		//
		// b
		{
			"*%a\n\n b",
			[]node.Node{
				&node.Hanging{"DescriptionList", 0, []node.Block{
					&node.Hat{
						[][]byte{[]byte("a")},
						&node.Line{"Line", nil},
					},
					&node.Line{"Line", []node.Inline{node.Text("b")}},
				}},
			},
		},
		//>*%a
		//>
		//> b
		{
			">*%a\n>\n> b",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Hanging{"DescriptionList", 0, []node.Block{
						&node.Hat{
							[][]byte{[]byte("a")},
							&node.Line{"Line", nil},
						},
						&node.Line{"Line", []node.Inline{node.Text("b")}},
					}},
				}},
			},
		},

		{
			"%a\n\nb",
			[]node.Node{
				&node.Hat{
					[][]byte{[]byte("a")},
					&node.Line{"Line", nil},
				},
				&node.Line{"Line", []node.Inline{node.Text("b")}},
			},
		},

		//*
		// %a
		//
		// b
		{
			"*\n %a\n\n b",
			[]node.Node{
				&node.Hanging{"DescriptionList", 0, []node.Block{
					&node.Line{"Line", nil},
					&node.Hat{
						[][]byte{[]byte("a")},
						&node.Line{"Line", nil},
					},
					&node.Line{"Line", []node.Inline{node.Text("b")}},
				}},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			test(t, c.in, c.out, nil)
		})
	}
}

func TestBlockEscape(t *testing.T) {
	cases := []struct {
		in  string
		out []node.Node
	}{
		{
			`\`,
			[]node.Node{
				&node.Line{"Line", nil},
			},
		},
		{
			`\\`,
			[]node.Node{
				&node.Line{"Line", []node.Inline{node.Text(`\`)}},
			},
		},
		{
			`\a`,
			[]node.Node{
				&node.Line{"Line", []node.Inline{node.Text("a")}},
			},
		},
		{
			`\>`,
			[]node.Node{
				&node.Line{"Line", []node.Inline{node.Text(">")}},
			},
		},
		{
			`\\>`,
			[]node.Node{
				&node.Line{"Line", []node.Inline{node.Text(`\>`)}},
			},
		},
		{
			"\\\n\\",
			[]node.Node{
				&node.Line{"Line", nil},
				&node.Line{"Line", nil},
			},
		},
		{
			"\\a\n\\b",
			[]node.Node{
				&node.Line{"Line", []node.Inline{node.Text("a")}},
				&node.Line{"Line", []node.Inline{node.Text("b")}},
			},
		},
		{
			"\\``",
			[]node.Node{
				&node.Line{"Line", []node.Inline{&node.Escaped{"Code", nil}}},
			},
		},
		{
			"\\\\``",
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					node.Text("``"),
				}},
			},
		},

		{
			"``\\**\n\\**",
			[]node.Node{
				&node.Fenced{"CodeBlock", [][]byte{[]byte("\\**"), []byte("\\**")}, nil},
			},
		},

		// verbatim elements
		{
			`\.image`,
			[]node.Node{
				&node.Line{"Line", []node.Inline{node.Text(".image")}},
			},
		},
		{
			`.image\.image`,
			[]node.Node{
				&node.HangingVerbatim{"Image", 0, [][]byte{
					[]byte(`\.image`),
				}},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			test(t, c.in, c.out, nil)
		})
	}
}

func TestInlineEscape(t *testing.T) {
	cases := []struct {
		in  string
		out []node.Node
	}{
		{
			`\`,
			[]node.Node{
				&node.Line{"Line", nil},
			},
		},
		{
			`\\`,
			[]node.Node{
				&node.Line{"Line", []node.Inline{node.Text(`\`)}},
			},
		},
		{
			`\\a`,
			[]node.Node{
				&node.Line{"Line", []node.Inline{node.Text(`\a`)}},
			},
		},
		{
			`\\**`,
			[]node.Node{
				&node.Line{"Line", []node.Inline{node.Text("**")}},
			},
		},
		{
			`\\\**`,
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					node.Text(`\`),
					&node.Uniform{"Strong", nil},
				}},
			},
		},
		{
			"\\\\`(",
			[]node.Node{
				&node.Line{"Line", []node.Inline{node.Text("`(")}},
			},
		},
		{
			"\\\\\\`(",
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					node.Text(`\`),
					&node.Escaped{"Code", nil},
				}},
			},
		},
		{
			`\\<`,
			[]node.Node{
				&node.Line{"Line", []node.Inline{node.Text(`\<`)}},
			},
		},
		{
			`\\\<`,
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					node.Text(`\<`),
				}},
			},
		},

		{
			`a\**`,
			[]node.Node{
				&node.Line{"Line", []node.Inline{node.Text("a**")}},
			},
		},
		{
			`\**\**`,
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					&node.Uniform{"Strong", []node.Inline{
						node.Text("**"),
					}},
				}},
			},
		},
		{
			"\\``\\``",
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					&node.Escaped{"Code", []byte(`\`)},
				}},
			},
		},
		{
			"``\\**\n\\**",
			[]node.Node{
				&node.Fenced{"CodeBlock", [][]byte{[]byte(`\**`), []byte(`\**`)}, nil},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			test(t, c.in, c.out, nil)
		})
	}
}

func TestLineComment(t *testing.T) {
	cases := []struct {
		in  string
		out []node.Node
	}{
		{
			"//",
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					node.LineComment(""),
				}},
			},
		},
		{
			"//a",
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					node.LineComment("a"),
				}},
			},
		},
		{
			"//\n//",
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					node.LineComment(""),
				}},
				&node.Line{"Line", []node.Inline{
					node.LineComment(""),
				}},
			},
		},

		// after inline
		{
			"a//",
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					node.Text("a"),
					node.LineComment(""),
				}},
			},
		},

		// escape
		{
			`\\//`,
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					node.Text("//"),
				}},
			},
		},
		{
			`a\//`,
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					node.Text("a//"),
				}},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			test(t, c.in, c.out, nil)
		})
	}
}

func TestInvalidUTF8Encoding(t *testing.T) {
	const fcb = "\x80" // first continuation byte

	cases := []struct {
		name string
		in   string
		out  []node.Node
	}{
		{
			"at the beginning",
			fcb + "a",
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					node.Text(string(unicode.ReplacementChar) + "a"),
				},
				}},
		},
		{
			"in the middle",
			"a" + fcb + "b",
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					node.Text("a" + string(unicode.ReplacementChar) + "b"),
				},
				}},
		},
		{
			"in the end",
			"a" + fcb,
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					node.Text("a" + string(unicode.ReplacementChar)),
				},
				}},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			test(t, c.in, c.out, []error{parser.ErrInvalidUTF8Encoding})
		})
	}
}

func TestNULL(t *testing.T) {
	const null = "\u0000"

	cases := []struct {
		name string
		in   string
		out  []node.Node
	}{
		{
			"at the beginning",
			null + "a",
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					node.Text(string(unicode.ReplacementChar) + "a"),
				},
				}},
		},
		{
			"in the middle",
			"a" + null + "b",
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					node.Text("a" + string(unicode.ReplacementChar) + "b"),
				},
				}},
		},
		{
			"in the end",
			"a" + null,
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					node.Text("a" + string(unicode.ReplacementChar)),
				},
				}},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			test(t, c.in, c.out, []error{parser.ErrIllegalNULL})
		})
	}
}

func TestBOM(t *testing.T) {
	const bom = "\uFEFF"

	t.Run("at the beginning", func(t *testing.T) {
		test(
			t,
			bom+"a",
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					node.Text("a"),
				},
				}},
			nil,
		)
	})

	cases := []struct {
		name string
		in   string
		out  []node.Node
	}{
		{
			"in the middle",
			"a" + bom + "b",
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					node.Text("a" + string(unicode.ReplacementChar) + "b"),
				},
				}},
		},
		{
			"in the end",
			"a" + bom,
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					node.Text("a" + string(unicode.ReplacementChar)),
				},
				}},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			test(t, c.in, c.out, []error{parser.ErrIllegalBOM})
		})
	}
}

// test compares the string representation of nodes generated by the parser from
// the argument in and the nodes of the argument out. Expected error must be
// encountered once; test calls t.Error() if it is encountered multiple times or
// if it is never encountered.
func test(t *testing.T, in string, out []node.Node, expectedErrors []error) {
	nodes, errs := parser.Parse(strings.NewReader(in))

	if expectedErrors == nil {
		for _, err := range errs {
			t.Errorf("got error %q", err)
		}
	} else {
		leftErrs := expectedErrors // errors we have not encountered yet
		for _, err := range errs {
			if i := errorIndex(leftErrs, err); i > -1 {
				// remove error
				leftErrs = append(leftErrs[:i], leftErrs[i+1:]...)
				continue
			}

			t.Errorf("got error %q", err)
		}

		// if some expected errors were not encountered
		for _, le := range leftErrs {
			t.Errorf("want error %q", le)
		}
	}

	got, want := stringifier.Stringify(node.BlocksToNodes(nodes)...), stringifier.Stringify(out...)
	if got != want {
		t.Errorf("\ngot\n%s\nwant\n%s", got, want)
	}
}

func errorIndex(errors []error, err error) int {
	for i, e := range errors {
		if err == e {
			return i
		}
	}
	return -1
}
