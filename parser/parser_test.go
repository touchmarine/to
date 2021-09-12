package parser_test

import (
	"github.com/touchmarine/to/node"
	"github.com/touchmarine/to/parser"
	"github.com/touchmarine/to/stringifier"
	"strings"
	"testing"
	"unicode"
)

func TestLeaf(t *testing.T) {
	cases := []struct {
		in  string
		out []node.Node
	}{
		{
			" ",
			[]node.Node{},
		},
		{
			" \n ",
			[]node.Node{},
		},
		{
			"a",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
			},
		},
		{
			"a\n",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
			},
		},
		{
			"a\nb",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a b")}}},
			},
		},
		{
			"a\n\nb",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("b")}}},
			},
		},
		{
			"a\n\n\nb",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("b")}}},
			},
		},
		{
			"a\n\n\n\nb",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("b")}}},
			},
		},
		{
			"a__\nb",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a")},
					&node.Uniform{"MA", []node.Inline{
						&node.Text{"MT", []byte(" b")},
					}},
				}},
			},
		},
		{
			"a\n>b",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
				&node.Walled{"B", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("b")}}},
				}},
			},
		},

		// nested
		{
			">a\n>b",
			[]node.Node{
				&node.Walled{"B", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a b")}}},
				}},
			},
		},
		{
			">a\n\n>b",
			[]node.Node{
				&node.Walled{"B", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
				}},
				&node.Walled{"B", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("b")}}},
				}},
			},
		},
		{
			">a\n\n\n>b",
			[]node.Node{
				&node.Walled{"B", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
				}},
				&node.Walled{"B", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("b")}}},
				}},
			},
		},

		{
			">a\nb",
			[]node.Node{
				&node.Walled{"B", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
				}},
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("b")}}},
			},
		},
		{
			">>a\n>b",
			[]node.Node{
				&node.Walled{"B", []node.Block{
					&node.Walled{"B", []node.Block{
						&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
					}},
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("b")}}},
				}},
			},
		},

		{
			">a\n>\n>b",
			[]node.Node{
				&node.Walled{"B", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("b")}}},
				}},
			},
		},

		{
			"*a\n b",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a b")}}},
				}},
			},
		},
		{
			"*a\n*b",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
				}},
				&node.Hanging{"A", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("b")}}},
				}},
			},
		},
		{
			"*a\n\n*b",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
				}},
				&node.Hanging{"A", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("b")}}},
				}},
			},
		},
		{
			"*a\n\n\n*b",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
				}},
				&node.Hanging{"A", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("b")}}},
				}},
			},
		},

		{
			"*a\nb",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
				}},
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("b")}}},
			},
		},
		{
			"**a\n b",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Hanging{"A", []node.Block{
						&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
					}},
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("b")}}},
				}},
			},
		},

		// spacing
		{
			"a \nb",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a b")}}},
			},
		},
		{
			"a  \nb",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a b")}}},
			},
		},
		{
			"a\n b",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a b")}}},
			},
		},
		{
			"a\n  b",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a b")}}},
			},
		},
		{
			"*a\n b",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a b")}}},
				}},
			},
		},
		{
			"*a\n  b",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a b")}}},
				}},
			},
		},

		// escape
		{
			"a\n\\__",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a __")},
				}},
			},
		},
		{
			"a\n\\\\__",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a \\")},
					&node.Uniform{"MA", nil},
				}},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			test(t, c.in, c.out, nil, parser.ElementMap{
				"T": {
					Name: "T",
					Type: node.TypeLeaf,
				},
				"A": {
					Name:      "A",
					Type:      node.TypeHanging,
					Delimiter: "*",
				},
				"B": {
					Name:      "B",
					Type:      node.TypeWalled,
					Delimiter: ">",
				},
				"MA": {
					Name:      "MA",
					Type:      node.TypeUniform,
					Delimiter: "_",
				},
				"MT": {
					Name: "MT",
					Type: node.TypeText,
				},
			})
		})
	}
}

func TestLine(t *testing.T) {
	cases := []struct {
		in  string
		out []node.Node
	}{
		{
			" ",
			[]node.Node{},
		},
		{
			" \n ",
			[]node.Node{},
		},
		{
			" \n \n",
			[]node.Node{},
		},
		{
			"a",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
			},
		},
		{
			"a_*",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a_*")}}},
			},
		},

		{
			"\na",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
			},
		},
		{
			"\n\na",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
			},
		},
		{
			" \na",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
			},
		},
		{
			"\t\na",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
			},
		},
		{
			"a\n",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
			},
		},
		{
			"a\nb",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a b")}}},
			},
		},
		{
			"a\n\nb",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("b")}}},
			},
		},
		{
			"a\n\n\nb",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("b")}}},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			test(t, c.in, c.out, nil, parser.ElementMap{
				"T": {
					Name: "T",
					Type: node.TypeLeaf,
				},
				"MT": {
					Name: "MT",
					Type: node.TypeText,
				},
			})
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
			[]node.Node{&node.Walled{"A", nil}},
		},
		{
			">\na",
			[]node.Node{
				&node.Walled{"A", nil},
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
			},
		},
		{
			">>",
			[]node.Node{
				&node.Walled{"A", []node.Block{
					&node.Walled{"A", nil},
				}},
			},
		},
		{
			">a",
			[]node.Node{
				&node.Walled{"A", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
				}},
			},
		},
		{
			">\n>",
			[]node.Node{
				&node.Walled{"A", nil},
			},
		},
		{
			">a\n>b",
			[]node.Node{
				&node.Walled{"A", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a b")}}},
				}},
			},
		},
		{
			">\n>>",
			[]node.Node{
				&node.Walled{"A", []node.Block{
					&node.Walled{"A", nil},
				}},
			},
		},
		{
			">a\n>>b",
			[]node.Node{
				&node.Walled{"A", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
					&node.Walled{"A", []node.Block{
						&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("b")}}},
					}},
				}},
			},
		},
		{
			">>\n>>",
			[]node.Node{
				&node.Walled{"A", []node.Block{
					&node.Walled{"A", nil},
				}},
			},
		},
		{
			">>a\n>>b",
			[]node.Node{
				&node.Walled{"A", []node.Block{
					&node.Walled{"A", []node.Block{
						&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a b")}}},
					}},
				}},
			},
		},
		{
			">>\n>",
			[]node.Node{
				&node.Walled{"A", []node.Block{
					&node.Walled{"A", nil},
				}},
			},
		},
		{
			">>a\n>b",
			[]node.Node{
				&node.Walled{"A", []node.Block{
					&node.Walled{"A", []node.Block{
						&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
					}},
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("b")}}},
				}},
			},
		},
		{
			">\n\n>",
			[]node.Node{
				&node.Walled{"A", nil},
				&node.Walled{"A", nil},
			},
		},
		{
			">a\n\n>b",
			[]node.Node{
				&node.Walled{"A", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
				}},
				&node.Walled{"A", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("b")}}},
				}},
			},
		},
		{
			">a\n \n>b",
			[]node.Node{
				&node.Walled{"A", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
				}},
				&node.Walled{"A", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("b")}}},
				}},
			},
		},

		// spacing
		{
			" >",
			[]node.Node{&node.Walled{"A", nil}},
		},

		// regression
		{
			"> >",
			[]node.Node{
				&node.Walled{"A", []node.Block{
					&node.Walled{"A", nil},
				}},
			},
		},
		{
			">\t>",
			[]node.Node{
				&node.Walled{"A", []node.Block{
					&node.Walled{"A", nil},
				}},
			},
		},
		{
			"> > >",
			[]node.Node{
				&node.Walled{"A", []node.Block{
					&node.Walled{"A", []node.Block{
						&node.Walled{"A", nil},
					}},
				}},
			},
		},
		{
			">\n >",
			[]node.Node{
				&node.Walled{"A", nil},
			},
		},
		{
			" >\n>",
			[]node.Node{
				&node.Walled{"A", nil},
			},
		},
		{
			">a\n >b",
			[]node.Node{
				&node.Walled{"A", []node.Block{
					&node.Leaf{"T", []node.Inline{
						&node.Text{"MT", []byte("a b")},
					}},
				}},
			},
		},

		{
			">\na",
			[]node.Node{
				&node.Walled{"A", nil},
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
			},
		},
		{
			">\n>\na",
			[]node.Node{
				&node.Walled{"A", nil},
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
			},
		},
		{
			">\n>\n>\na",
			[]node.Node{
				&node.Walled{"A", nil},
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
			},
		},
		{
			">\n>\n>\n>\na",
			[]node.Node{
				&node.Walled{"A", nil},
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			test(t, c.in, c.out, nil, parser.ElementMap{
				"T": {
					Name: "T",
					Type: node.TypeLeaf,
				},
				"A": {
					Name:      "A",
					Type:      node.TypeWalled,
					Delimiter: ">",
				},
				"MT": {
					Name: "MT",
					Type: node.TypeText,
				},
			})
		})
	}
}

