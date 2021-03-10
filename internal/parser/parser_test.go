package parser_test

import (
	"strings"
	"testing"
	"to/internal/node"
	"to/internal/parser"
	"to/internal/stringifier"
	"unicode"
)

func TestLine(t *testing.T) {
	cases := []struct {
		in  string
		out []node.Node
	}{
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
			[]node.Node{&node.Walled{"Blockquote", nil}},
		},
		{
			">\na",
			[]node.Node{
				&node.Walled{"Blockquote", nil},
				&node.Line{"Line", []node.Inline{node.Text("a")}},
			},
		},
		{
			">>",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Walled{"Blockquote", nil},
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
				&node.Walled{"Blockquote", nil},
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
					&node.Walled{"Blockquote", nil},
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
					&node.Walled{"Blockquote", nil},
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
					&node.Walled{"Blockquote", nil},
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
				&node.Walled{"Blockquote", nil},
				&node.Walled{"Blockquote", nil},
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
			[]node.Node{&node.Hanging{"DescriptionList", nil}},
		},
		{
			"**",
			[]node.Node{
				&node.Hanging{"DescriptionList", []node.Block{
					&node.Hanging{"DescriptionList", nil},
				}},
			},
		},
		{
			"*a",
			[]node.Node{
				&node.Hanging{"DescriptionList", []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
				}},
			},
		},
		{
			"*\n*",
			[]node.Node{
				&node.Hanging{"DescriptionList", nil},
				&node.Hanging{"DescriptionList", nil},
			},
		},
		{
			"*\n\n*",
			[]node.Node{
				&node.Hanging{"DescriptionList", nil},
				&node.Hanging{"DescriptionList", nil},
			},
		},
		{
			"*a\nb",
			[]node.Node{
				&node.Hanging{"DescriptionList", []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
				}},
				&node.Line{"Line", []node.Inline{node.Text("b")}},
			},
		},
		{
			"*a\n b",
			[]node.Node{
				&node.Hanging{"DescriptionList", []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
					&node.Line{"Line", []node.Inline{node.Text("b")}},
				}},
			},
		},
		{
			"*a\n  b",
			[]node.Node{
				&node.Hanging{"DescriptionList", []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
					&node.Line{"Line", []node.Inline{node.Text("b")}},
				}},
			},
		},

		// spacing
		{
			" *a\n b",
			[]node.Node{
				&node.Hanging{"DescriptionList", []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
				}},
				&node.Line{"Line", []node.Inline{node.Text("b")}},
			},
		},
		{
			" *a\n  b",
			[]node.Node{
				&node.Hanging{"DescriptionList", []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
					&node.Line{"Line", []node.Inline{node.Text("b")}},
				}},
			},
		},

		// nested
		{
			"*\n *",
			[]node.Node{
				&node.Hanging{"DescriptionList", []node.Block{
					&node.Hanging{"DescriptionList", nil},
				}},
			},
		},
		{
			"*a\n *b",
			[]node.Node{
				&node.Hanging{"DescriptionList", []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
					&node.Hanging{"DescriptionList", []node.Block{
						&node.Line{"Line", []node.Inline{node.Text("b")}},
					}},
				}},
			},
		},
		{
			"**a\n b",
			[]node.Node{
				&node.Hanging{"DescriptionList", []node.Block{
					&node.Hanging{"DescriptionList", []node.Block{
						&node.Line{"Line", []node.Inline{node.Text("a")}},
					}},
					&node.Line{"Line", []node.Inline{node.Text("b")}},
				}},
			},
		},
		{
			"**a\n  b",
			[]node.Node{
				&node.Hanging{"DescriptionList", []node.Block{
					&node.Hanging{"DescriptionList", []node.Block{
						&node.Line{"Line", []node.Inline{node.Text("a")}},
						&node.Line{"Line", []node.Inline{node.Text("b")}},
					}},
				}},
			},
		},
		{
			"**a\n   b",
			[]node.Node{
				&node.Hanging{"DescriptionList", []node.Block{
					&node.Hanging{"DescriptionList", []node.Block{
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
					&node.Hanging{"DescriptionList", nil},
				}},
			},
		},
		{
			">*\n*",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Hanging{"DescriptionList", nil},
				}},
				&node.Hanging{"DescriptionList", nil},
			},
		},
		{
			">*\n>*",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Hanging{"DescriptionList", nil},
					&node.Hanging{"DescriptionList", nil},
				}},
			},
		},
		{
			">*\n> *",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Hanging{"DescriptionList", []node.Block{
						&node.Hanging{"DescriptionList", nil},
					}},
				}},
			},
		},
		{
			">*\n> a",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Hanging{"DescriptionList", []node.Block{
						&node.Line{"Line", []node.Inline{node.Text("a")}},
					}},
				}},
			},
		},
		{
			"> *\n>  a",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Hanging{"DescriptionList", []node.Block{
						&node.Line{"Line", []node.Inline{node.Text("a")}},
					}},
				}},
			},
		},

		// nested+spacing
		{
			" *\n *",
			[]node.Node{
				&node.Hanging{"DescriptionList", nil},
				&node.Hanging{"DescriptionList", nil},
			},
		},
		{
			" *\n  *",
			[]node.Node{
				&node.Hanging{"DescriptionList", []node.Block{
					&node.Hanging{"DescriptionList", nil},
				}},
			},
		},

		// tab (equals 8 spaces in this regard)
		{
			"*a\n\tb",
			[]node.Node{
				&node.Hanging{"DescriptionList", []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
					&node.Line{"Line", []node.Inline{node.Text("b")}},
				}},
			},
		},
		{
			"\t*a\n\tb",
			[]node.Node{
				&node.Hanging{"DescriptionList", []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
				}},
				&node.Line{"Line", []node.Inline{node.Text("b")}},
			},
		},
		{
			"\t*a\n\t b",
			[]node.Node{
				&node.Hanging{"DescriptionList", []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
					&node.Line{"Line", []node.Inline{node.Text("b")}},
				}},
			},
		},
		{
			"\t*a\n \tb",
			[]node.Node{
				&node.Hanging{"DescriptionList", []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
					&node.Line{"Line", []node.Inline{node.Text("b")}},
				}},
			},
		},
		{
			"\t*a\n  \tb",
			[]node.Node{
				&node.Hanging{"DescriptionList", []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
					&node.Line{"Line", []node.Inline{node.Text("b")}},
				}},
			},
		},

		{
			"\t\t*a\n                b",
			[]node.Node{
				&node.Hanging{"DescriptionList", []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
				}},
				&node.Line{"Line", []node.Inline{node.Text("b")}},
			},
		},
		{
			"\t\t*a\n                 b",
			[]node.Node{
				&node.Hanging{"DescriptionList", []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
					&node.Line{"Line", []node.Inline{node.Text("b")}},
				}},
			},
		},
		{
			"                *a\n\t\tb",
			[]node.Node{
				&node.Hanging{"DescriptionList", []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
				}},
				&node.Line{"Line", []node.Inline{node.Text("b")}},
			},
		},
		{
			"               *a\n\t\tb",
			[]node.Node{
				&node.Hanging{"DescriptionList", []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
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

func TestFenced(t *testing.T) {
	cases := []struct {
		in  string
		out []node.Node
	}{
		{
			"``",
			[]node.Node{&node.Fenced{"CodeBlock", nil}},
		},
		{
			"``a",
			[]node.Node{
				&node.Fenced{"CodeBlock", [][]byte{[]byte("a")}},
			},
		},
		{
			"``\na",
			[]node.Node{
				&node.Fenced{"CodeBlock", [][]byte{nil, []byte("a")}},
			},
		},
		{
			"``\n\na",
			[]node.Node{
				&node.Fenced{"CodeBlock", [][]byte{
					nil,
					nil,
					[]byte("a")},
				},
			},
		},
		{
			"````",
			[]node.Node{
				&node.Fenced{"CodeBlock", nil},
			},
		},
		{
			"``\n``",
			[]node.Node{
				&node.Fenced{"CodeBlock", nil},
			},
		},
		{
			"```\n```",
			[]node.Node{
				&node.Fenced{"CodeBlock", nil},
			},
		},
		{
			"```\n``\n```",
			[]node.Node{
				&node.Fenced{"CodeBlock", [][]byte{nil, []byte("``")}},
			},
		},
		{
			"```\n`````",
			[]node.Node{
				&node.Fenced{"CodeBlock", nil},
				&node.Fenced{"CodeBlock", nil},
			},
		},
		{
			"``\n``a",
			[]node.Node{
				&node.Fenced{"CodeBlock", nil},
				&node.Line{"Line", []node.Inline{node.Text("a")}},
			},
		},

		// nesting
		{
			"``\n>",
			[]node.Node{
				&node.Fenced{"CodeBlock", [][]byte{nil, []byte(">")}},
			},
		},
		{
			">``\na",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Fenced{"CodeBlock", nil},
				}},
				&node.Line{"Line", []node.Inline{node.Text("a")}},
			},
		},
		{
			">``\n>a",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Fenced{"CodeBlock", [][]byte{nil, []byte("a")}},
				}},
			},
		},
		{
			">``\n>a\n>``",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Fenced{"CodeBlock", [][]byte{nil, []byte("a")}},
				}},
			},
		},
		{
			">``\n>a\n>``b",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Fenced{"CodeBlock", [][]byte{nil, []byte("a")}},
					&node.Line{"Line", []node.Inline{node.Text("b")}},
				}},
			},
		},
		{
			">``\n>a\n>``\nb",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Fenced{"CodeBlock", [][]byte{nil, []byte("a")}},
				}},
				&node.Line{"Line", []node.Inline{node.Text("b")}},
			},
		},

		// nesting+spacing
		{
			"> ``\n>a",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Fenced{"CodeBlock", [][]byte{nil, []byte("a")}},
				}},
			},
		},
		{
			"> ``\n> a",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Fenced{"CodeBlock", [][]byte{nil, []byte("a")}},
				}},
			},
		},
		{
			"> ``\n>  a",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Fenced{"CodeBlock", [][]byte{nil, []byte(" a")}},
				}},
			},
		},
		{
			">  ``\n>  a",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Fenced{"CodeBlock", [][]byte{nil, []byte("a")}},
				}},
			},
		},

		// tab
		{
			">\t``\n>a",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Fenced{"CodeBlock", [][]byte{nil, []byte("a")}},
				}},
			},
		},
		{
			">\t``\n>        a",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Fenced{"CodeBlock", [][]byte{nil, []byte("a")}},
				}},
			},
		},
		{
			">\t``\n>         a",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Fenced{"CodeBlock", [][]byte{nil, []byte(" a")}},
				}},
			},
		},
		{
			">\t``\n>            a",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Fenced{"CodeBlock", [][]byte{nil, []byte("    a")}},
				}},
			},
		},
		{
			"> ``\n>\ta",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Fenced{"CodeBlock", [][]byte{nil, []byte("       a")}},
				}},
			},
		},
		{
			"> \t``\n>\t a",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Fenced{"CodeBlock", [][]byte{nil, []byte("a")}},
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
			[]node.Node{&node.Walled{"Blockquote", nil}},
		},

		// space
		{
			" >",
			[]node.Node{&node.Walled{"Blockquote", nil}},
		},
		{
			"> >",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Walled{"Blockquote", nil},
				}},
			},
		},
		{
			">  >",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Walled{"Blockquote", nil},
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
			[]node.Node{&node.Walled{"Blockquote", nil}},
		},
		{
			">\t>",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Walled{"Blockquote", nil},
				}},
			},
		},
		{
			">\t\t>",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Walled{"Blockquote", nil},
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
			[]node.Node{&node.Walled{"Blockquote", nil}},
		},
		{
			"> \t>",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Walled{"Blockquote", nil},
				}},
			},
		},
		{
			">  \t>",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Walled{"Blockquote", nil},
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

func TestForward(t *testing.T) {
	cases := []struct {
		in  string
		out []node.Node
	}{
		{
			"<",
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					&node.Forward{"Link", nil, nil},
				}},
			},
		},
		{
			"<>",
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					&node.Forward{"Link", nil, nil},
				}},
			},
		},
		{
			"<><",
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					&node.Forward{"Link", nil, nil},
				}},
			},
		},
		{
			"<><>",
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					&node.Forward{"Link", nil, nil},
				}},
			},
		},

		{
			"<a",
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					&node.Forward{"Link", []byte("a"), nil},
				}},
			},
		},
		{
			"<a>",
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					&node.Forward{"Link", []byte("a"), nil},
				}},
			},
		},
		{
			"<a><b",
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					&node.Forward{"Link", []byte("b"), []node.Inline{
						node.Text("a"),
					}},
				}},
			},
		},
		{
			"<a><b>",
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					&node.Forward{"Link", []byte("b"), []node.Inline{
						node.Text("a"),
					}},
				}},
			},
		},
		{
			"<a><b><c>",
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					&node.Forward{"Link", []byte("b"), []node.Inline{
						node.Text("a"),
					}},
					&node.Forward{"Link", []byte("c"), nil},
				}},
			},
		},

		// nested
		{
			"<**",
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					&node.Forward{"Link", []byte("**"), nil},
				}},
			},
		},
		{
			"<**>",
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					&node.Forward{"Link", []byte("**"), nil},
				}},
			},
		},
		{
			"<a><**",
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					&node.Forward{"Link", []byte("**"), []node.Inline{
						node.Text("a"),
					}},
				}},
			},
		},
		{
			"<a><**>",
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					&node.Forward{"Link", []byte("**"), []node.Inline{
						node.Text("a"),
					}},
				}},
			},
		},

		{
			"<**><a",
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					&node.Forward{"Link", []byte("a"), []node.Inline{
						&node.Uniform{"Strong", nil},
					}},
				}},
			},
		},
		{
			"<**><a>",
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					&node.Forward{"Link", []byte("a"), []node.Inline{
						&node.Uniform{"Strong", nil},
					}},
				}},
			},
		},
		{
			"<**a**><b",
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					&node.Forward{"Link", []byte("b"), []node.Inline{
						&node.Uniform{"Strong", []node.Inline{
							node.Text("a"),
						}},
					}},
				}},
			},
		},
		{
			"<**a**><b>",
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					&node.Forward{"Link", []byte("b"), []node.Inline{
						&node.Uniform{"Strong", []node.Inline{
							node.Text("a"),
						}},
					}},
				}},
			},
		},

		{
			"<**<a>><b",
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					&node.Forward{"Link", []byte("**<a"), nil},
					node.Text(">"),
					&node.Forward{"Link", []byte("b"), nil},
				}},
			},
		},

		{
			"|**<**><a>",
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					&node.Uniform{"Strong", []node.Inline{
						&node.Forward{"Link", nil, nil},
					}},
					node.Text(">"),
					&node.Forward{"Link", []byte("a"), nil},
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
			"|",
			[]node.Node{
				&node.Line{"Line", nil},
			},
		},
		{
			"||",
			[]node.Node{
				&node.Line{"Line", []node.Inline{node.Text("|")}},
			},
		},
		{
			"|a",
			[]node.Node{
				&node.Line{"Line", []node.Inline{node.Text("a")}},
			},
		},
		{
			"|>",
			[]node.Node{
				&node.Line{"Line", []node.Inline{node.Text(">")}},
			},
		},
		{
			"||>",
			[]node.Node{
				&node.Line{"Line", []node.Inline{node.Text("|>")}},
			},
		},
		{
			"|\n|",
			[]node.Node{
				&node.Line{"Line", nil},
				&node.Line{"Line", nil},
			},
		},
		{
			"|a\n|b",
			[]node.Node{
				&node.Line{"Line", []node.Inline{node.Text("a")}},
				&node.Line{"Line", []node.Inline{node.Text("b")}},
			},
		},
		{
			"|``",
			[]node.Node{
				&node.Line{"Line", []node.Inline{&node.Escaped{"Code", nil}}},
			},
		},
		{
			"||``",
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					node.Text("|"),
					&node.Escaped{"Code", nil},
				}},
			},
		},

		{
			"``|**\n|**",
			[]node.Node{
				&node.Fenced{"CodeBlock", [][]byte{[]byte("|**"), []byte("|**")}},
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
				&node.Line{"Line", []node.Inline{node.Text(`\`)}},
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
				&node.Line{"Line", []node.Inline{node.Text(`\a`)}},
			},
		},
		{
			`\**`,
			[]node.Node{
				&node.Line{"Line", []node.Inline{node.Text("**")}},
			},
		},
		{
			`\\**`,
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					node.Text(`\`),
					&node.Uniform{"Strong", nil},
				}},
			},
		},
		{
			"\\`(",
			[]node.Node{
				&node.Line{"Line", []node.Inline{node.Text("`(")}},
			},
		},
		{
			"\\\\`(",
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					node.Text(`\`),
					&node.Escaped{"Code", nil},
				}},
			},
		},
		{
			`\<`,
			[]node.Node{
				&node.Line{"Line", []node.Inline{node.Text("<")}},
			},
		},
		{
			`\\<`,
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					node.Text(`\`),
					&node.Forward{"Link", nil, nil},
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
			`|**\**`,
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					&node.Uniform{"Strong", []node.Inline{
						node.Text("**"),
					}},
				}},
			},
		},
		{
			"|``\\``",
			[]node.Node{
				&node.Line{"Line", []node.Inline{
					&node.Escaped{"Code", []byte(`\`)},
				}},
			},
		},
		{
			"``\\**\n\\**",
			[]node.Node{
				&node.Fenced{"CodeBlock", [][]byte{[]byte(`\**`), []byte(`\**`)}},
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
			`\//`,
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
