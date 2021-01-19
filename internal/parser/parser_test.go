package parser_test

import (
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
			[]tl{{token.BlockDelim, "|"}},
			[]node.Block{
				&node.Paragraph{},
			},
		},
		{
			[]tl{{token.BlockDelim, "|"}, {token.BlockDelim, "|"}},
			[]node.Block{
				&node.Paragraph{
					[]node.Block{
						&node.Paragraph{},
					},
				},
			},
		},
		{
			[]tl{{token.BlockDelim, "|"}, {token.BlockDelim, "|"},
				{token.Text, "a"}},
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
	}

	for _, c := range cases {
		var name string
		for _, pair := range c.tokens {
			name += pair.lit
		}

		t.Run(name, func(t *testing.T) {
			s := &scanner{tokenLiterals: c.tokens}
			p := parser.New(s)
			blocks := p.Parse()

			print := printer.New()
			got := print.Print(blocks)
			print.Reset()
			want := print.Print(c.blocks)

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