func TestVerbatimWalled(t *testing.T) {
	cases := []struct {
		in  string
		out []node.Node
	}{
		{
			"/",
			[]node.Node{
				&node.VerbatimWalled{"A", nil},
			},
		},
		{
			"/a",
			[]node.Node{
				&node.VerbatimWalled{"A", [][]byte{[]byte("a")}},
			},
		},
		{
			"/a\n",
			[]node.Node{
				&node.VerbatimWalled{"A", [][]byte{[]byte("a")}},
			},
		},
		{
			"/a\n/b",
			[]node.Node{
				&node.VerbatimWalled{"A", [][]byte{
					[]byte("a"),
					[]byte("b"),
				}},
			},
		},
		{
			"/a\n/\n/b",
			[]node.Node{
				&node.VerbatimWalled{"A", [][]byte{
					[]byte("a"),
					nil,
					[]byte("b"),
				}},
			},
		},

		// no nested content allowed
		{
			"/>a",
			[]node.Node{
				&node.VerbatimWalled{"A", [][]byte{[]byte(">a")}},
			},
		},
		{
			"/**a",
			[]node.Node{
				&node.VerbatimWalled{"A", [][]byte{[]byte("**a")}},
			},
		},
		{
			`/\**a`,
			[]node.Node{
				&node.VerbatimWalled{"A", [][]byte{[]byte(`\**a`)}},
			},
		},
		{
			`/\\**a`,
			[]node.Node{
				&node.VerbatimWalled{"A", [][]byte{[]byte(`\\**a`)}},
			},
		},

		// spacing
		{
			"/\n/",
			[]node.Node{
				&node.VerbatimWalled{"A", [][]byte{nil, nil}},
			},
		},
		{
			"/ \n/ ",
			[]node.Node{
				&node.VerbatimWalled{"A", [][]byte{
					[]byte(" "),
					[]byte(" "),
				}},
			},
		},
		{
			"/ a",
			[]node.Node{
				&node.VerbatimWalled{"A", [][]byte{[]byte(" a")}},
			},
		},
		{
			"/a\n/ b",
			[]node.Node{
				&node.VerbatimWalled{"A", [][]byte{
					[]byte("a"),
					[]byte(" b"),
				}},
			},
		},
		{
			">/ a\n>/ b",
			[]node.Node{
				&node.Walled{"B", []node.Block{
					&node.VerbatimWalled{"A", [][]byte{
						[]byte(" a"),
						[]byte(" b"),
					}},
				}},
			},
		},
		{
			"*/ a\n / b",
			[]node.Node{
				&node.Hanging{"C", []node.Block{
					&node.VerbatimWalled{"A", [][]byte{
						[]byte(" a"),
						[]byte(" b"),
					}},
				}},
			},
		},
		{
			" / a",
			[]node.Node{
				&node.VerbatimWalled{"A", [][]byte{[]byte(" a")}},
			},
		},
		{
			" / \n / ",
			[]node.Node{
				&node.VerbatimWalled{"A", [][]byte{
					[]byte(" "),
					[]byte(" "),
				}},
			},
		},
		{
			"*/\n / b",
			[]node.Node{
				&node.Hanging{"C", []node.Block{
					&node.VerbatimWalled{"A", [][]byte{
						nil,
						[]byte(" b"),
					}},
				}},
			},
		},

		// continuation (stop)
		{
			"/a\nb",
			[]node.Node{
				&node.VerbatimWalled{"A", [][]byte{[]byte("a")}},
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("b")}}},
			},
		},
		{
			"/a\n\n/b",
			[]node.Node{
				&node.VerbatimWalled{"A", [][]byte{[]byte("a")}},
				&node.VerbatimWalled{"A", [][]byte{[]byte("b")}},
			},
		},
		{
			"/a\n>b",
			[]node.Node{
				&node.VerbatimWalled{"A", [][]byte{[]byte("a")}},
				&node.Walled{"B", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("b")}}},
				}},
			},
		},
		{
			"/a\n*b",
			[]node.Node{
				&node.VerbatimWalled{"A", [][]byte{[]byte("a")}},
				&node.Hanging{"C", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("b")}}},
				}},
			},
		},
		{
			">/a\n/b",
			[]node.Node{
				&node.Walled{"B", []node.Block{
					&node.VerbatimWalled{"A", [][]byte{[]byte("a")}},
				}},
				&node.VerbatimWalled{"A", [][]byte{[]byte("b")}},
			},
		},
		{
			"*/a\n/b",
			[]node.Node{
				&node.Hanging{"C", []node.Block{
					&node.VerbatimWalled{"A", [][]byte{[]byte("a")}},
				}},
				&node.VerbatimWalled{"A", [][]byte{[]byte("b")}},
			},
		},

		// nested
		{
			">/a",
			[]node.Node{
				&node.Walled{"B", []node.Block{
					&node.VerbatimWalled{"A", [][]byte{[]byte("a")}},
				}},
			},
		},
		{
			">/a\n>/b",
			[]node.Node{
				&node.Walled{"B", []node.Block{
					&node.VerbatimWalled{"A", [][]byte{
						[]byte("a"),
						[]byte("b"),
					}},
				}},
			},
		},
		{
			"*/a\n /b",
			[]node.Node{
				&node.Hanging{"C", []node.Block{
					&node.VerbatimWalled{"A", [][]byte{
						[]byte("a"),
						[]byte("b"),
					}},
				}},
			},
		},

		{
			">/a\n>/\n>/b",
			[]node.Node{
				&node.Walled{"B", []node.Block{
					&node.VerbatimWalled{"A", [][]byte{
						[]byte("a"),
						nil,
						[]byte("b"),
					}},
				}},
			},
		},
		{
			"*/a\n /\n /b",
			[]node.Node{
				&node.Hanging{"C", []node.Block{
					&node.VerbatimWalled{"A", [][]byte{
						[]byte("a"),
						nil,
						[]byte("b"),
					}},
				}},
			},
		},
		{
			"*\n /a\n /\n /b",
			[]node.Node{
				&node.Hanging{"C", []node.Block{
					&node.VerbatimWalled{"A", [][]byte{
						[]byte("a"),
						nil,
						[]byte("b"),
					}},
				}},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			test(t, c.in, c.out, nil, parser.ElementMap{
				"T": {
					Name: "T",
					Type: node.TypeLeaf,
				},
				"A": {
					Name:      "A",
					Type:      node.TypeVerbatimWalled,
					Delimiter: "/",
				},
				"B": {
					Name:      "B",
					Type:      node.TypeWalled,
					Delimiter: ">",
				},
				"C": {
					Name:      "C",
					Type:      node.TypeHanging,
					Delimiter: "*",
				},
				"MT": {
					Name: "MT",
					Type: node.TypeText,
				},
			})
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
			[]node.Node{&node.Hanging{"A", nil}},
		},
		{
			"**",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Hanging{"A", nil},
				}},
			},
		},
		{
			"*a",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
				}},
			},
		},
		{
			"*\n*",
			[]node.Node{
				&node.Hanging{"A", nil},
				&node.Hanging{"A", nil},
			},
		},
		{
			"*\n\n*",
			[]node.Node{
				&node.Hanging{"A", nil},
				&node.Hanging{"A", nil},
			},
		},
		{
			"*\n\n\n*",
			[]node.Node{
				&node.Hanging{"A", nil},
				&node.Hanging{"A", nil},
			},
		},
		{
			"*\n\t\n*",
			[]node.Node{
				&node.Hanging{"A", nil},
				&node.Hanging{"A", nil},
			},
		},
		{
			"*a\n\n*b",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
				}},
				&node.Hanging{"A", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("b")}}},
				}},
			},
		},
		{
			"*a\nb",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
				}},
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("b")}}},
			},
		},
		{
			"*a\n b",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a b")}}},
				}},
			},
		},
		{
			"*a\n  b",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a b")}}},
				}},
			},
		},

		// spacing
		{
			" *a\n b",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
				}},
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("b")}}},
			},
		},
		{
			" *a\n  b",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a b")}}},
				}},
			},
		},

		// nested
		{
			"*\n *",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Hanging{"A", nil},
				}},
			},
		},
		{
			"*a\n *b",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
					&node.Hanging{"A", []node.Block{
						&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("b")}}},
					}},
				}},
			},
		},
		{
			"**a\n b",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Hanging{"A", []node.Block{
						&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
					}},
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("b")}}},
				}},
			},
		},
		{
			"**a\n  b",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Hanging{"A", []node.Block{
						&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a b")}}},
					}},
				}},
			},
		},
		{
			"**a\n   b",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Hanging{"A", []node.Block{
						&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a b")}}},
					}},
				}},
			},
		},

		{
			">*",
			[]node.Node{
				&node.Walled{"B", []node.Block{
					&node.Hanging{"A", nil},
				}},
			},
		},
		{
			">*\n*",
			[]node.Node{
				&node.Walled{"B", []node.Block{
					&node.Hanging{"A", nil},
				}},
				&node.Hanging{"A", nil},
			},
		},
		{
			">*\n>*",
			[]node.Node{
				&node.Walled{"B", []node.Block{
					&node.Hanging{"A", nil},
					&node.Hanging{"A", nil},
				}},
			},
		},
		{
			">*\n> *",
			[]node.Node{
				&node.Walled{"B", []node.Block{
					&node.Hanging{"A", []node.Block{
						&node.Hanging{"A", nil},
					}},
				}},
			},
		},
		{
			">*\n> a",
			[]node.Node{
				&node.Walled{"B", []node.Block{
					&node.Hanging{"A", []node.Block{
						&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
					}},
				}},
			},
		},
		{
			"> *\n>  a",
			[]node.Node{
				&node.Walled{"B", []node.Block{
					&node.Hanging{"A", []node.Block{
						&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
					}},
				}},
			},
		},

		// nested+spacing
		{
			" *\n *",
			[]node.Node{
				&node.Hanging{"A", nil},
				&node.Hanging{"A", nil},
			},
		},
		{
			" *\n  *",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Hanging{"A", nil},
				}},
			},
		},

		// tab (equals 8 spaces in this regard)
		{
			"*a\n\tb",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a b")}}},
				}},
			},
		},
		{
			"\t*a\n\tb",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
				}},
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("b")}}},
			},
		},
		{
			"\t*a\n\t b",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a b")}}},
				}},
			},
		},
		{
			"\t*a\n \tb",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a b")}}},
				}},
			},
		},
		{
			"\t*a\n  \tb",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a b")}}},
				}},
			},
		},

		{
			"\t\t*a\n                b",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
				}},
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("b")}}},
			},
		},
		{
			"\t\t*a\n                 b",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a b")}}},
				}},
			},
		},
		{
			"                *a\n\t\tb",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
				}},
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("b")}}},
			},
		},
		{
			"               *a\n\t\tb",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a b")}}},
				}},
			},
		},

		// nested+blank lines
		{
			"*\n *",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Hanging{"A", nil},
				}},
			},
		},
		//*
		//
		// *
		{
			"*\n\n *",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Hanging{"A", nil},
				}},
			},
		},
		{
			"*\n \t\n *",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Hanging{"A", nil},
				}},
			},
		},

		//**
		//
		//*
		{
			"**\n\n*",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Hanging{"A", nil},
				}},
				&node.Hanging{"A", nil},
			},
		},
		{
			"a\n\n**\n\n*",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a")},
				}},
				&node.Hanging{"A", []node.Block{
					&node.Hanging{"A", nil},
				}},
				&node.Hanging{"A", nil},
			},
		},

		{
			"**\n\na",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Hanging{"A", nil},
				}},
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a")},
				}},
			},
		},
		{
			"**\n\n a",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Hanging{"A", nil},
					&node.Leaf{"T", []node.Inline{
						&node.Text{"MT", []byte("a")},
					}},
				}},
			},
		},
		{
			"**\n\n  a",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Hanging{"A", []node.Block{
						&node.Leaf{"T", []node.Inline{
							&node.Text{"MT", []byte("a")},
						}},
					}},
				}},
			},
		},
		{
			"**\n\n  a\nb",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Hanging{"A", []node.Block{
						&node.Leaf{"T", []node.Inline{
							&node.Text{"MT", []byte("a")},
						}},
					}},
				}},
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("b")},
				}},
			},
		},

		// regression
		{
			"*\n >b",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Walled{"B", []node.Block{
						&node.Leaf{"T", []node.Inline{
							&node.Text{"MT", []byte("b")},
						}},
					}},
				}},
			},
		},
		{
			"*>a\n >b",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Walled{"B", []node.Block{
						&node.Leaf{"T", []node.Inline{
							&node.Text{"MT", []byte("a b")},
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
			[]node.Node{&node.Hanging{"A", []node.Block{
				&node.Hanging{"A", []node.Block{
					&node.Leaf{"T", []node.Inline{
						&node.Text{"MT", []byte("a")},
					}},
					&node.Hanging{"A", []node.Block{
						&node.Leaf{"T", []node.Inline{
							&node.Text{"MT", []byte("b")},
						}},
					}},
				}},
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("c")},
				}},
			}}},
		},

		{
			"*\n  >*a",
			[]node.Node{&node.Hanging{"A", []node.Block{
				&node.Walled{"B", []node.Block{
					&node.Hanging{"A", []node.Block{
						&node.Leaf{"T", []node.Inline{
							&node.Text{"MT", []byte("a")},
						}},
					}},
				}},
			}}},
		},
		{
			"*\n\t>*a",
			[]node.Node{&node.Hanging{"A", []node.Block{
				&node.Walled{"B", []node.Block{
					&node.Hanging{"A", []node.Block{
						&node.Leaf{"T", []node.Inline{
							&node.Text{"MT", []byte("a")},
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
			[]node.Node{&node.Hanging{"A", []node.Block{
				&node.Walled{"B", []node.Block{
					&node.Hanging{"A", nil},
				}},
			}}},
		},
		//  >*a
		// > *b
		{
			"  >*a\n > *b",
			[]node.Node{
				&node.Walled{"B", []node.Block{
					&node.Hanging{"A", []node.Block{
						&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
						&node.Hanging{"A", []node.Block{
							&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("b")}}},
						}},
					}},
				}},
			},
		},
		{
			"  > *a\n >  *b",
			[]node.Node{
				&node.Walled{"B", []node.Block{
					&node.Hanging{"A", []node.Block{
						&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
						&node.Hanging{"A", []node.Block{
							&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("b")}}},
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
				&node.Walled{"B", []node.Block{
					&node.Hanging{"A", []node.Block{
						&node.Hanging{"A", nil},
					}},
				}},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			test(t, c.in, c.out, nil, parser.ElementMap{
				"T": {
					Name: "T",
					Type: node.TypeLeaf,
				},
				"A": {
					Name:      "A",
					Type:      node.TypeHanging,
					Delimiter: "*",
				},
				"B": {
					Name:      "B",
					Type:      node.TypeWalled,
					Delimiter: ">",
				},
				"MT": {
					Name: "MT",
					Type: node.TypeText,
				},
			})
		})
	}
}

func TestRankedHanging(t *testing.T) {
	cases := []struct {
		in  string
		out []node.Node
	}{
		{
			"=",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("=")},
				}},
			},
		},
		{
			"==",
			[]node.Node{&node.RankedHanging{"A", 2, nil}},
		},
		{
			"= =",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("= =")},
				}},
			},
		},
		{
			"== ==",
			[]node.Node{
				&node.RankedHanging{"A", 2, []node.Block{
					&node.RankedHanging{"A", 2, nil},
				}},
			},
		},
		{
			"==a",
			[]node.Node{&node.RankedHanging{"A", 2, []node.Block{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a")},
				}},
			}}},
		},
		{
			"==\n==",
			[]node.Node{
				&node.RankedHanging{"A", 2, nil},
				&node.RankedHanging{"A", 2, nil},
			},
		},

		{
			"==a\nb",
			[]node.Node{
				&node.RankedHanging{"A", 2, []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
				}},
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("b")}}},
			},
		},
		{
			"==a\n b",
			[]node.Node{
				&node.RankedHanging{"A", 2, []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
				}},
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("b")}}},
			},
		},
		{
			"==a\n  b",
			[]node.Node{
				&node.RankedHanging{"A", 2, []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a b")}}},
				}},
			},
		},
		{
			"==a\n   b",
			[]node.Node{
				&node.RankedHanging{"A", 2, []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a b")}}},
				}},
			},
		},

		// spacing
		{
			" ==a\n  b",
			[]node.Node{
				&node.RankedHanging{"A", 2, []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
				}},
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("b")}}},
			},
		},
		{
			" ==a\n   b",
			[]node.Node{
				&node.RankedHanging{"A", 2, []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a b")}}},
				}},
			},
		},

		// nested
		{
			"==\n  ==",
			[]node.Node{
				&node.RankedHanging{"A", 2, []node.Block{
					&node.RankedHanging{"A", 2, nil},
				}},
			},
		},
		{
			"==a\n  ==b",
			[]node.Node{
				&node.RankedHanging{"A", 2, []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
					&node.RankedHanging{"A", 2, []node.Block{
						&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("b")}}},
					}},
				}},
			},
		},
		{
			"== ==a\n  b",
			[]node.Node{
				&node.RankedHanging{"A", 2, []node.Block{
					&node.RankedHanging{"A", 2, []node.Block{
						&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
					}},
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("b")}}},
				}},
			},
		},
		{
			"== ==a\n     b",
			[]node.Node{
				&node.RankedHanging{"A", 2, []node.Block{
					&node.RankedHanging{"A", 2, []node.Block{
						&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a b")}}},
					}},
				}},
			},
		},
		{
			"== ==a\n      b",
			[]node.Node{
				&node.RankedHanging{"A", 2, []node.Block{
					&node.RankedHanging{"A", 2, []node.Block{
						&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a b")}}},
					}},
				}},
			},
		},

		{
			">==",
			[]node.Node{
				&node.Walled{"B", []node.Block{
					&node.RankedHanging{"A", 2, nil},
				}},
			},
		},
		{
			">==\n==",
			[]node.Node{
				&node.Walled{"B", []node.Block{
					&node.RankedHanging{"A", 2, nil},
				}},
				&node.RankedHanging{"A", 2, nil},
			},
		},
		{
			">==\n>==",
			[]node.Node{
				&node.Walled{"B", []node.Block{
					&node.RankedHanging{"A", 2, nil},
					&node.RankedHanging{"A", 2, nil},
				}},
			},
		},
		{
			">==\n>  ==",
			[]node.Node{
				&node.Walled{"B", []node.Block{
					&node.RankedHanging{"A", 2, []node.Block{
						&node.RankedHanging{"A", 2, nil},
					}},
				}},
			},
		},
		{
			">==\n>  a",
			[]node.Node{
				&node.Walled{"B", []node.Block{
					&node.RankedHanging{"A", 2, []node.Block{
						&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
					}},
				}},
			},
		},
		{
			"> ==\n>   a",
			[]node.Node{
				&node.Walled{"B", []node.Block{
					&node.RankedHanging{"A", 2, []node.Block{
						&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
					}},
				}},
			},
		},

		// nested+spacing
		{
			" ==\n ==",
			[]node.Node{
				&node.RankedHanging{"A", 2, nil},
				&node.RankedHanging{"A", 2, nil},
			},
		},
		{
			" ==\n   ==",
			[]node.Node{
				&node.RankedHanging{"A", 2, []node.Block{
					&node.RankedHanging{"A", 2, nil},
				}},
			},
		},

		// tab (equals 8 spaces in this regard)
		{
			"==a\n\tb",
			[]node.Node{
				&node.RankedHanging{"A", 2, []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a b")}}},
				}},
			},
		},
		{
			"\t==a\n\tb",
			[]node.Node{
				&node.RankedHanging{"A", 2, []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
				}},
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("b")}}},
			},
		},
		{
			"\t==a\n\t  b",
			[]node.Node{
				&node.RankedHanging{"A", 2, []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a b")}}},
				}},
			},
		},
		{
			"\t==a\n  \tb",
			[]node.Node{
				&node.RankedHanging{"A", 2, []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a b")}}},
				}},
			},
		},
		{
			"\t==a\n   \tb",
			[]node.Node{
				&node.RankedHanging{"A", 2, []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a b")}}},
				}},
			},
		},

		{
			"\t\t==a\n                 b",
			[]node.Node{
				&node.RankedHanging{"A", 2, []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
				}},
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("b")}}},
			},
		},
		{
			"\t\t==a\n                  b",
			[]node.Node{
				&node.RankedHanging{"A", 2, []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a b")}}},
				}},
			},
		},
		{
			"                 ==a\n\t\tb",
			[]node.Node{
				&node.RankedHanging{"A", 2, []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
				}},
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("b")}}},
			},
		},
		{
			"              ==a\n\t\tb",
			[]node.Node{
				&node.RankedHanging{"A", 2, []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a b")}}},
				}},
			},
		},

		// regression
		{
			"==\n  >b",
			[]node.Node{
				&node.RankedHanging{"A", 2, []node.Block{
					&node.Walled{"B", []node.Block{
						&node.Leaf{"T", []node.Inline{
							&node.Text{"MT", []byte("b")},
						}},
					}},
				}},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			test(t, c.in, c.out, nil, parser.ElementMap{
				"T": {
					Name: "T",
					Type: node.TypeLeaf,
				},
				"A": {
					Name:      "A",
					Type:      node.TypeRankedHanging,
					Delimiter: "=",
				},
				"B": {
					Name:      "B",
					Type:      node.TypeWalled,
					Delimiter: ">",
				},
				"MT": {
					Name: "MT",
					Type: node.TypeText,
				},
			})
		})
	}
}

