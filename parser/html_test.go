package parser_test

import (
	"strings"
	"testing"

	"github.com/touchmarine/to/node"
	"github.com/touchmarine/to/parser"
	"github.com/touchmarine/to/printer"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func TestOutHTML(t *testing.T) {
	hels := map[string]atom.Atom{
		"A": atom.B,
		"B": atom.Blockquote,
	}
	els := parser.Elements{
		"A": parser.Element{
			Name:      "A",
			Type:      node.TypeUniform,
			Delimiter: "*",
		},
		"B": parser.Element{
			Name:      "B",
			Type:      node.TypeWalled,
			Delimiter: ">",
		},
	}
	cases := []struct {
		name, in, out string
		hels          map[string]atom.Atom // overrides default hels
	}{
		{
			name: "a",
			in:   "a",
			out:  "a",
			hels: hels,
		},
		{
			name: "**a**",
			in:   "**a**",
			out:  "<b>a</b>",
			hels: hels,
		},
		{
			name: "**a** no hels",
			in:   "**a**",
			out:  `<span class="A">a</span>`,
			hels: nil,
		},
		{
			name: ">a",
			in:   ">a",
			out:  "<blockquote>a</blockquote>",
			hels: hels,
		},
		{
			name: ">a no hels",
			in:   ">a",
			out:  `<div class="B">a</div>`,
			hels: nil,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := parser.Parser{
				Elements: els,
				TabWidth: 8,
			}
			n, err := p.Parse(nil, []byte(c.in))
			if err != nil {
				t.Fatalf("parse To: %v", err)
			}
			m, err := parser.OutHTML(els, c.hels, n)
			if err != nil {
				t.Fatalf("out html: %v", err)
			}
			h := renderHTML(t, m)
			if h != c.out {
				t.Errorf("\ngot:\n%s\nwant:\n%s", h, c.out)
			}
		})
	}
}

func renderHTML(t *testing.T, n *html.Node) string {
	t.Helper()
	var b strings.Builder
	if err := html.Render(&b, n); err != nil {
		t.Fatalf("render html: %v", err)
	}
	h := b.String()
	h = strings.TrimPrefix(h, "<body>")
	h = strings.TrimSuffix(h, "</body>")
	return h
}

func TestParseFromHTML(t *testing.T) {
	hels := map[atom.Atom]string{
		atom.B:      "A",
		atom.Strong: "A",
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
			in: `<span class="A">a</div>`,
			to: "**a**",
		},
		{
			in: "<strong>a</strong>",
			to: "**a**",
		},
		{
			in: "<p>a</p>",
			to: "a",
		},
		{
			in: "a**",
			to: `a\**`,
		},
		{
			in: `a\**`,
			to: `a\\\**`,
		},

		{
			in: "a\nb",
			to: "a b",
		},
		{
			in: "a\n\nb",
			to: "a b",
		},
		{
			in: "a<br>b",
			to: "a\n\nb",
		},
		{
			in: "<br>",
			to: "",
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
