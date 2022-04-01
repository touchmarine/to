package parser_test

import (
	"strings"
	"testing"

	"github.com/touchmarine/to/node"
	"github.com/touchmarine/to/parser"
	"github.com/touchmarine/to/printer"
	"golang.org/x/net/html/atom"
)

func TestParseFromHTML(t *testing.T) {
	hels := map[atom.Atom]string{
		atom.B: "A",
	}
	els := parser.Elements{
		"A": parser.Element{
			Name:      "A",
			Type:      node.TypeUniform,
			Delimiter: "*",
		},
	}
	cases := []struct {
		in, to string
	}{
		{
			in: "a",
			to: "a",
		},
		{
			in: "<b>a</b>",
			to: "**a**",
		},
		{
			in: `<div class="A">a</b>`,
			to: "**a**",
		},
		{
			in: "<p>a</p>",
			to: "a",
		},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			n, err := parser.ParseFromHTML(els, hels, []byte(c.in))
			if err != nil {
				t.Fatalf("parse from html: %v", err)
			}
			y := print(t, els, n)

			p := parser.Parser{
				Elements: els,
				TabWidth: 8,
			}
			m, err := p.Parse(nil, []byte(c.to))
			if err != nil {
				t.Fatalf("parse To: %v", err)
			}
			x := print(t, els, m)

			if y != x {
				t.Errorf("\ngot:\n%s\nwant:\n%s\n---\ngot tree:\n%s\nwant tree:\n%s\n",
					y, x, printTree(t, n), printTree(t, m))
			}
		})
	}
}

func printTree(t *testing.T, n *node.Node) string {
	t.Helper()
	var b strings.Builder
	if err := node.Fprint(&b, n); err != nil {
		t.Fatalf("print node tree: %v", err)
	}
	return b.String()
}

func print(t *testing.T, els parser.Elements, n *node.Node) string {
	t.Helper()
	var b strings.Builder
	if err := (printer.Printer{Elements: els}).Fprint(&b, n); err != nil {
		t.Fatalf("print node: %v", err)
	}
	return b.String()
}