func TestVerbatimLine(t *testing.T) {
	cases := []struct {
		in  string
		out []node.Node
	}{
		{
			".image",
			[]node.Node{&node.VerbatimLine{"A", nil}},
		},
		{
			".imagea",
			[]node.Node{&node.VerbatimLine{"A", []byte("a")}},
		},
		{
			".image a",
			[]node.Node{&node.VerbatimLine{"A", []byte(" a")}},
		},
		{
			".imagea ",
			[]node.Node{&node.VerbatimLine{"A", []byte("a ")}},
		},
		{
			".image*",
			[]node.Node{&node.VerbatimLine{"A", []byte("*")}},
		},
		{
			`.image\**`,
			[]node.Node{&node.VerbatimLine{"A", []byte(`\**`)}},
		},

		{
			".image\n      a",
			[]node.Node{
				&node.VerbatimLine{"A", nil},
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a")},
				}},
			},
		},
		{
			">.image",
			[]node.Node{
				&node.Walled{"B", []node.Block{
					&node.VerbatimLine{"A", nil},
				}},
			},
		},
		{
			">\n.image",
			[]node.Node{
				&node.Walled{"B", nil},
				&node.VerbatimLine{"A", nil},
			},
		},
		{
			".image\n>",
			[]node.Node{
				&node.VerbatimLine{"A", nil},
				&node.Walled{"B", nil},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			test(t, c.in, c.out, nil, parser.ElementMap{
				"T": {
					Name: "T",
					Type: node.TypeLeaf,
				},
				"A": {
					Name:      "A",
					Type:      node.TypeVerbatimLine,
					Delimiter: ".image",
				},
				"B": {
					Name:      "B",
					Type:      node.TypeWalled,
					Delimiter: ">",
				},
				"MT": {
					Name: "MT",
					Type: node.TypeText,
				},
			})
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
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("1")},
				}},
			},
		},
		{
			"1.",
			[]node.Node{&node.Hanging{"A", nil}},
		},
		{
			"1.1.",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Hanging{"A", nil},
				}},
			},
		},
		{
			"1.a",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
				}},
			},
		},
		{
			"1.\n1.",
			[]node.Node{
				&node.Hanging{"A", nil},
				&node.Hanging{"A", nil},
			},
		},
		{
			"1.a\nb",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
				}},
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("b")}}},
			},
		},
		{
			"1.a\n b",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
				}},
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("b")}}},
			},
		},
		{
			"1.a\n  b",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a b")}}},
				}},
			},
		},

		// spacing
		{
			" 1.a\n b",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
				}},
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("b")}}},
			},
		},
		{
			" 1.a\n  b",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
				}},
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("b")}}},
			},
		},
		{
			" 1.a\n   b",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a b")}}},
				}},
			},
		},

		// nested
		{
			"1.\n  1.",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Hanging{"A", nil},
				}},
			},
		},
		{
			"1.a\n  1.b",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
					&node.Hanging{"A", []node.Block{
						&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("b")}}},
					}},
				}},
			},
		},

		{
			">1.",
			[]node.Node{
				&node.Walled{"B", []node.Block{
					&node.Hanging{"A", nil},
				}},
			},
		},
		{
			">1.\n1.",
			[]node.Node{
				&node.Walled{"B", []node.Block{
					&node.Hanging{"A", nil},
				}},
				&node.Hanging{"A", nil},
			},
		},
		{
			">1.\n>1.",
			[]node.Node{
				&node.Walled{"B", []node.Block{
					&node.Hanging{"A", nil},
					&node.Hanging{"A", nil},
				}},
			},
		},
		{
			">1.\n>  1.",
			[]node.Node{
				&node.Walled{"B", []node.Block{
					&node.Hanging{"A", []node.Block{
						&node.Hanging{"A", nil},
					}},
				}},
			},
		},
		{
			">1.\n>  a",
			[]node.Node{
				&node.Walled{"B", []node.Block{
					&node.Hanging{"A", []node.Block{
						&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
					}},
				}},
			},
		},
		{
			"> 1.\n>   a",
			[]node.Node{
				&node.Walled{"B", []node.Block{
					&node.Hanging{"A", []node.Block{
						&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
					}},
				}},
			},
		},

		{
			"1.-\n\n1.",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Hanging{"C", nil},
				}},
				&node.Hanging{"A", nil},
			},
		},

		// regression
		{
			"1.\n  >b",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Walled{"B", []node.Block{
						&node.Leaf{"T", []node.Inline{
							&node.Text{"MT", []byte("b")},
						}},
					}},
				}},
			},
		},
		{
			"1.>a\n  >b",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Walled{"B", []node.Block{
						&node.Leaf{"T", []node.Inline{
							&node.Text{"MT", []byte("a b")},
						}},
					}},
				}},
			},
		},
		{
			"1. >a\n   >b",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Walled{"B", []node.Block{
						&node.Leaf{"T", []node.Inline{
							&node.Text{"MT", []byte("a b")},
						}},
					}},
				}},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			test(t, c.in, c.out, nil, parser.ElementMap{
				"T": {
					Name: "T",
					Type: node.TypeLeaf,
				},
				"A": {
					Name:      "A",
					Type:      node.TypeHanging,
					Delimiter: "1.",
				},
				"B": {
					Name:      "B",
					Type:      node.TypeWalled,
					Delimiter: ">",
				},
				"C": {
					Name:      "C",
					Type:      node.TypeHanging,
					Delimiter: "-",
				},
				"MT": {
					Name: "MT",
					Type: node.TypeText,
				},
			})
		})
	}
}

