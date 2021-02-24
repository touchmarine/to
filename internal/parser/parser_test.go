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

func TestWalledOnlyLineChildren(t *testing.T) {
	cases := []struct {
		in  string
		out []node.Node
	}{
		{
			"|",
			[]node.Node{&node.Walled{"Paragraph", nil}},
		},
		{
			"|\na",
			[]node.Node{
				&node.Walled{"Paragraph", nil},
				&node.Line{"Line", []node.Inline{node.Text("a")}},
			},
		},
		{
			"||",
			[]node.Node{
				&node.Walled{"Paragraph", []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("|")}},
				}},
			},
		},
		{
			"|>",
			[]node.Node{
				&node.Walled{"Paragraph", []node.Block{
					&node.Line{"Line", []node.Inline{node.Text(">")}},
				}},
			},
		},
		{
			"|a",
			[]node.Node{
				&node.Walled{"Paragraph", []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
				}},
			},
		},
		{
			"|\n|",
			[]node.Node{&node.Walled{"Paragraph", nil}},
		},
		{
			"|a\n|b",
			[]node.Node{
				&node.Walled{"Paragraph", []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
					&node.Line{"Line", []node.Inline{node.Text("b")}},
				}},
			},
		},
		{
			"|\n||",
			[]node.Node{
				&node.Walled{"Paragraph", []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("|")}},
				}},
			},
		},
		{
			"||\n||",
			[]node.Node{
				&node.Walled{"Paragraph", []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("|")}},
					&node.Line{"Line", []node.Inline{node.Text("|")}},
				}},
			},
		},
		{
			"||\n|",
			[]node.Node{
				&node.Walled{"Paragraph", []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("|")}},
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

func TestHanging(t *testing.T) {
	cases := []struct {
		in  string
		out []node.Node
	}{
		{
			"*",
			[]node.Node{&node.Walled{"DescriptionList", nil}},
		},
		{
			"*\na",
			[]node.Node{
				&node.Walled{"DescriptionList", nil},
				&node.Line{"Line", []node.Inline{node.Text("a")}},
			},
		},
		{
			"**",
			[]node.Node{
				&node.Walled{"DescriptionList", []node.Block{
					&node.Walled{"DescriptionList", nil},
				}},
			},
		},
		{
			"*a",
			[]node.Node{
				&node.Walled{"DescriptionList", []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
				}},
			},
		},
		{
			"*\n*",
			[]node.Node{
				&node.Walled{"DescriptionList", nil},
				&node.Walled{"DescriptionList", nil},
			},
		},
		{
			"*\n        ",
			[]node.Node{
				&node.Walled{"DescriptionList", nil},
			},
		},
		{
			"*a\n        b",
			[]node.Node{
				&node.Walled{"DescriptionList", []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
					&node.Line{"Line", []node.Inline{node.Text("b")}},
				}},
			},
		},
		{
			"*\n        *",
			[]node.Node{
				&node.Walled{"DescriptionList", []node.Block{
					&node.Walled{"DescriptionList", nil},
				}},
			},
		},
		{
			"*a\n        *b",
			[]node.Node{
				&node.Walled{"DescriptionList", []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
					&node.Walled{"DescriptionList", []node.Block{
						&node.Line{"Line", []node.Inline{node.Text("b")}},
					}},
				}},
			},
		},
		{
			"**\n                ",
			[]node.Node{
				&node.Walled{"DescriptionList", []node.Block{
					&node.Walled{"DescriptionList", nil},
				}},
			},
		},
		{
			"**a\n                b",
			[]node.Node{
				&node.Walled{"DescriptionList", []node.Block{
					&node.Walled{"DescriptionList", []node.Block{
						&node.Line{"Line", []node.Inline{node.Text("a")}},
						&node.Line{"Line", []node.Inline{node.Text("b")}},
					}},
				}},
			},
		},
		{
			"**a\n                 b",
			[]node.Node{
				&node.Walled{"DescriptionList", []node.Block{
					&node.Walled{"DescriptionList", []node.Block{
						&node.Line{"Line", []node.Inline{node.Text("a")}},
						&node.Line{"Line", []node.Inline{node.Text("b")}},
					}},
				}},
			},
		},
		{
			"**a\n\t        b",
			[]node.Node{
				&node.Walled{"DescriptionList", []node.Block{
					&node.Walled{"DescriptionList", []node.Block{
						&node.Line{"Line", []node.Inline{node.Text("a")}},
						&node.Line{"Line", []node.Inline{node.Text("b")}},
					}},
				}},
			},
		},
		{
			"**a\n        \tb",
			[]node.Node{
				&node.Walled{"DescriptionList", []node.Block{
					&node.Walled{"DescriptionList", []node.Block{
						&node.Line{"Line", []node.Inline{node.Text("a")}},
						&node.Line{"Line", []node.Inline{node.Text("b")}},
					}},
				}},
			},
		},
		{
			"**\n        ",
			[]node.Node{
				&node.Walled{"DescriptionList", []node.Block{
					&node.Walled{"DescriptionList", nil},
				}},
			},
		},
		{
			"**a\n        b",
			[]node.Node{
				&node.Walled{"DescriptionList", []node.Block{
					&node.Walled{"DescriptionList", []node.Block{
						&node.Line{"Line", []node.Inline{node.Text("a")}},
					}},
					&node.Line{"Line", []node.Inline{node.Text("b")}},
				}},
			},
		},
		{
			"*\n\n*",
			[]node.Node{
				&node.Walled{"DescriptionList", nil},
				&node.Walled{"DescriptionList", nil},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			test(t, c.in, c.out, nil)
		})
	}
}

