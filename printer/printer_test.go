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
		// base
		{"", ""},
		{"a", "a"},
		{"a\n", "a"},
		{"\na", "a"},
		{"a\nb", "a\nb"},

		// hanging
		{"*", ""},
		{"*a", "* a"},
		{"*\n a", "* a"},
		{"*a\n b", "* a\n  b"},
		{"*a\n *b", "* a\n\n  * b"},
		{"*a\n\n *b", "* a\n\n  * b"},
		{"*a\n \n *b", "* a\n\n  * b"},
		{"*a\n\n\n *b", "* a\n\n  * b"},

		{"=*", ""},
		{"=*a", "= * a"},
		{"=\n *", ""},
		{"=\n *a", "= * a"},

		{".image*", ".image *"},
		{".image*a", ".image *a"},
		{".image\n      *", ".image *"},
		{".image\n      *a", ".image *a"},

		{">*", ""},
		{">*a", "> * a"},
		{">\n>*", ""},
		{">\n>*a", "> * a"},

		{"``*", "``*\n``"},
		{"``*a", "``*a\n``"},
		{"``\n *", "``\n *\n``"},
		{"``\n *a", "``\n *a\n``"},

		{"-*", ""},
		{"-*a", "- * a"},
		{"-\n *", ""},
		{"-\n *a", "- * a"},

		{"%*", "% *"},
		{"%*a", "% *a"},
		{"%\n*", ""},
		{"%\n*a", "* a"},
		{"%\n\n*", ""},
		{"%\n\n*a", "* a"},

		// ranked hanging
		{"=", ""},
		{"=a", "= a"},
		{"=a\n b", "= a\n  b"},
		{"==", ""},
		{"==a", "== a"},
		{"==a\n  b", "== a\n   b"},
		{"==a\n\n  b", "== a\n\n   b"},
		{"==a\n \n  b", "== a\n\n   b"},
		{"==a\n\n\n  b", "== a\n\n   b"},

		// verbatim hanging
		{".image", ""},
		{".imagea", ".image a"},
		{".imagea\n      b", ".image a\n       b"},
		{".imagea\n\n      b", ".image a\n       b"},
		{".imagea\n \n      b", ".image a\n       b"},

		// walled
		{">", ""},
		{">a", "> a"},
		{">\n>", ""},
		{">a\n>b", "> a\n> b"},
		{">a\n>\n>b", "> a\n>\n> b"},
		{">a\n> \n>b", "> a\n>\n> b"},
		{">a\n>\n>\n>b", "> a\n>\n> b"},

		{">``a\n>b", "> ``a\n> b\n> ``"},

		// fenced
		{"``", ""},
		{"``a", "``a\n``"},
		{"``a``", "``a``\n``"},
		{"``a\nb", "``a\nb\n``"},
		{"``a\n ", "``a\n \n``"},
		{"``a\n\nb", "``a\n\nb\n``"},
		{"``a\n \nb", "``a\n \nb\n``"},
		{"``a\n\n\nb", "``a\n\n\nb\n``"},

		// groups
		{"-", ""},
		{"-a", "- a"},
		{"-a\n-b", "- a\n- b"},
		{"-a\n -b\n -c", "- a\n\n  - b\n  - c"},

		//{"-a<b>c", "- a<b>c"},

		// hats
		{"%", ""},
		{"% ", ""},
		{"%a", "% a"},
		{"% a", "% a"},
		{"%\n%", ""},
		{"%a\n%b", "% a\n% b"},
		{"%a\n\n", "% a"},
		{"%a\nb", "% a\nb"},
		{"%\n\na", "a"},
		{"%a\n\nb", "% a\n\nb"},
		{"%a\n \nb", "% a\n\nb"},
		{"%a\n\n\nb", "% a\n\nb"},

		// multiple blocks
		{"a\n\nb", "a\n\nb"},
		{"a\n\n\nb", "a\n\nb"},
		{"a\n \nb", "a\n\nb"},
		{"a\n*b", "a\n\n* b"},
		{"*a\nb", "* a\n\nb"},
		{"*a\n*b", "* a\n\n* b"},
		{"*a\n>b", "* a\n\n> b"},
		{">a\n\n>b", "> a\n\n> b"},
		{">a\n \n>b", "> a\n\n> b"},
		{">a\n\n\n>b", "> a\n\n> b"},
		{">a\n*b", "> a\n\n* b"},

		// uniform
		{"a__", "a"},
		{"a__b", "a__b__"},
		{"a``__b", "a``__b``"},
		{"a<__b>", "a<__b>"},
		{"a<__b><__c>", "a<__b__><__c>"},

		// escaped
		{"a``", "a"},
		{"a``b", "a``b``"},
		{"a`` b", "a`` b``"},
		{"a``b ", "a``b ``"},
		{"a__``b", "a__``b``__"},
		{"a<``b>", "a<``b>"},
		//{"a<``b><c>", "a<``b><c>``>"},
		//{"a<``b><``c>", "a<``b><``c>"},

		// forward
		{"a<", "a"},
		{"a<>", "a"},
		{"a<b", "a<b>"},
		{"a<b>", "a<b>"},
		{"a< b>", "a< b>"},
		{"a<b >", "a<b >"},
		{"a<<>", "a<<>"},
		{"a<><>", "a"},
		{"a<b><>", "a<b><>"},
		{"a< b><>", "a< b><>"},
		{"a<b ><>", "a<b ><>"},
		{"a<><b>", "a<b>"},
		{"a<b><c>", "a<b><c>"},

		// TODO: {"- <b>", "- <b>"},
		{"- a <b>", "- a <b>"},

		// block escape
		{`\**`, ""},
		{`\**a`, `\**a**`},
		{"\n\\**", ""},
		{"\n\\**a", `\**a**`},
		// TODO: {"a\n\\**", "a"},
		{"a\n\\**a", "a\n\\**a**"},
		{`>\**`, ""},
		{`>\**a`, `> \**a**`},

		// inline escape
		{"a*", "a*"},
		{`a\*`, "a*"},
		{`a\\*`, `a\\*`},
		{"a**", "a"},
		{`a\**`, `a\**`},
		{`a\\**`, `a\\`},
		{`a\**b`, `a\**b`},
		{`a\\**b`, `a\\**b**`},
		{`a\//`, `a\//`},

		{`.toc\==a`, `.toc \==a`},

		// block and inline escape
		// TODO: {`\\`, ""},
		{`\\*`, `\*`},
		{`\\**`, `\\**`},
		{`\\\`, `\\`},
		{`\\\a`, `\\a`},

		// escaped inlines
		{"a``", "a"},
		{"a`{", "a"},
		{"a``b", "a``b``"},
		{"a`{b", "a``b``"},
		{"a`{``", "a`{``}`"},
		{"a`{```{", "a`[```{]`"},
		{"a`[```{`[", "a`(```{`[)`"},

		{"a`[b", "a``b``"},
		{"a`(b", "a``b``"},
		{"a`<b", "a``b``"},
		{"a`}b", "a``b``"},
		{"a`]b", "a``b``"},
		{"a`)b", "a``b``"},
		{"a`>b", "a``b``"},

		// comments
		{"//", ""},
		{"//a", "// a"},
		{"// a", "// a"},
		// TODO: {"//\n//", ""},
		{"//a\n//b", "// a\n// b"},

		// TODO: {"a//", "a"},
		{"a//b", "a // b"},
		{"-\n a//b", "- a // b"},
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