func TestFenced(t *testing.T) {
	cases := []struct {
		in  string
		out []node.Node
	}{
		{
			"`",
			[]node.Node{&node.Fenced{"A", nil, nil}},
		},
		{
			"`a",
			[]node.Node{
				&node.Fenced{"A", [][]byte{[]byte("a")}, nil},
			},
		},
		{
			"`a`",
			[]node.Node{
				&node.Fenced{"A", [][]byte{[]byte("a`")}, nil},
			},
		},
		{
			"`\na",
			[]node.Node{
				&node.Fenced{"A", [][]byte{nil, []byte("a")}, nil},
			},
		},
		{
			"`\n a",
			[]node.Node{
				&node.Fenced{"A", [][]byte{nil, []byte(" a")}, nil},
			},
		},
		{
			"`\n\na",
			[]node.Node{
				&node.Fenced{
					"A",
					[][]byte{
						nil,
						nil,
						[]byte("a")},
					nil,
				},
			},
		},
		{
			"``",
			[]node.Node{
				&node.Fenced{"A", [][]byte{[]byte("`")}, nil},
			},
		},
		{
			"`\n`",
			[]node.Node{
				&node.Fenced{"A", [][]byte{nil}, nil},
			},
		},
		{
			"`\n`a",
			[]node.Node{
				&node.Fenced{"A", nil, []byte("a")},
			},
		},

		// escape
		{
			"`\\\n``",
			[]node.Node{
				&node.Fenced{"A", [][]byte{nil, []byte("``")}, nil},
			},
		},
		{
			"`\\\n``\n`",
			[]node.Node{
				&node.Fenced{"A", [][]byte{nil, []byte("``"), []byte("`")}, nil},
			},
		},
		{
			"`\\\n``\n\\`",
			[]node.Node{
				&node.Fenced{"A", [][]byte{nil, []byte("``")}, nil},
			},
		},

		// closing delimiter spacing
		{
			"`\n `",
			[]node.Node{
				&node.Fenced{"A", [][]byte{nil, []byte(" `")}, nil},
			},
		},
		{
			"`\\\n \\`",
			[]node.Node{
				&node.Fenced{"A", [][]byte{nil, []byte(" \\`")}, nil},
			},
		},

		// nested
		{
			"`\n>",
			[]node.Node{
				&node.Fenced{"A", [][]byte{nil, []byte(">")}, nil},
			},
		},
		{
			">`\na",
			[]node.Node{
				&node.Walled{"B", []node.Block{
					&node.Fenced{"A", nil, nil},
				}},
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
			},
		},
		{
			">`\n>a",
			[]node.Node{
				&node.Walled{"B", []node.Block{
					&node.Fenced{"A", [][]byte{nil, []byte("a")}, nil},
				}},
			},
		},
		{
			">`\n>a\n>`",
			[]node.Node{
				&node.Walled{"B", []node.Block{
					&node.Fenced{"A", [][]byte{nil, []byte("a")}, nil},
				}},
			},
		},
		{
			">`\n>a\n>`b",
			[]node.Node{
				&node.Walled{"B", []node.Block{
					&node.Fenced{"A", [][]byte{nil, []byte("a")}, []byte("b")},
				}},
			},
		},
		{
			">`\n>a\n>`\nb",
			[]node.Node{
				&node.Walled{"B", []node.Block{
					&node.Fenced{"A", [][]byte{nil, []byte("a")}, nil},
				}},
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("b")}}},
			},
		},

		// spacing
		{
			"> `\n>a",
			[]node.Node{
				&node.Walled{"B", []node.Block{
					&node.Fenced{"A", [][]byte{nil, []byte("a")}, nil},
				}},
			},
		},
		{
			"> `\n> a",
			[]node.Node{
				&node.Walled{"B", []node.Block{
					&node.Fenced{"A", [][]byte{nil, []byte("a")}, nil},
				}},
			},
		},
		{
			"> `\n>  a",
			[]node.Node{
				&node.Walled{"B", []node.Block{
					&node.Fenced{"A", [][]byte{nil, []byte(" a")}, nil},
				}},
			},
		},
		{
			">  `\n> a",
			[]node.Node{
				&node.Walled{"B", []node.Block{
					&node.Fenced{"A", [][]byte{nil, []byte("a")}, nil},
				}},
			},
		},
		{
			">  `\n>  a",
			[]node.Node{
				&node.Walled{"B", []node.Block{
					&node.Fenced{"A", [][]byte{nil, []byte("a")}, nil},
				}},
			},
		},

		{
			"* `\n a",
			[]node.Node{
				&node.Walled{"C", []node.Block{
					&node.Fenced{"A", [][]byte{nil, []byte("a")}, nil},
				}},
			},
		},
		{
			"* `\n  a",
			[]node.Node{
				&node.Walled{"C", []node.Block{
					&node.Fenced{"A", [][]byte{nil, []byte("a")}, nil},
				}},
			},
		},
		{
			"* `\n   a",
			[]node.Node{
				&node.Walled{"C", []node.Block{
					&node.Fenced{"A", [][]byte{nil, []byte(" a")}, nil},
				}},
			},
		},
		{
			"*  `\n  a",
			[]node.Node{
				&node.Walled{"C", []node.Block{
					&node.Fenced{"A", [][]byte{nil, []byte("a")}, nil},
				}},
			},
		},
		{
			"*  `\n   a",
			[]node.Node{
				&node.Walled{"C", []node.Block{
					&node.Fenced{"A", [][]byte{nil, []byte("a")}, nil},
				}},
			},
		},

		// tab
		{
			">\t`\n>a",
			[]node.Node{
				&node.Walled{"B", []node.Block{
					&node.Fenced{"A", [][]byte{nil, []byte("a")}, nil},
				}},
			},
		},
		{
			">\t`\n>        a",
			[]node.Node{
				&node.Walled{"B", []node.Block{
					&node.Fenced{"A", [][]byte{nil, []byte("a")}, nil},
				}},
			},
		},
		{
			">\t`\n>         a",
			[]node.Node{
				&node.Walled{"B", []node.Block{
					&node.Fenced{"A", [][]byte{nil, []byte(" a")}, nil},
				}},
			},
		},
		{
			">\t`\n>            a",
			[]node.Node{
				&node.Walled{"B", []node.Block{
					&node.Fenced{"A", [][]byte{nil, []byte("    a")}, nil},
				}},
			},
		},
		{
			"> `\n>\ta",
			[]node.Node{
				&node.Walled{"B", []node.Block{
					&node.Fenced{"A", [][]byte{nil, []byte("       a")}, nil},
				}},
			},
		},
		{
			"> \t`\n>\t a",
			[]node.Node{
				&node.Walled{"B", []node.Block{
					&node.Fenced{"A", [][]byte{nil, []byte("a")}, nil},
				}},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			test(t, c.in, c.out, nil, parser.ElementMap{
				"T": {
					Name: "T",
					Type: node.TypeLeaf,
				},
				"A": {
					Name:      "A",
					Type:      node.TypeFenced,
					Delimiter: "`",
				},
				"B": {
					Name:      "B",
					Type:      node.TypeWalled,
					Delimiter: ">",
				},
				"C": {
					Name:      "C",
					Type:      node.TypeHanging,
					Delimiter: "*",
				},
				"MT": {
					Name: "MT",
					Type: node.TypeText,
				},
			})
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
				&node.Walled{"A", nil},
			},
		},

		// space
		{
			" >",
			[]node.Node{&node.Walled{"A", nil}},
		},
		{
			"> >",
			[]node.Node{
				&node.Walled{"A", []node.Block{
					&node.Walled{"A", nil},
				}},
			},
		},
		{
			">  >",
			[]node.Node{
				&node.Walled{"A", []node.Block{
					&node.Walled{"A", nil},
				}},
			},
		},
		{
			"> ",
			[]node.Node{
				&node.Walled{"A", nil},
			},
		},
		{
			"> a",
			[]node.Node{
				&node.Walled{"A", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
				}},
			},
		},

		// tab
		{
			"\t>",
			[]node.Node{&node.Walled{"A", nil}},
		},
		{
			">\t>",
			[]node.Node{
				&node.Walled{"A", []node.Block{
					&node.Walled{"A", nil},
				}},
			},
		},
		{
			">\t\t>",
			[]node.Node{
				&node.Walled{"A", []node.Block{
					&node.Walled{"A", nil},
				}},
			},
		},
		{
			">\t",
			[]node.Node{
				&node.Walled{"A", nil},
			},
		},
		{
			">\ta",
			[]node.Node{
				&node.Walled{"A", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
				}},
			},
		},

		// space+tab
		{
			" \t>",
			[]node.Node{&node.Walled{"A", nil}},
		},
		{
			"> \t>",
			[]node.Node{
				&node.Walled{"A", []node.Block{
					&node.Walled{"A", nil},
				}},
			},
		},
		{
			">  \t>",
			[]node.Node{
				&node.Walled{"A", []node.Block{
					&node.Walled{"A", nil},
				}},
			},
		},
		{
			"> \t",
			[]node.Node{
				&node.Walled{"A", nil},
			},
		},
		{
			"> \ta",
			[]node.Node{
				&node.Walled{"A", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a")}}},
				}},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			test(t, c.in, c.out, nil, parser.ElementMap{
				"T": {
					Name: "T",
					Type: node.TypeLeaf,
				},
				"A": {
					Name:      "A",
					Type:      node.TypeWalled,
					Delimiter: ">",
				},
				"MT": {
					Name: "MT",
					Type: node.TypeText,
				},
			})
		})
	}
}