func TestHangingSpaces(t *testing.T) {
	cases := []struct {
		in  string
		out []node.Node
	}{
		{
			"*",
			[]node.Node{&node.Walled{"DescriptionList", nil}},
		},
		{
			"*\na",
			[]node.Node{
				&node.Walled{"DescriptionList", nil},
				&node.Line{"Line", []node.Inline{node.Text("a")}},
			},
		},
		{
			"**",
			[]node.Node{
				&node.Walled{"DescriptionList", []node.Block{
					&node.Walled{"DescriptionList", nil},
				}},
			},
		},
		{
			"*a",
			[]node.Node{
				&node.Walled{"DescriptionList", []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
				}},
			},
		},
		{
			"*\n*",
			[]node.Node{
				&node.Walled{"DescriptionList", nil},
				&node.Walled{"DescriptionList", nil},
			},
		},
		{
			"*\n        ",
			[]node.Node{
				&node.Walled{"DescriptionList", nil},
			},
		},
		{
			"*a\n        b",
			[]node.Node{
				&node.Walled{"DescriptionList", []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
					&node.Line{"Line", []node.Inline{node.Text("b")}},
				}},
			},
		},
		{
			"*\n        *",
			[]node.Node{
				&node.Walled{"DescriptionList", []node.Block{
					&node.Walled{"DescriptionList", nil},
				}},
			},
		},
		{
			"*a\n        *b",
			[]node.Node{
				&node.Walled{"DescriptionList", []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
					&node.Walled{"DescriptionList", []node.Block{
						&node.Line{"Line", []node.Inline{node.Text("b")}},
					}},
				}},
			},
		},
		{
			"**\n                ",
			[]node.Node{
				&node.Walled{"DescriptionList", []node.Block{
					&node.Walled{"DescriptionList", nil},
				}},
			},
		},
		{
			"**a\n                b",
			[]node.Node{
				&node.Walled{"DescriptionList", []node.Block{
					&node.Walled{"DescriptionList", []node.Block{
						&node.Line{"Line", []node.Inline{node.Text("a")}},
						&node.Line{"Line", []node.Inline{node.Text("b")}},
					}},
				}},
			},
		},
		{
			"**\n        ",
			[]node.Node{
				&node.Walled{"DescriptionList", []node.Block{
					&node.Walled{"DescriptionList", nil},
				}},
			},
		},
		{
			"**a\n        b",
			[]node.Node{
				&node.Walled{"DescriptionList", []node.Block{
					&node.Walled{"DescriptionList", []node.Block{
						&node.Line{"Line", []node.Inline{node.Text("a")}},
					}},
					&node.Line{"Line", []node.Inline{node.Text("b")}},
				}},
			},
		},
		{
			"*\n\n*",
			[]node.Node{
				&node.Walled{"DescriptionList", nil},
				&node.Walled{"DescriptionList", nil},
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
		{
			" >",
			[]node.Node{&node.Walled{"Blockquote", nil}},
		},
		{
			"\t>",
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
			">\t>",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Walled{"Blockquote", nil},
				}},
			},
		},
		{
			"> ",
			[]node.Node{
				&node.Walled{"Blockquote", nil},
			},
		},
		{
			"> \n> ",
			[]node.Node{
				&node.Walled{"Blockquote", nil},
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
		{
			"> a",
			[]node.Node{
				&node.Walled{"Blockquote", []node.Block{
					&node.Line{"Line", []node.Inline{node.Text("a")}},
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
			">\t``\n>        a",
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
			"__**a**",
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
			"__**a**__",
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
			"__**a__",
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
