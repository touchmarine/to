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
			[]tl{{token.VLINE, "|"}},
			[]node.Block{
				&node.Paragraph{},
			},
		},
		{
			[]tl{{token.VLINE, "|"}, {token.VLINE, "|"}},
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
				{token.VLINE, "|"},
				{token.VLINE, "|"},
				{token.TEXT, "a"},
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
				{token.VLINE, "|"},
				{token.LINEFEED, "\n"},
				{token.VLINE, "|"},
				{token.TEXT, "a"},
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
				{token.VLINE, "|"},
				{token.TEXT, "a"},
				{token.LINEFEED, "\n"},
				{token.VLINE, "|"},
				{token.TEXT, "b"},
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
				{token.VLINE, "|"},
				{token.TEXT, "a"},
				{token.LINEFEED, "\n"},
				{token.VLINE, "|"},
				{token.VLINE, "|"},
				{token.TEXT, "b"},
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
		{
			[]tl{
				{token.VLINE, "|"},
				{token.TEXT, "a"},
				{token.LINEFEED, "\n"},
				{token.VLINE, "|"},
				{token.VLINE, "|"},
				{token.TEXT, "b"},
				{token.LINEFEED, "\n"},
				{token.VLINE, "|"},
				{token.VLINE, "|"},
				{token.VLINE, "|"},
				{token.TEXT, "c"},
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
								&node.Paragraph{
									[]node.Block{
										node.Lines{
											"c",
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			[]tl{
				{token.VLINE, "|"},
				{token.TEXT, "a"},
				{token.LINEFEED, "\n"},
				{token.TEXT, "b"},
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
				{token.VLINE, "|"},
				{token.VLINE, "|"},
				{token.TEXT, "a"},
				{token.LINEFEED, "\n"},
				{token.TEXT, "b"},
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
		{
			[]tl{
				{token.VLINE, "|"},
				{token.VLINE, "|"},
				{token.TEXT, "a"},
				{token.LINEFEED, "\n"},
				{token.VLINE, "|"},
				{token.TEXT, "b"},
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
				{token.VLINE, "|"},
				{token.VLINE, "|"},
				{token.VLINE, "|"},
				{token.TEXT, "a"},
				{token.LINEFEED, "\n"},
				{token.VLINE, "|"},
				{token.VLINE, "|"},
				{token.TEXT, "b"},
				{token.LINEFEED, "\n"},
				{token.VLINE, "|"},
				{token.TEXT, "c"},
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
			[]tl{{token.GT, ">"}},
			[]node.Block{
				&node.Blockquote{},
			},
		},
		{
			[]tl{{token.GT, ">"}, {token.GT, ">"}},
			[]node.Block{
				&node.Blockquote{
					[]node.Block{
						&node.Blockquote{},
					},
				},
			},
		},
		{
			[]tl{
				{token.GT, ">"},
				{token.GT, ">"},
				{token.TEXT, "a"},
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
			[]tl{{token.GT, ">"}, {token.VLINE, "|"}},
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
				{token.GT, ">"},
				{token.VLINE, "|"},
				{token.GT, ">"},
				{token.TEXT, "a"},
				{token.LINEFEED, "\n"},
				{token.TEXT, "b"},
				{token.LINEFEED, "\n"},
				{token.VLINE, "|"},
				{token.TEXT, "c"},
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
		{
			[]tl{
				{token.VLINE, "|"},
				{token.TEXT, "a"},
				{token.LINEFEED, "\n"},
				{token.VLINE, "|"},
				{token.GT, ">"},
				{token.TEXT, "b"},
				{token.LINEFEED, "\n"},
				{token.VLINE, "|"},
				{token.GT, ">"},
				{token.VLINE, "|"},
				{token.TEXT, "c"},
			},
			[]node.Block{
				&node.Paragraph{
					[]node.Block{
						node.Lines{
							"a",
						},
						&node.Blockquote{
							[]node.Block{
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
					},
				},
			},
		},
		{
			[]tl{
				{token.VLINE, "|"},
				{token.TEXT, "a"},
				{token.LINEFEED, "\n"},
				{token.VLINE, "|"},
				{token.GT, ">"},
				{token.LINEFEED, "\n"},
				{token.VLINE, "|"},
				{token.GT, ">"},
				{token.VLINE, "|"},
				{token.TEXT, "c"},
				{token.LINEFEED, "\n"},
				{token.TEXT, "d"},
			},
			[]node.Block{
				&node.Paragraph{
					[]node.Block{
						node.Lines{
							"a",
						},
						&node.Blockquote{
							[]node.Block{
								&node.Paragraph{
									[]node.Block{
										node.Lines{
											"c",
										},
									},
								},
							},
						},
					},
				},
				node.Lines{
					"d",
				},
			},
		},
		{
			[]tl{
				{token.VLINE, "|"},
				{token.GT, ">"},
				{token.LINEFEED, "\n"},
				{token.VLINE, "|"},
				{token.VLINE, "|"},
			},
			[]node.Block{
				&node.Paragraph{
					[]node.Block{
						&node.Blockquote{},
						&node.Paragraph{},
					},
				},
			},
		},
		{
			[]tl{
				{token.GT, ">"},
				{token.VLINE, "|"},
				{token.GT, ">"},
				{token.VLINE, "|"},
				{token.LINEFEED, "\n"},
				{token.GT, ">"},
				{token.VLINE, "|"},
				{token.GT, ">"},
				{token.VLINE, "|"},
				{token.TEXT, "a"},
				{token.LINEFEED, "\n"},
				{token.GT, ">"},
				{token.VLINE, "|"},
				{token.GT, ">"},
				{token.VLINE, "|"},
				{token.GT, ">"},
				{token.LINEFEED, "\n"},
				{token.GT, ">"},
				{token.TEXT, "b"},
			},
			[]node.Block{
				&node.Blockquote{
					[]node.Block{
						&node.Paragraph{
							[]node.Block{
								&node.Blockquote{
									[]node.Block{
										&node.Paragraph{
											[]node.Block{
												node.Lines{
													"a",
												},
												&node.Blockquote{},
											},
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
			},
		},
		{
			// pyramid
			// a
			// |b
			// |>c
			// |>|d
			// |>e
			// |f
			// g
			[]tl{
				{token.TEXT, "a"},
				{token.LINEFEED, "\n"},
				{token.VLINE, "|"},
				{token.TEXT, "b"},
				{token.LINEFEED, "\n"},
				{token.VLINE, "|"},
				{token.GT, ">"},
				{token.TEXT, "c"},
				{token.LINEFEED, "\n"},
				{token.VLINE, "|"},
				{token.GT, ">"},
				{token.VLINE, "|"},
				{token.TEXT, "d"},
				{token.LINEFEED, "\n"},
				{token.VLINE, "|"},
				{token.GT, ">"},
				{token.TEXT, "e"},
				{token.LINEFEED, "\n"},
				{token.VLINE, "|"},
				{token.TEXT, "f"},
				{token.LINEFEED, "\n"},
				{token.TEXT, "g"},
			},
			[]node.Block{
				node.Lines{"a"},
				&node.Paragraph{
					[]node.Block{
						node.Lines{"b"},
						&node.Blockquote{
							[]node.Block{
								node.Lines{"c"},
								&node.Paragraph{
									[]node.Block{
										node.Lines{"d"},
									},
								},
								node.Lines{"e"},
							},
						},
						node.Lines{"f"},
					},
				},
				node.Lines{"g"},
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