func TestUniform(t *testing.T) {
	cases := []struct {
		in  string
		out []node.Node
	}{
		{
			"a__",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a")},
					&node.Uniform{"MA", nil},
				}},
			},
		},
		{
			"a____",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a")},
					&node.Uniform{"MA", nil},
				}},
			},
		},
		{
			"a__b",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a")},
					&node.Uniform{"MA", []node.Inline{
						&node.Text{"MT", []byte("b")},
					}},
				}},
			},
		},
		{
			"a__b__",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a")},
					&node.Uniform{"MA", []node.Inline{
						&node.Text{"MT", []byte("b")},
					}},
				}},
			},
		},
		{
			"a__b__c",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a")},
					&node.Uniform{"MA", []node.Inline{
						&node.Text{"MT", []byte("b")},
					}},
					&node.Text{"MT", []byte("c")},
				}},
			},
		},

		// left-right delimiter
		{
			"((",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Uniform{"MB", nil},
				}},
			},
		},
		{
			"(())",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Uniform{"MB", nil},
				}},
			},
		},
		{
			"((a",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Uniform{"MB", []node.Inline{
						&node.Text{"MT", []byte("a")},
					}},
				}},
			},
		},
		{
			"((a))",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Uniform{"MB", []node.Inline{
						&node.Text{"MT", []byte("a")},
					}},
				}},
			},
		},
		{
			"((a))b",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Uniform{"MB", []node.Inline{
						&node.Text{"MT", []byte("a")},
					}},
					&node.Text{"MT", []byte("b")},
				}},
			},
		},

		// across lines
		{
			"a__\nb",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a")},
					&node.Uniform{"MA", []node.Inline{
						&node.Text{"MT", []byte(" b")},
					}},
				}},
			},
		},
		{
			"a__\n>b",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a")},
					&node.Uniform{"MA", nil},
				}},
				&node.Walled{"A", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("b")}}},
				}},
			},
		},
		{
			">a__\nb",
			[]node.Node{
				&node.Walled{"A", []node.Block{
					&node.Leaf{"T", []node.Inline{
						&node.Text{"MT", []byte("a")}, &node.Uniform{"MA", nil},
					}},
				}},
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("b")}}},
			},
		},

		// across line spacing
		{
			"a__ \nb",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a")},
					&node.Uniform{"MA", []node.Inline{&node.Text{"MT", []byte(" b")}}},
				}},
			},
		},
		{
			"a__  \nb",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a")},
					&node.Uniform{"MA", []node.Inline{&node.Text{"MT", []byte(" b")}}},
				}},
			},
		},
		{
			"a__\n b",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a")},
					&node.Uniform{"MA", []node.Inline{&node.Text{"MT", []byte(" b")}}},
				}},
			},
		},
		{
			"a__\n  b",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a")},
					&node.Uniform{"MA", []node.Inline{&node.Text{"MT", []byte(" b")}}},
				}},
			},
		},
		{
			"*a__\n b",
			[]node.Node{
				&node.Hanging{"B", []node.Block{
					&node.Leaf{"T", []node.Inline{
						&node.Text{"MT", []byte("a")},
						&node.Uniform{"MA", []node.Inline{&node.Text{"MT", []byte(" b")}}},
					}},
				}},
			},
		},
		{
			"*a__\n  b",
			[]node.Node{
				&node.Hanging{"B", []node.Block{
					&node.Leaf{"T", []node.Inline{
						&node.Text{"MT", []byte("a")},
						&node.Uniform{"MA", []node.Inline{&node.Text{"MT", []byte(" b")}}},
					}},
				}},
			},
		},

		// nested
		{
			"a__**",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a")},
					&node.Uniform{"MA", []node.Inline{
						&node.Uniform{"MC", nil},
					}},
				}},
			},
		},
		{
			"a__**b",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a")},
					&node.Uniform{"MA", []node.Inline{
						&node.Uniform{"MC", []node.Inline{
							&node.Text{"MT", []byte("b")},
						}},
					}},
				}},
			},
		},
		{
			"a__**b**c",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a")},
					&node.Uniform{"MA", []node.Inline{
						&node.Uniform{"MC", []node.Inline{
							&node.Text{"MT", []byte("b")},
						}},
						&node.Text{"MT", []byte("c")},
					}},
				}},
			},
		},
		{
			"a__**b__c",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a")},
					&node.Uniform{"MA", []node.Inline{
						&node.Uniform{"MC", []node.Inline{
							&node.Text{"MT", []byte("b")},
						}},
					}},
					&node.Text{"MT", []byte("c")},
				}},
			},
		},
		{
			"a__**b**c__d",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a")},
					&node.Uniform{"MA", []node.Inline{
						&node.Uniform{"MC", []node.Inline{
							&node.Text{"MT", []byte("b")},
						}},
						&node.Text{"MT", []byte("c")},
					}},
					&node.Text{"MT", []byte("d")},
				}},
			},
		},

		// nested across lines
		{
			"a__**\nb",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a")},
					&node.Uniform{"MA", []node.Inline{
						&node.Uniform{"MC", []node.Inline{
							&node.Text{"MT", []byte(" b")},
						}},
					}},
				}},
			},
		},
		{
			"a__**b\nc**__d",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a")},
					&node.Uniform{"MA", []node.Inline{
						&node.Uniform{"MC", []node.Inline{
							&node.Text{"MT", []byte("b c")},
						}},
					}},
					&node.Text{"MT", []byte("d")},
				}},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			test(t, c.in, c.out, nil, parser.ElementMap{
				"T": {
					Name: "T",
					Type: node.TypeLeaf,
				},
				"MA": {
					Name:      "MA",
					Type:      node.TypeUniform,
					Delimiter: "_",
				},
				"MB": {
					Name:      "MB",
					Type:      node.TypeUniform,
					Delimiter: "(",
				},
				"MC": {
					Name:      "MC",
					Type:      node.TypeUniform,
					Delimiter: "*",
				},
				"A": {
					Name:      "A",
					Type:      node.TypeWalled,
					Delimiter: ">",
				},
				"B": {
					Name:      "B",
					Type:      node.TypeHanging,
					Delimiter: "*",
				},
				"MT": {
					Name: "MT",
					Type: node.TypeText,
				},
			})
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
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a")},
					&node.Escaped{"MA", nil},
				}},
			},
		},
		{
			"a```",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a")},
					&node.Escaped{"MA", []byte("`")},
				}},
			},
		},
		{
			"a````",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a")},
					&node.Escaped{"MA", nil},
				}},
			},
		},
		{
			"a`````",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a")},
					&node.Escaped{"MA", nil},
					&node.Text{"MT", []byte("`")},
				}},
			},
		},
		{
			"a``\\```",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a")},
					&node.Escaped{"MA", []byte("```")},
				}},
			},
		},
		{
			"a``\\`",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a")},
					&node.Escaped{"MA", []byte("`")},
				}},
			},
		},
		{
			"a``\\`\\``",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a")},
					&node.Escaped{"MA", []byte("`")},
				}},
			},
		},
		{
			"a``\\``\\``",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a")},
					&node.Escaped{"MA", []byte("``")},
				}},
			},
		},
		{
			"a``\\```\\``",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a")},
					&node.Escaped{"MA", []byte("```")},
				}},
			},
		},

		// left-right delim
		{
			"a[[",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a")},
					&node.Escaped{"MB", nil},
				}},
			},
		},
		{
			"a[[[",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a")},
					&node.Escaped{"MB", []byte("[")},
				}},
			},
		},
		{
			"a[[]]",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a")},
					&node.Escaped{"MB", nil},
				}},
			},
		},

		{
			"a\\````",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a`")},
					&node.Escaped{"MA", []byte("`")},
				}},
			},
		},
		{
			"a\\[[]]",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a[[]]")},
				}},
			},
		},

		// across lines
		{
			"a``\nb",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a")},
					&node.Escaped{"MA", []byte(" b")},
				}},
			},
		},
		{
			"a``\n>b",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a")},
					&node.Escaped{"MA", nil},
				}},
				&node.Walled{"B", []node.Block{
					&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("b")}}},
				}},
			},
		},
		{
			">a``\nb",
			[]node.Node{
				&node.Walled{"B", []node.Block{
					&node.Leaf{"T", []node.Inline{
						&node.Text{"MT", []byte("a")}, &node.Escaped{"MA", nil},
					}},
				}},
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("b")}}},
			},
		},

		// across line spacing
		{
			"a`` \nb",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a")},
					&node.Escaped{"MA", []byte(" b")},
				}},
			},
		},
		{
			"a``  \nb",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a")},
					&node.Escaped{"MA", []byte(" b")},
				}},
			},
		},
		{
			"a``\n b",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a")},
					&node.Escaped{"MA", []byte(" b")},
				}},
			},
		},
		{
			"a``\n  b",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a")},
					&node.Escaped{"MA", []byte(" b")},
				}},
			},
		},
		{
			"*a``\n b",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Leaf{"T", []node.Inline{
						&node.Text{"MT", []byte("a")},
						&node.Escaped{"MA", []byte(" b")},
					}},
				}},
			},
		},
		{
			"*a``\n  b",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Leaf{"T", []node.Inline{
						&node.Text{"MT", []byte("a")},
						&node.Escaped{"MA", []byte(" b")},
					}},
				}},
			},
		},

		// nested elements are not allowed
		{
			"a``__``",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a")},
					&node.Escaped{"MA", []byte("__")},
				}},
			},
		},
		{
			"a__``__b``c",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a")},
					&node.Uniform{"MC", []node.Inline{
						&node.Escaped{"MA", []byte("__b")},
						&node.Text{"MT", []byte("c")},
					}},
				}},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			test(t, c.in, c.out, nil, parser.ElementMap{
				"T": {
					Name: "T",
					Type: node.TypeLeaf,
				},
				"MA": {
					Name:      "MA",
					Type:      node.TypeEscaped,
					Delimiter: "`",
				},
				"MB": {
					Name:      "MB",
					Type:      node.TypeEscaped,
					Delimiter: "[",
				},
				"MC": {
					Name:      "MC",
					Type:      node.TypeUniform,
					Delimiter: "_",
				},
				"A": {
					Name:      "A",
					Type:      node.TypeHanging,
					Delimiter: "*",
				},
				"B": {
					Name:      "B",
					Type:      node.TypeWalled,
					Delimiter: ">",
				},
				"MT": {
					Name: "MT",
					Type: node.TypeText,
				},
			})
		})
	}
}

