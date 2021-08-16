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
		{"a\nb", "a b"},

		// verbatim line
		{".image", ""},
		{".imagea", ".image a"},
		{".imagea ", ".image a"},

		// hanging
		{"*", ""},
		{"*a", "* a"},
		{"*\n a", "* a"},
		{"*a\n b", "* a b"},
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

		{">*", ""},
		{">*a", "> * a"},
		{">\n>*", ""},
		{">\n>*a", "> * a"},

		{"`*", "`*\n`"},
		{"`*a", "`*a\n`"},
		{"`\n *", "`\n *\n`"},
		{"`\n *a", "`\n *a\n`"},

		{"-*", ""},
		{"-*a", "- * a"},
		{"-\n *", ""},
		{"-\n *a", "- * a"},

		// ranked hanging
		{"=", ""},
		{"=a", "= a"},
		{"=a\n b", "= a b"},
		{"==", ""},
		{"==a", "== a"},
		{"==a\n  b", "== a b"},
		{"==a\n\n  b", "== a\n\n   b"},
		{"==a\n \n  b", "== a\n\n   b"},
		{"==a\n\n\n  b", "== a\n\n   b"},

		// walled
		{">", ""},
		{">a", "> a"},
		{">\n>", ""},
		{">a\n>b", "> a b"},
		{">a\n>\n>b", "> a\n>\n> b"},
		{">a\n> \n>b", "> a\n>\n> b"},
		{">a\n>\n>\n>b", "> a\n>\n> b"},

		{">`a\n>b", "> `a\n> b\n> `"},

		// verbatim walled
		{"/", ""},
		{"/a", "/a"},
		{"/ a", "/ a"},
		{"/\n/", ""},
		{"/a\n/b", "/a\n/b"},
		{"/ a\n/ b", "/ a\n/ b"},
		{"/a\n/\n/b", "/a\n/\n/b"},

		// fenced
		{"`", ""},
		{"`a", "`a\n`"},
		{"`a`", "`a`\n`"},
		{"`a\nb", "`a\nb\n`"},
		{"`a\n ", "`a\n \n`"},
		{"`a\n\nb", "`a\n\nb\n`"},
		{"`a\n \nb", "`a\n \nb\n`"},
		{"`a\n\n\nb", "`a\n\n\nb\n`"},

		{"`\\a", "`a\n`"},
		{"`\\\n`", "`\\\n`\n\\`"},

		// groups
		{"-", ""},
		{"-a", "- a"},
		{"-a\n-b", "- a\n- b"},
		{"-a\n -b\n -c", "- a\n\n  - b\n  - c"},

		// sticky
		{"/a\nb", "/a\nb"},
		{"/a\n/b\nc", "/a\n/b\nc"},
		{"a\n/b", "a\n\n/b"},
		{"a\n/b\n/c", "a\n\n/b\n/c"},

		{"a\n+b", "a\n+ b"},
		{"a\n+b\n+c", "a\n+ b c"},
		{"+a\nb", "+ a\n\nb"},
		{"+a\n+b\nc", "+ a b\n\nc"},

		// multiple blocks
		{"a\n\nb", "a\n\nb"},
		{"a\n\n\nb", "a\n\nb"},
		{"a\n \nb", "a\n\nb"},
		{"a\n*b", "a\n\n* b"},
		{"*a\nb", "* a\n\nb"},
		{"*a\n*b", "* a\n* b"},
		{"*a\n>b", "* a\n\n> b"},
		{">a\n\n>b", "> a\n\n> b"},
		{">a\n \n>b", "> a\n\n> b"},
		{">a\n\n\n>b", "> a\n\n> b"},
		{">a\n*b", "> a\n\n* b"},

		// uniform
		{"a__", "a"},
		{"a__b", "a__b__"},
		{"a``__b", "a``__b``"},

		// escaped
		{"a``", "a"},
		{"a``b", "a``b``"},
		{"a`` b", "a`` b``"},
		{"a``b ", "a``b ``"},
		{"a``b``", "a``b``"},
		{"a``\\b``", "a``\\b``\\``"},
		{"a``\\b``\\``", "a``\\b``\\``"},
		{"a``\\b\\``", "a``b``"},

		{"a__``b", "a__``b``__"},

		{"a//b", "a//b//"},
		{"a// b", "a// b//"},
		{"a //b", "a //b//"},
		{"a //b// c", "a //b// c"},

		// composite
		{"(())[[]]", ""},
		{"((a))[[b]]", "((a))[[b]]"},
		{"a((b))[[c]]", "a((b))[[c]]"},
		{"(( a ))[[b]]", "(( a ))[[b]]"},
		{"((a))[[ b ]]", "((a))[[ b ]]"},
		{"((((c))[[d]]a))[[b]]", "((((c))[[d]]a))[[b]]"},

		{"((``a))[[b]]", "((``a))[[b]]``))"},
		{"((``a``))[[b]]", "((``a``))[[b]]"},

		// escape
		{`\`, `\`},
		{`\\`, `\`},
		{`\\\`, `\\\`},
		{`\\\\`, `\\\`},

		{"a", "a"},
		{`\a`, `\a`},
		{`\\a`, `\a`},
		{`\\\a`, `\\\a`},
		{`\\\\a`, `\\\a`},

		{`a\`, `a\`},
		{`a\\`, `a\`},
		{`a\\\`, `a\\\`},
		{`a\\\\`, `a\\\`},

		{"*a", "* a"},
		{`\*`, `\*`},
		{`\\*`, `\\*`},
		{`\\\*`, `\\*`},
		{`\\\\*`, `\\\*`},

		{"\\*\n\\*", `\* *`},

		{"**a", "**a**"},
		{`\**`, `\**`},
		{`\\**`, `\`},
		{`\\**a`, `\\**a**`},
		{`\\\**`, `\\\**`},
		{`\\\**a`, `\\\**a`},
		{`\\\\**a`, `\\\\**a**`},

		{`a\**`, `a\**`},
		{`a\\**`, `a\`},
		{`a\\\**`, `a\\\**`},
		{`a\\\\**`, `a\\\`},

		{`a\***`, `a*`},
		{`a\***b`, `a\***b**`},
		{`a\*\**`, `a\*\**`},
		{`a\*\*\*`, `a\*\**`},
		{`a\*\*\**`, `a\*\*\**`},
		{`a\*\*\*\*`, `a\*\*\**`},

		// nested escape
		{">*a", "> * a"},
		{`>\*`, `> \*`},
		{`>\\*`, `> \\*`},
		{`>\\\*`, `> \\*`},
		{`>\\\\*`, `> \\\*`},

		// escaped in verbatim
		{"`\n\\\\", "`\n\\\\\n`"},
		{"``a\\\\", "``a\\\\``"},

		// escape common elements
		{"a\\``", "a\\``"},
		{`a\[[`, `a\[[`},
		{`a\//`, `a\//`},
		{`a\(())[[]]`, `a\(())`},
		{`a(())\[[]]`, `a\[[]]`},
		{`a\((a))[[b]]`, `a\((a))[[b]]`},
		{`a((a))\[[b]]`, `a((a))\[[b]]`},
		{`a\((a))\[[b]]`, `a\((a))\[[b]]`},
	}

	for _, c := range cases {
		name := strings.ReplaceAll(c.in, "/", "2F") // %2F is URL-escaped slash

		t.Run(fmt.Sprintf("%q", name), func(t *testing.T) {
			blocks, errs := parser.Parse([]byte(c.in))
			if errs != nil {
				t.Fatal(errs)
			}

			nodes := node.BlocksToNodes(blocks)

			conf := config.Default
			nodes = transformer.Paragraph(nodes)
			nodes = transformer.Group(conf.Groups, nodes)
			nodes = transformer.Sequence(nodes)
			nodes = transformer.Composite(conf.Composites, nodes)
			nodes = transformer.GroupStickies(conf.Stickies, nodes)

			var b strings.Builder
			printer.Fprint(&b, conf, nodes)

			printed := b.String()

			if printed != c.out {
				t.Errorf("got %q, want %q", printed, c.out)
			}
		})
	}
}
