package printer_test

import (
	"fmt"
	"github.com/touchmarine/to/config"
	"github.com/touchmarine/to/node"
	"github.com/touchmarine/to/parser"
	"github.com/touchmarine/to/printer"
	"github.com/touchmarine/to/transformer"
	"strings"
	"testing"
)

func TestFprint(t *testing.T) {
	cases := []struct {
		in  string
		out string
	}{
		{"", ""},
		{"a", "a"},

		{"\na", "a"},
		{"a\nb", "a\nb"},

		{"a\n\nb", "a\n\nb"},
		{">a\n\n>b", "> a\n\n> b"},
		{">a\n*b", "> a\n\n* b"},

		// blocks
		{"*", "*"},
		{"*a", "* a"},
		{">", ">"},
		{">a", "> a"},
		{"``", "``\n``"},
		{"``a", "``a\n``"},
		{"=", "="},
		{"=a", "= a"},
		{"==", "=="},
		{"==a", "== a"},
		{".image", ".image"},
		{".imagea", ".image a"},

		{"``", "``\n``"},
		{"``a\nb", "``a\nb\n``"},

		// hanging indentation
		{"*\n a", "* a"},
		{"*a\n b", "* a\n  b"},
		{"*a\n *b", "* a\n  * b"},

		{"=a\n\n>b", "= a\n\n> b"},
		{".imagea\n      b", ".image a\n       b"},

		// groups
		{"-", "-"},

		// inlines
		{"a__", "a____"},
		{"a__b", "a__b__"},
		{"a``", "a````"},
		{"a``b", "a``b``"},
		{"a<>", "a<>"},
		{"a<b>", "a<b>"},
		{"a<><>", "a<>"},
		{"a<b><c>", "a<b><c>"},

		// block-escape inlines
		{`\**`, `\****`},
		{`\**a`, `\**a**`},

		{"a**", "a****"},
		{`a\*`, "a*"},
		{`a\**`, `a\**`},
		{`\\**`, `\\**`},
		{`.toc \==a`, `.toc \==a`},

		{`a\**b`, `a\**b`},
		{`a\\**b`, `a\\**b**`},

		// escaped inlines
		{"a``", "a````"},
		{"a`{", "a````"},
		{"a`{``", "a`{``}`"},
		{"a`{```{", "a`[```{]`"},
		{"a`[```{`[", "a`(```{`[)`"},

		// comments
		{"//", "//"},
		{"//a", "// a"},
		{"// a", "//  a"},

		// hats
		{"%h", "% h"},
		{"% h", "% h"},
		{"%h\n\n", "% h"},
		{"%h\na", "% h\na"},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("%q", c.in), func(t *testing.T) {
			r := strings.NewReader(c.in)
			blocks, errs := parser.Parse(r)
			if errs != nil {
				t.Fatal(errs)
			}

			nodes := node.BlocksToNodes(blocks)

			conf := config.Default
			nodes = transformer.Paragraph(nodes)
			nodes = transformer.Group(conf.Groups, nodes)
			nodes = transformer.Sequence(conf.Elements, nodes)
			nodes = transformer.BlankLine(nodes)

			var b strings.Builder
			printer.Fprint(&b, conf, nodes)

			printed := b.String()

			if printed != c.out {
				t.Errorf("got %q, want %q", printed, c.out)
			}
		})
	}
}