func TestPrefixed(t *testing.T) {
	cases := []struct {
		in  string
		out []node.Node
	}{
		{
			"^^",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Prefixed{"MA", nil},
				}},
			},
		},
		{
			"a^^",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a")},
					&node.Prefixed{"MA", nil},
				}},
			},
		},
		{
			"^^a",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Prefixed{"MA", nil},
					&node.Text{"MT", []byte("a")},
				}},
			},
		},

		// url matcher
		{
			"http://",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Prefixed{"MB", nil},
				}},
			},
		},
		{
			"ahttp://",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a")},
					&node.Prefixed{"MB", nil},
				}},
			},
		},
		{
			"http://a",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Prefixed{"MB", []byte("a")},
				}},
			},
		},
		{
			"http://a.b",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Prefixed{"MB", []byte("a.b")},
				}},
			},
		},
		{
			"http://a.b/c",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Prefixed{"MB", []byte("a.b/c")},
				}},
			},
		},
		{
			"(http://a.b/c)",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("(")},
					&node.Prefixed{"MB", []byte("a.b/c")},
					&node.Text{"MT", []byte(")")},
				}},
			},
		},

		// nested url matcher
		{
			"**http://",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Uniform{"MC", []node.Inline{
						&node.Prefixed{"MB", nil},
					}},
				}},
			},
		},
		{
			"**http**://",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Uniform{"MC", []node.Inline{
						&node.Text{"MT", []byte("http")},
					}},
					&node.Text{"MT", []byte("://")},
				}},
			},
		},
		{
			"**http://**",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Uniform{"MC", []node.Inline{
						&node.Prefixed{"MB", nil},
					}},
				}},
			},
		},
		{
			"http**://",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("http")},
					&node.Uniform{"MC", []node.Inline{
						&node.Text{"MT", []byte("://")},
					}},
				}},
			},
		},
		{
			"http://**", // domain cannot contain "*"
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Prefixed{"MB", nil},
					&node.Uniform{"MC", nil},
				}},
			},
		},
		{
			"http://a.b/**",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Prefixed{"MB", []byte("a.b/")},
					&node.Uniform{"MC", nil},
				}},
			},
		},

		// escape
		{
			`\^^`,
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("^^")},
				}},
			},
		},
		{
			`\http://`,
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("http://")},
				}},
			},
		},
		{
			`\http`,
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte(`\http`)},
				}},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			test(t, c.in, c.out, nil, parser.ElementMap{
				"T": {
					Name: "T",
					Type: node.TypeLeaf,
				},
				"MA": {
					Name:      "MA",
					Type:      node.TypePrefixed,
					Delimiter: "^^",
				},
				"MB": {
					Name:      "MB",
					Type:      node.TypePrefixed,
					Delimiter: "http://",
					Matcher:   "url",
				},
				"MC": {
					Name:      "MC",
					Type:      node.TypeUniform,
					Delimiter: "*",
				},
				"MD": {
					Name:      "MD",
					Type:      node.TypeUniform,
					Delimiter: "_",
				},
				"MT": {
					Name: "MT",
					Type: node.TypeText,
				},
			})
		})
	}
}

