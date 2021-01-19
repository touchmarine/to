package parser_test

import (
	"testing"
	"to/internal/node"
	"to/internal/parser"
	"to/internal/scanner"
)

func TestParse_Parse(t *testing.T) {
	cases := []struct {
		src    string
		blocks []node.Block
	}{
		{
			"|",
			[]node.Block{},
		},
		{
			"||",
			[]node.Block{},
		},
	}

	for _, c := range cases {
		t.Run(c.src, func(t *testing.T) {
			s := scanner.New(
				c.src,
				scanner.ScanComments,
				func(err error, errCount uint) {},
			)

			p := parser.New(s)
			blocks := p.Parse()

			got := node.Pretty(blocks)
			want := node.Pretty(c.blocks)

			if got != want {
				t.Errorf("got %s, want %s\n", got, want)
			}
		})
	}
}
