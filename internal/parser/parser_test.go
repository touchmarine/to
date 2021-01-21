package parser_test

import (
	"strconv"
	"testing"
	"to/internal/node"
	"to/internal/parser"
	"to/internal/printer"
	"to/internal/token"
)

func TestParse_Parse(t *testing.T) {
	cases := []struct {
		tokens []tl
		blocks []node.Block
	}{
		{
			[]tl{{token.Pipeline, "|"}},
			[]node.Block{
				&node.Paragraph{},
			},
		},
		{
			[]tl{{token.Pipeline, "|"}, {token.Pipeline, "|"}},
			[]node.Block{
				&node.Paragraph{
					[]node.Block{
						&node.Paragraph{},
					},
				},
			},
		},
		{
			[]tl{
				{token.Pipeline, "|"},
				{token.Pipeline, "|"},
				{token.Text, "a"},
			},
			[]node.Block{
				&node.Paragraph{
					[]node.Block{
						&node.Paragraph{
							[]node.Block{
								node.Lines{
									"a",
								},
							},
						},
					},
				},
			},
		},
		{
			[]tl{
				{token.Pipeline, "|"},
				{token.Newline, "\n"},
				{token.Pipeline, "|"},
				{token.Text, "a"},
			},
			[]node.Block{
				&node.Paragraph{
					[]node.Block{
						node.Lines{
							"a",
						},
					},
				},
			},
		},
		{
			[]tl{
				{token.Pipeline, "|"},
				{token.Text, "a"},
				{token.Newline, "\n"},
				{token.Pipeline, "|"},
				{token.Text, "b"},
			},
			[]node.Block{
				&node.Paragraph{
					[]node.Block{
						node.Lines{
							"a",
							"b",
						},
					},
				},
			},
		},
		{
			[]tl{
				{token.Pipeline, "|"},
				{token.Text, "a"},
				{token.Newline, "\n"},
				{token.Text, "b"},
			},
			[]node.Block{
				&node.Paragraph{
					[]node.Block{
						node.Lines{
							"a",
						},
					},
				},
				node.Lines{
					"b",
				},
			},
		},
		{
			[]tl{
				{token.Pipeline, "|"},
				{token.Pipeline, "|"},
				{token.Text, "a"},
				{token.Newline, "\n"},
				{token.Text, "b"},
			},
			[]node.Block{
				&node.Paragraph{
					[]node.Block{
						&node.Paragraph{
							[]node.Block{
								node.Lines{
									"a",
								},
							},
						},
					},
				},
				node.Lines{
					"b",
				},
			},
		},
		/*
			{
				[]tl{
					{token.Pipeline, "|"},
					{token.Text, "a"},
					{token.Newline, "\n"},
					{token.Pipeline, "|"},
					{token.Pipeline, "|"},
					{token.Text, "b"},
				},
				[]node.Block{
					&node.Paragraph{
						[]node.Block{
							node.Lines{
								"a",
							},
							&node.Paragraph{
								[]node.Block{
									node.Lines{
										"b",
									},
								},
							},
						},
					},
				},
			},
		*/
		{
			[]tl{
				{token.Pipeline, "|"},
				{token.Pipeline, "|"},
				{token.Text, "a"},
				{token.Newline, "\n"},
				{token.Pipeline, "|"},
				{token.Text, "b"},
			},
			[]node.Block{
				&node.Paragraph{
					[]node.Block{
						&node.Paragraph{
							[]node.Block{
								node.Lines{
									"a",
								},
							},
						},
						node.Lines{
							"b",
						},
					},
				},
			},
		},
		{
			[]tl{
				{token.Pipeline, "|"},
				{token.Pipeline, "|"},
				{token.Pipeline, "|"},
				{token.Text, "a"},
				{token.Newline, "\n"},
				{token.Pipeline, "|"},
				{token.Pipeline, "|"},
				{token.Text, "b"},
				{token.Newline, "\n"},
				{token.Pipeline, "|"},
				{token.Text, "c"},
			},
			[]node.Block{
				&node.Paragraph{
					[]node.Block{
						&node.Paragraph{
							[]node.Block{
								&node.Paragraph{
									[]node.Block{
										node.Lines{
											"a",
										},
									},
								},
								node.Lines{
									"b",
								},
							},
						},
						node.Lines{
							"c",
						},
					},
				},
			},
		},
		{
			[]tl{{token.GreaterThan, ">"}},
			[]node.Block{
				&node.Blockquote{},
			},
		},
		{
			[]tl{{token.GreaterThan, ">"}, {token.GreaterThan, ">"}},
			[]node.Block{
				&node.Blockquote{
					[]node.Block{
						&node.Blockquote{},
					},
				},
			},
		},
		{
			[]tl{{token.GreaterThan, ">"}, {token.Pipeline, "|"}},
			[]node.Block{
				&node.Blockquote{
					[]node.Block{
						&node.Paragraph{},
					},
				},
			},
		},
		{
			[]tl{
				{token.GreaterThan, ">"},
				{token.GreaterThan, ">"},
				{token.Text, "a"},
			},
			[]node.Block{
				&node.Blockquote{
					[]node.Block{
						&node.Blockquote{
							[]node.Block{
								node.Lines{
									"a",
								},
							},
						},
					},
				},
			},
		},
		{
			[]tl{
				{token.GreaterThan, ">"},
				{token.Pipeline, "|"},
				{token.GreaterThan, ">"},
				{token.Text, "a"},
				{token.Newline, "\n"},
				{token.Text, "b"},
				{token.Newline, "\n"},
				{token.Pipeline, "|"},
				{token.Text, "c"},
			},
			[]node.Block{
				&node.Blockquote{
					[]node.Block{
						&node.Paragraph{
							[]node.Block{
								&node.Blockquote{
									[]node.Block{
										node.Lines{
											"a",
										},
									},
								},
							},
						},
					},
				},
				node.Lines{
					"b",
				},
				&node.Paragraph{
					[]node.Block{
						node.Lines{
							"c",
						},
					},
				},
			},
		},
	}

	for _, c := range cases {
		var name string
		for _, pair := range c.tokens {
			name += pair.lit
		}

		t.Run(literal(name), func(t *testing.T) {
			s := &scanner{tokenLiterals: c.tokens}
			p := parser.New(s)
			blocks := p.Parse()

			got := printer.Print(blocks)
			want := printer.Print(c.blocks)

			if got != want {
				t.Errorf("\ngot\n%s\nwant\n%s", got, want)
			}
		})
	}
}

// token-literal pair struct
type tl struct {
	tok token.Token
	lit string
}

type scanner struct {
	i             int  // index
	tokenLiterals []tl // token-literal pairs
}

func (s *scanner) Scan() (token.Token, string) {
	if s.i >= len(s.tokenLiterals) {
		return token.EOF, ""
	}

	p := s.tokenLiterals[s.i]
	s.i++
	return p.tok, p.lit
}

func literal(s string) string {
	q := strconv.Quote(s)
	return q[1 : len(q)-1]
}