func TestPrecedence(t *testing.T) {
	cases := []struct {
		in  string
		out []node.Node
	}{
		// hanging
		{
			"*",
			[]node.Node{&node.Hanging{"A", nil}},
		},
		{
			"**",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Uniform{"MA", nil},
				}},
			},
		},
		{
			"* *",
			[]node.Node{
				&node.Hanging{"A", []node.Block{
					&node.Hanging{"A", nil},
				}},
			},
		},

		// across line
		{
			"a**\n*",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a")},
					&node.Uniform{"MA", nil},
				}},
				&node.Hanging{"A", nil},
			},
		},
		{
			"a**\n**b",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a")},
					&node.Uniform{"MA", []node.Inline{&node.Text{"MT", []byte(" ")}}},
					&node.Text{"MT", []byte("b")},
				}},
			},
		},
		{
			"a**\n* *",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a")},
					&node.Uniform{"MA", nil},
				}},
				&node.Hanging{"A", []node.Block{
					&node.Hanging{"A", nil},
				}},
			},
		},

		// fenced
		{
			"`",
			[]node.Node{&node.Fenced{"B", nil, nil}},
		},
		{
			"``",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Escaped{"MB", nil},
				}},
			},
		},
		{
			"` `",
			[]node.Node{
				&node.Fenced{"B", [][]byte{[]byte(" `")}, nil},
			},
		},

		// across line
		{
			"a``\n`",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a")},
					&node.Escaped{"MB", nil},
				}},
				&node.Fenced{"B", nil, nil},
			},
		},
		{
			"a``\n``b",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a")},
					&node.Escaped{"MB", []byte("")},
					&node.Text{"MT", []byte("b")},
				}},
			},
		},
		{
			"a``\\\n``b",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a")},
					&node.Escaped{"MB", []byte(" ``b")},
				}},
			},
		},
		{
			"a``\\\n\\``b",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a")},
					&node.Escaped{"MB", []byte("")},
					&node.Text{"MT", []byte("b")},
				}},
			},
		},
		{
			"a``\n` `",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a")},
					&node.Escaped{"MB", nil},
				}},
				&node.Fenced{"B", [][]byte{[]byte(" `")}, nil},
			},
		},

		// ranked hanging - should never have a same delimiter as any
		// inline element as there is no way to escape inline precedence
		{
			"==",
			[]node.Node{
				&node.RankedHanging{"C", 2, nil},
			},
		},
		{
			"===",
			[]node.Node{
				&node.RankedHanging{"C", 3, nil},
			},
		},
		{
			"= =",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("= =")},
				}},
			},
		},

		// longer block delimiter
		{
			"-",
			[]node.Node{&node.Hanging{"D", nil}},
		},
		{
			"- -",
			[]node.Node{&node.Hanging{"D", []node.Block{
				&node.Hanging{"D", nil},
			}}},
		},
		{
			"--",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Uniform{"MD", nil},
				}},
			},
		},
		{
			"---",
			[]node.Node{&node.VerbatimLine{"DLong", nil}},
		},
	}

	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			test(t, c.in, c.out, nil, parser.ElementMap{
				"T": {
					Name: "T",
					Type: node.TypeLeaf,
				},
				"A": {
					Name:      "A",
					Type:      node.TypeHanging,
					Delimiter: "*",
				},
				"MA": {
					Name:      "MA",
					Type:      node.TypeUniform,
					Delimiter: "*",
				},

				"B": {
					Name:      "B",
					Type:      node.TypeFenced,
					Delimiter: "`",
				},
				"MB": {
					Name:      "MB",
					Type:      node.TypeEscaped,
					Delimiter: "`",
				},

				"C": {
					Name:      "C",
					Type:      node.TypeRankedHanging,
					Delimiter: "=",
				},

				"D": {
					Name:      "D",
					Type:      node.TypeHanging,
					Delimiter: "-",
				},
				"DLong": {
					Name:      "DLong",
					Type:      node.TypeVerbatimLine,
					Delimiter: "---",
				},
				"MD": {
					Name:      "MD",
					Type:      node.TypeUniform,
					Delimiter: "-",
				},
				"MT": {
					Name: "MT",
					Type: node.TypeText,
				},
			})
		})
	}
}

func TestEscape(t *testing.T) {
	cases := []struct {
		in  string
		out []node.Node
	}{
		{
			`\`,
			[]node.Node{
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte(`\`)}}},
			},
		},
		{
			`\\`,
			[]node.Node{
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte(`\`)}}},
			},
		},
		{
			`\\\`,
			[]node.Node{
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte(`\\`)}}},
			},
		},
		{
			`\\\\`,
			[]node.Node{
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte(`\\`)}}},
			},
		},
		{
			"\\\n\\",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte(`\ \`)}}},
			},
		},
		{
			"\\a\n\\b",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte(`\a \b`)}}},
			},
		},

		{
			`\a`,
			[]node.Node{
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte(`\a`)}}},
			},
		},
		{
			`\\a`,
			[]node.Node{
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte(`\a`)}}},
			},
		},

		{
			`\**`,
			[]node.Node{
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("**")}}},
			},
		},
		{
			`\\**`,
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte(`\`)},
					&node.Uniform{"MA", nil},
				}},
			},
		},
		{
			`\\\**`,
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte(`\**`)},
				}},
			},
		},

		{
			`a\**`,
			[]node.Node{
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a**")}}},
			},
		},
		{
			`\**\**`,
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("****")},
				}},
			},
		},

		{
			`\!`,
			[]node.Node{
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("!")}}},
			},
		},
		{
			`\\!`,
			[]node.Node{
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte(`\!`)}}},
			},
		},
		{
			`\\\!`,
			[]node.Node{
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte(`\!`)}}},
			},
		},

		// in verbatim
		{
			"`\n\\\\",
			[]node.Node{
				&node.Fenced{"B", [][]byte{nil, []byte(`\\`)}, nil},
			},
		},
		{
			"``a\\\\",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Escaped{"MB", []byte(`a\\`)},
				}},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			test(t, c.in, c.out, nil, parser.ElementMap{
				"T": {
					Name: "T",
					Type: node.TypeLeaf,
				},
				"A": {
					Name:      "A",
					Type:      node.TypeHanging,
					Delimiter: "*",
				},
				"B": {
					Name:      "B",
					Type:      node.TypeFenced,
					Delimiter: "`",
				},
				"MA": {
					Name:      "MA",
					Type:      node.TypeUniform,
					Delimiter: "*",
				},
				"MB": {
					Name:      "MB",
					Type:      node.TypeEscaped,
					Delimiter: "`",
				},
				"MT": {
					Name: "MT",
					Type: node.TypeText,
				},
			})
		})
	}
}

func TestLineBreak(t *testing.T) {
	// test one-character prefixed element, the character being the escape
	// "\"

	cases := []struct {
		in  string
		out []node.Node
	}{
		{
			`\`,
			[]node.Node{
				&node.Leaf{"T", []node.Inline{&node.Prefixed{"MA", nil}}},
			},
		},
		{
			`\\`,
			[]node.Node{
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte(`\`)}}},
			},
		},
		{
			`\\\`,
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte(`\`)},
					&node.Prefixed{"MA", nil},
				}},
			},
		},
		{
			`\\\\`,
			[]node.Node{
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte(`\\`)}}},
			},
		},
		{
			"\\\n\\",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Prefixed{"MA", nil},
					&node.Text{"MT", []byte(" ")},
					&node.Prefixed{"MA", nil},
				}},
			},
		},
		{
			"\\a\n\\b",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Prefixed{"MA", nil},
					&node.Text{"MT", []byte("a ")},
					&node.Prefixed{"MA", nil},
					&node.Text{"MT", []byte("b")},
				}},
			},
		},

		{
			`\a`,
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Prefixed{"MA", nil},
					&node.Text{"MT", []byte("a")},
				}},
			},
		},
		{
			`\\a`,
			[]node.Node{
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte(`\a`)}}},
			},
		},

		{
			`\**`,
			[]node.Node{
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("**")}}},
			},
		},
		{
			`\\**`,
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte(`\`)},
					&node.Uniform{"MB", nil},
				}},
			},
		},
		{
			`\\\**`,
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte(`\**`)},
				}},
			},
		},

		{
			`a\**`,
			[]node.Node{
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("a**")}}},
			},
		},
		{
			`\**\**`,
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("****")},
				}},
			},
		},

		{
			`\!`,
			[]node.Node{
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte("!")}}},
			},
		},
		{
			`\\!`,
			[]node.Node{
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte(`\!`)}}},
			},
		},
		{
			`\\\!`,
			[]node.Node{
				&node.Leaf{"T", []node.Inline{&node.Text{"MT", []byte(`\!`)}}},
			},
		},

		// in verbatim
		{
			"`\n\\\\",
			[]node.Node{
				&node.Fenced{"A", [][]byte{nil, []byte(`\\`)}, nil},
			},
		},
		{
			"``a\\\\",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Escaped{"MC", []byte(`a\\`)},
				}},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			test(t, c.in, c.out, nil, parser.ElementMap{
				"T": {
					Name: "T",
					Type: node.TypeLeaf,
				},
				"A": {
					Name:      "A",
					Type:      node.TypeFenced,
					Delimiter: "`",
				},
				"MA": {
					Name:      "MA",
					Type:      node.TypePrefixed,
					Delimiter: `\`,
				},
				"MB": {
					Name:      "MB",
					Type:      node.TypeUniform,
					Delimiter: "*",
				},
				"MC": {
					Name:      "MC",
					Type:      node.TypeEscaped,
					Delimiter: "`",
				},
				"MT": {
					Name: "MT",
					Type: node.TypeText,
				},
			})
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
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte(string(unicode.ReplacementChar) + "a")},
				},
				}},
		},
		{
			"in the middle",
			"a" + fcb + "b",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a" + string(unicode.ReplacementChar) + "b")},
				},
				}},
		},
		{
			"in the end",
			"a" + fcb,
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a" + string(unicode.ReplacementChar))},
				},
				}},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			test(t, c.in, c.out, []error{parser.ErrInvalidUTF8Encoding}, parser.ElementMap{
				"T": {
					Name: "T",
					Type: node.TypeLeaf,
				},
				"MT": {
					Name: "MT",
					Type: node.TypeText,
				},
			})
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
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte(string(unicode.ReplacementChar) + "a")},
				},
				}},
		},
		{
			"in the middle",
			"a" + null + "b",
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a" + string(unicode.ReplacementChar) + "b")},
				},
				}},
		},
		{
			"in the end",
			"a" + null,
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a" + string(unicode.ReplacementChar))},
				},
				}},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			test(t, c.in, c.out, []error{parser.ErrIllegalNULL}, parser.ElementMap{
				"T": {
					Name: "T",
					Type: node.TypeLeaf,
				},
				"MT": {
					Name: "MT",
					Type: node.TypeText,
				},
			})
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
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a")},
				}},
			},
			nil,
			parser.ElementMap{
				"T": {
					Name: "T",
					Type: node.TypeLeaf,
				},
				"MT": {
					Name: "MT",
					Type: node.TypeText,
				},
			},
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
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a" + string(unicode.ReplacementChar) + "b")},
				}},
			},
		},
		{
			"in the end",
			"a" + bom,
			[]node.Node{
				&node.Leaf{"T", []node.Inline{
					&node.Text{"MT", []byte("a" + string(unicode.ReplacementChar))},
				}},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			test(t, c.in, c.out, []error{parser.ErrIllegalBOM}, parser.ElementMap{
				"T": {
					Name: "T",
					Type: node.TypeLeaf,
				},
				"MT": {
					Name: "MT",
					Type: node.TypeText,
				},
			})
		})
	}
}

// test compares the string representation of nodes generated by the parser from
// the argument in and the nodes of the argument out. Expected error must be
// encountered once; test calls t.Error() if it is encountered multiple times or
// if it is never encountered.
//
// Note on custom element naming:
// Use uppercase characters and prefix inline elements with M.
func test(t *testing.T, in string, out []node.Node, expectedErrors []error, elementMap parser.ElementMap) {
	r := strings.NewReader(in)
	nodes, errs := parser.Parse(r, elementMap)

	if expectedErrors == nil {
		for _, err := range errs {
			t.Errorf("got error %q", err)
		}
	} else {
		unencountered := expectedErrors

		for _, err := range errs {
			if i := errorIndex(unencountered, err); i > -1 {
				// expected error, remove it from unencounterd
				unencountered = append(unencountered[:i], unencountered[i+1:]...)
			} else {
				t.Errorf("got error %q", err)
			}
		}

		// unencountered expected errors
		for _, err := range unencountered {
			t.Errorf("want error %q", err)
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
