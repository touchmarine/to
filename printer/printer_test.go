// package printer_test contines tests the printer package.
package printer_test

import (
	"flag"
	"fmt"
	"strings"
	"testing"

	"github.com/touchmarine/to/config"
	"github.com/touchmarine/to/matcher"
	"github.com/touchmarine/to/node"
	"github.com/touchmarine/to/parser"
	"github.com/touchmarine/to/printer"
	"github.com/touchmarine/to/transformer"
	"github.com/touchmarine/to/transformer/group"
	"github.com/touchmarine/to/transformer/paragraph"
	"github.com/touchmarine/to/transformer/sticky"
)

var (
	printTree = flag.Bool("print-tree", false, "print node tree")
	noReprint = flag.Bool("no-reprint", false, "don't run reprint tests")
)

func TestText(t *testing.T) {
	cases := []struct {
		in  string
		out string
	}{
		{"", ""},
		{"a", "a"},
		{"a\n ", "a"},
		{"\na", "a"},

		{"ab", "ab"},
		{"a b", "a b"},
		{"a\nb", "a\nb"},
		{"a\n b", "a\nb"},
		{"a\n\n b", "a\n\nb"},
		{"ab\n c", "ab\nc"},
		{"ab\n\n c", "ab\n\nc"},

		{"a **", "a ****"},

		// interrupted by empty blocks
		{"a\n>\n*\nb", "a\n\n>\n\n*\n\nb"},
		{"a\n>b\n*\nc", "a\n\n> b\n\n*\n\nc"},
		{"a\n>\n*b\nc", "a\n\n>\n\n* b\n\nc"},
	}

	elements := config.Elements{
		"A": {
			Type:      node.TypeWalled.String(),
			Delimiter: ">",
		},
		"B": {
			Type:      node.TypeWalled.String(),
			Delimiter: "*",
		},
		"MA": {
			Type:      node.TypeUniform.String(),
			Delimiter: "*",
		},
	}
	for _, c := range cases {
		name := fmt.Sprintf("%q", c.in)
		t.Run(name, func(t *testing.T) {
			test(t, elements, nil, c.in, c.out, 0)
		})
	}
}

func TestVerbatimLine(t *testing.T) {
	cases := []struct {
		in  string
		out string
	}{
		{".a", ".a"},
		{".A", ".A"},
		{".aa", ".aa"},
		{".aa ", ".aa"},
		{".a a", ".a a"},
		// would be nested-but can only contain verbatim
		{".a>", ".a>"},
		{".a>b", ".a>b"},

		// nested
		{">.a", "> .a"},
		{">.ab", "> .ab"},
	}

	elements := config.Elements{
		"A": {
			Type:      node.TypeVerbatimLine.String(),
			Delimiter: ".a",
		},
		"B": {
			Type:      node.TypeWalled.String(),
			Delimiter: ">",
		},
	}
	for _, c := range cases {
		name := c.in
		t.Run(name, func(t *testing.T) {
			test(t, elements, nil, c.in, c.out, 0)
		})
	}
}

func TestHanging(t *testing.T) {
	cases := []struct {
		in  string
		out string
	}{
		{"-", "-"},
		{"-a", "- a"},
		{"-\n a", "- a"},
		{"-a\n b", "- a\n  b"},
		{"-a\n -b", "- a\n\n  - b"},
		{"-a\n\n -b", "- a\n\n  - b"},
		{"-a\n \n -b", "- a\n\n  - b"},
		{"-a\n\n\n -b", "- a\n\n  - b"},

		// nested
		{"->", "- >"},
		{"->a", "- > a"},
		{"-\n>", "-\n\n>"},
		{"-\n>a", "-\n\n> a"},
		{"-\n >", "- >"},
		{"-\n >a", "- > a"},
		{">-", "> -"},
		{">-a", "> - a"},
		{">\n>-", "> -"},
		{">\n>-a", "> - a"},
	}

	elements := config.Elements{
		"A": {
			Type:      node.TypeHanging.String(),
			Delimiter: "-",
		},
		"B": {
			Type:      node.TypeWalled.String(),
			Delimiter: ">",
		},
	}
	for _, c := range cases {
		name := fmt.Sprintf("%q", c.in)
		t.Run(name, func(t *testing.T) {
			test(t, elements, nil, c.in, c.out, 0)
		})
	}
}

func TestRankedHanging(t *testing.T) {
	cases := []struct {
		in  string
		out string
	}{

		{"==", "=="},
		{"==a", "== a"},
		{"==\n  a", "== a"},
		{"==a\n  b", "== a\n   b"},
		{"==a", "== a"},
		{"==a\n  b", "== a\n   b"},
		{"==a\n\n  b", "== a\n\n   b"},
		{"==a\n \n  b", "== a\n\n   b"},
		{"==a\n\n\n  b", "== a\n\n   b"},

		// nested
		{"==>", "== >"},
		{"==>a", "== > a"},
		{"==\n>", "==\n\n>"},
		{"==\n>a", "==\n\n> a"},
		{"==\n  >", "== >"},
		{"==\n  >a", "== > a"},
		{">==", "> =="},
		{">==a", "> == a"},
		{">\n>==", "> =="},
		{">\n>==a", "> == a"},
	}

	elements := config.Elements{
		"A": {
			Type:      node.TypeRankedHanging.String(),
			Delimiter: "=",
		},
		"B": {
			Type:      node.TypeWalled.String(),
			Delimiter: ">",
		},
	}
	for _, c := range cases {
		name := fmt.Sprintf("%q", c.in)
		t.Run(name, func(t *testing.T) {
			test(t, elements, nil, c.in, c.out, 0)
		})
	}
}

func TestWalled(t *testing.T) {
	cases := []struct {
		in  string
		out string
	}{
		{"+", "+"},
		{"+a", "+ a"},
		{"+\n+a", "+ a"},
		{"+a\n+b", "+ a\n+ b"},
		{"+a\n++b", "+ a\n+\n+ + b"},
		{"+a\n+\n++b", "+ a\n+\n+ + b"},
		{"+a\n++\n++b", "+ a\n+\n+ + b"},
		{"+a\n+\n+\n++b", "+ a\n+\n+ + b"},

		// nested
		{"+>", "+ >"},
		{"+>a", "+ > a"},
		{"+\n>", "+\n\n>"},
		{"+\n>a", "+\n\n> a"},
		{"+\n+>", "+ >"},
		{"+\n+>a", "+ > a"},
		{">+", "> +"},
		{">+a", "> + a"},
		{">\n>+", "> +"},
		{">\n>+a", "> + a"},
	}

	elements := config.Elements{
		"A": {
			Type:      node.TypeWalled.String(),
			Delimiter: "+",
		},
		"B": {
			Type:      node.TypeWalled.String(),
			Delimiter: ">",
		},
	}
	for _, c := range cases {
		name := fmt.Sprintf("%q", c.in)
		t.Run(name, func(t *testing.T) {
			test(t, elements, nil, c.in, c.out, 0)
		})
	}
}

func TestVerbatimWalled(t *testing.T) {
	cases := []struct {
		in  string
		out string
	}{
		{"!", "!"},
		{"!a", "!a"},
		{"!a ", "!a"},
		{"! a ", "! a"},
		{"!\n!a", "!\n!a"},
		{"!a\n!b", "!a\n!b"},
		{"!a\n!\n!b", "!a\n!\n!b"},
		{"!a\n!\n!\n!b", "!a\n!\n!\n!b"},
		// would be nested-but can only contain verbatim
		{"!>", "!>"},
		{"!>a", "!>a"},
		{"!\n>", "!\n\n>"},
		{"!\n>a", "!\n\n> a"},
		{"!\n!>", "!\n!>"},
		{"!\n!>a", "!\n!>a"},

		// nested
		{">!", "> !"},
		{">!a", "> !a"},
		{">\n>!", "> !"},
		{">\n>!a", "> !a"},
	}

	elements := config.Elements{
		"A": {
			Type:      node.TypeVerbatimWalled.String(),
			Delimiter: "!",
		},
		"B": {
			Type:      node.TypeWalled.String(),
			Delimiter: ">",
		},
	}
	for _, c := range cases {
		name := fmt.Sprintf("%q", c.in)
		t.Run(name, func(t *testing.T) {
			test(t, elements, nil, c.in, c.out, 0)
		})
	}
}

func TestFenced(t *testing.T) {
	cases := []struct {
		in  string
		out string
	}{
		{"`", "`\n`"},
		{"`a", "`a\n`"},
		{"`a`", "`a`\n`"},
		{"`a\n", "`a\n`"},
		{"`a\nb", "`a\nb\n`"},
		{"`a\n\nb", "`a\n\nb\n`"},
		{"`a\n \nb", "`a\n \nb\n`"},
		{"`a\n\n\nb", "`a\n\n\nb\n`"},

		// trailing text
		{"`a`b", "`a`b\n`"},
		{"`a\nb\n`", "`a\nb\n`"},
		{"`a\nb\n`c", "`a\nb\n`"},
		{"`a\nb\n`\nc", "`a\nb\n`\n\nc"},

		// escape
		{"``", "``\n`"},
		{"``\nb", "``\nb\n`"},
		{"`a\n`", "`a\n`"},
		{"`a\nb`", "`a\nb`\n`"},

		// unnecessary escape
		{"`\\a\nb", "`a\nb\n`"},
		{"`a\nb\n\\`", "`a\nb\n\\`\n`"},

		// nested
		{">`", "> `\n> `"},
		{">\n>`", "> `\n> `"},
		{">`a\n>b", "> `a\n> b\n> `"},
		{">`a\n>b`", "> `a\n> b`\n> `"},
		{">`a\n>b\n>`", "> `a\n> b\n> `"},
		{">\n>`a\n>b", "> `a\n> b\n> `"},
	}

	elements := config.Elements{
		"A": {
			Type:      node.TypeFenced.String(),
			Delimiter: "`",
		},
		"B": {
			Type:      node.TypeWalled.String(),
			Delimiter: ">",
		},
	}
	for _, c := range cases {
		name := fmt.Sprintf("%q", c.in)
		t.Run(name, func(t *testing.T) {
			test(t, elements, nil, c.in, c.out, 0)
		})
	}
}

func TestGroup(t *testing.T) {
	t.Run("paragraph", func(t *testing.T) {
		cases := []struct {
			in  string
			out string
		}{
			{"a\n\nb", "a\n\nb"},
			{">a\n>\n>b", "> a\n>\n> b"},

			// interrupted by empty blocks
			{"a\n>\n*\nb", "a\n\n>\n\n*\n\nb"},
			{"a\n>b\n*\nc", "a\n\n> b\n\n*\n\nc"},
			{"a\n>\n*b\nc", "a\n\n>\n\n* b\n\nc"},
		}

		elements := config.Elements{
			"A": {
				Type:      node.TypeWalled.String(),
				Delimiter: ">",
			},
			"B": {
				Type:      node.TypeWalled.String(),
				Delimiter: "*",
			},
		}
		transformers := []transformer.Transformer{
			paragraph.Transformer{paragraph.Map{
				"PA": node.TypeLeaf,
			}},
		}
		for _, c := range cases {
			name := fmt.Sprintf("%q", c.in)
			t.Run(name, func(t *testing.T) {
				test(t, elements, transformers, c.in, c.out, 0)
			})
		}
	})

	t.Run("list", func(t *testing.T) {
		cases := []struct {
			in  string
			out string
		}{
			{"-a\n-", "- a\n-"},
			{"-a\n-b", "- a\n- b"},
			{"-a\n\n-b", "- a\n- b"},

			// nested
			{"-a\n-", "- a\n-"},
			{"-a\n-b", "- a\n- b"},

			// interrupted by empty blocks
			{"-a\n>\n-b", "- a\n\n>\n\n- b"},
			{"-a\n>\n\n>\n-b", "- a\n\n>\n\n>\n\n- b"},
			{"-a\n>b\n\n>\n-c", "- a\n\n> b\n\n>\n\n- c"},
			{"-a\n>\n\n>b\n-c", "- a\n\n>\n\n> b\n\n- c"},
			{"-a\n>\n*\n-b", "- a\n\n>\n\n*\n\n- b"},
			{"-a\n>b\n*\n-c", "- a\n\n> b\n\n*\n\n- c"},
			{"-a\n>\n*b\n-c", "- a\n\n>\n\n* b\n\n- c"},
		}

		elements := config.Elements{
			"A": {
				Type:      node.TypeHanging.String(),
				Delimiter: "-",
			},
			"B": {
				Type:      node.TypeWalled.String(),
				Delimiter: ">",
			},
			"C": {
				Type:      node.TypeWalled.String(),
				Delimiter: "*",
			},
		}
		transformers := []transformer.Transformer{
			group.Transformer{group.Map{
				"LA": "A",
			}},
		}
		for _, c := range cases {
			name := fmt.Sprintf("%q", c.in)
			t.Run(name, func(t *testing.T) {
				test(t, elements, transformers, c.in, c.out, 0)
			})
		}
	})

	t.Run("sticky", func(t *testing.T) {
		cases := []struct {
			in  string
			out string
		}{
			// sticky before
			{"!\na", "!\na"},
			{"!a\nb", "!a\nb"},
			{"!a\n\nb", "!a\nb"},
			{"a\n!", "a\n\n!"},
			{"a\n!b", "a\n\n!b"},
			{"a\n\n!b", "a\n\n!b"},

			// sticky after
			{"a\n+", "a\n+"},
			{"a\n+b", "a\n+ b"},
			{"a\n\n+b", "a\n+ b"},
			{"+\na", "+\n\na"},
			{"+a\nb", "+ a\n\nb"},
			{"+a\n\nb", "+ a\n\nb"},

			{"!a\n>\nb", "!a\n>\n\nb"},
		}

		elements := config.Elements{
			"A": {
				Type:      node.TypeVerbatimWalled.String(),
				Delimiter: "!",
			},
			"B": {
				Type:      node.TypeWalled.String(),
				Delimiter: "+",
			},
			"C": {
				Type:      node.TypeWalled.String(),
				Delimiter: ">",
			},
		}
		transformers := []transformer.Transformer{
			sticky.Transformer{sticky.Map{
				"SA": sticky.Sticky{
					Element: "A",
				},
				"SB": sticky.Sticky{
					Element: "B",
					After:   true,
				},
			}},
		}
		for _, c := range cases {
			name := fmt.Sprintf("%q", c.in)
			t.Run(name, func(t *testing.T) {
				test(t, elements, transformers, c.in, c.out, 0)
			})
		}
	})

	t.Run("inline sticky", func(t *testing.T) {
		// note: inline sticky use the exact same transformer as normal
		// sticky which doesn't differentiate between blocks and inlines
		cases := []struct {
			in  string
			out string
		}{
			{"(())****", "(())****"},
			{"((a))****", "((a))****"},
			{"(())**a**", "(())**a**"},
			{"((a))**b**", "((a))**b**"},
			{"((a)) **b**", "((a)) **b**"},
			{"((a))b**c**", "((a)) b **c**"},
			{"((a))\n**b**", "((a))\n**b**"},

			{"a\n((b))**c**", "a\n((b))**c**"},
			{"((a))**b**((c))**d**", "((a))**b** ((c))**d**"},
		}

		elements := config.Elements{
			"A": {
				Type:      node.TypeUniform.String(),
				Delimiter: "(",
			},
			"B": {
				Type:      node.TypeUniform.String(),
				Delimiter: "*",
			},
		}
		transformers := []transformer.Transformer{
			sticky.Transformer{sticky.Map{
				"SA": sticky.Sticky{
					Element: "A",
					Target:  "B",
				},
			}},
		}
		for _, c := range cases {
			name := fmt.Sprintf("%q", c.in)
			t.Run(name, func(t *testing.T) {
				test(t, elements, transformers, c.in, c.out, 0)
			})
		}
	})
}

func TestUniform(t *testing.T) {
	cases := []struct {
		in  string
		out string
	}{
		{"**", "****"},
		{"** ", "****"},
		{"**a", "**a**"},
		{"**a**b", "**a** b"},
		{"**\n", "****"},
		{"**\n ", "****"},
		{"**\na", "**\na**"},
		{"**\na**", "**\na**"},
		{"**\na**b", "**\na** b"},
		{"**\n**", "**\n**"},

		{"a**", "a ****"},

		// nested
		{"**__", "**____**"},
		{"**a__b", "**a __b__**"},
		{"**a__b__", "**a __b__**"},
		{"**a__b**", "**a __b__**"},
		{"**a__b**__", "**a __b__** ____"},
		{"**a__b__**", "**a __b__**"},

		// left-right delimiter
		{"((a", "((a))"},
	}

	elements := config.Elements{
		"MA": {
			Type:      node.TypeUniform.String(),
			Delimiter: "*",
		},
		"MB": {
			Type:      node.TypeUniform.String(),
			Delimiter: "_",
		},
		"MC": {
			Type:      node.TypeUniform.String(),
			Delimiter: "(",
		},
	}
	for _, c := range cases {
		name := fmt.Sprintf("%q", c.in)
		t.Run(name, func(t *testing.T) {
			test(t, elements, nil, c.in, c.out, 0)
		})
	}
}

func TestEscaped(t *testing.T) {
	cases := []struct {
		in  string
		out string
	}{
		{"``", "````"},
		{"`` ", "````"},
		{"``a", "``a``"},
		{"``a``b", "``a`` b"},
		{"``\n", "````"},
		{"``\n ", "````"},
		{"``\na", "``\na``"},
		{"``\na``", "``\na``"},
		{"``\na``b", "``\na`` b"},
		{"``\n``", "````"},
		{"`````", "```` `"},

		{"a``", "a ````"},

		// would be nested
		{"``__", "``__``"},
		{"``a__b", "``a__b``"},
		{"``a__b__", "``a__b__``"},
		{"``a__b``", "``a__b``"},
		{"``a__b``__", "``a__b`` ____"},
		{"``a__b__``", "``a__b__``"},

		// escape
		{"```", "``\\`\\``"},
		{"``\\`", "``\\`\\``"},
		{"``\\``", "``\\``\\``"},
		{"``\\a``b", "``\\a``b\\``"},
		{"``\\a``b", "``\\a``b\\``"},

		// left-right delimiter
		{"[[a", "[[a]]"},

		{"a\n``b``", "a\n``b``"},
	}

	elements := config.Elements{
		"MA": {
			Type:      node.TypeEscaped.String(),
			Delimiter: "`",
		},
		"MB": {
			Type:      node.TypeUniform.String(),
			Delimiter: "_",
		},
		"MC": {
			Type:      node.TypeEscaped.String(),
			Delimiter: "[",
		},
	}
	for _, c := range cases {
		name := fmt.Sprintf("%q", c.in)
		t.Run(name, func(t *testing.T) {
			test(t, elements, nil, c.in, c.out, 0)
		})
	}
}

func TestPrefixed(t *testing.T) {
	cases := []struct {
		in  string
		out string
	}{
		{`\`, `\`},
		{`\a`, `\ a`},
		{`a\`, `a \`},
	}

	elements := config.Elements{
		"MA": {
			Type:      node.TypePrefixed.String(),
			Delimiter: `\`,
		},
	}
	for _, c := range cases {
		name := fmt.Sprintf("%q", c.in)
		t.Run(name, func(t *testing.T) {
			test(t, elements, nil, c.in, c.out, 0)
		})
	}

	t.Run("with content", func(t *testing.T) {
		cases := []struct {
			in  string
			out string
		}{
			{"a:", "a:"},
			{"a:b", "a:b"},
			{"ba:", "b a:"},
		}

		elements := config.Elements{
			"MA": {
				Type:      node.TypePrefixed.String(),
				Delimiter: "a:",
				Matcher:   "url",
			},
		}
		for _, c := range cases {
			name := fmt.Sprintf("%q", c.in)
			t.Run(name, func(t *testing.T) {
				test(t, elements, nil, c.in, c.out, 0)
			})
		}
	})
}

func TestEscape(t *testing.T) {
	cases := []struct {
		in  string
		out string
	}{
		{`\`, `\`},      // \
		{`\\`, `\`},     // \
		{`\\\`, `\\\`},  // \\
		{`\\\\`, `\\\`}, // \\

		{"a", "a"},        // a
		{`\a`, `\a`},      // \a
		{`\\a`, `\a`},     // \a
		{`\\\a`, `\\\a`},  // \\a
		{`\\\\a`, `\\\a`}, // \\a

		// punctuation
		{"!", "!"},         // !
		{`\!`, `!`},        // !
		{`\\!`, `\\!`},     // \!
		{`\\\!`, `\\!`},    // \!
		{`\\\\!`, `\\\\!`}, // \\!

		{`a\`, `a\`},      // a\
		{`a\\`, `a\`},     // a\
		{`a\\\`, `a\\\`},  // a\\
		{`a\\\\`, `a\\\`}, // a\\

		{"*a", "* a"},      // B(a)
		{`\*`, `\*`},       // *
		{`\\*`, `\\*`},     // \*
		{`\\\*`, `\\*`},    // \*
		{`\\\\*`, `\\\\*`}, // \\*

		// text block
		{"\\*\n\\*", "\\*\n\\*"}, // * *

		{"**a", "**a**"},          // I(a)
		{`\**`, `\**`},            // **
		{`\\**`, `\\ ****`},       // \
		{`\\**a`, `\\ **a**`},     // \I(a)
		{`\\\**`, `\\\**`},        // \**
		{`\\\**a`, `\\\**a`},      // \**a
		{`\\\\**a`, `\\\\ **a**`}, // \\I(A)

		{`a\**`, `a\**`},          // a**
		{`a\\**`, `a\\ ****`},     // a\
		{`a\\\**`, `a\\\**`},      // a\**
		{`a\\\\**`, `a\\\\ ****`}, // a\\

		{`a\***`, `a\* ****`},     // a*
		{`a\***b`, `a\* **b**`},   // a*I(b)
		{`a\*\**`, `a\*\**`},      // a***
		{`a\*\*\*`, `a\*\**`},     // a***
		{`a\*\*\**`, `a\*\*\**`},  // a****
		{`a\*\*\*\*`, `a\*\*\**`}, // a****

		// prefixed, non-punctuation delimiter
		{"http://a", "http://a"},          // I(a)
		{`\http://`, `\http://`},          // http://
		{`\\http://`, `\\ http://`},       // \
		{`\\http://a`, `\\ http://a`},     // \I(a)
		{`\\\http://`, `\\\http://`},      // \http://
		{`\\\http://a`, `\\\http://a`},    // \http://a
		{`\\\\http://a`, `\\\\ http://a`}, // \\I(A)

		{`a\http://`, `a\http://`},        // ahttp://
		{`a\\http://`, `a\\ http://`},     // a\
		{`a\\\http://`, `a\\\http://`},    // a\http://
		{`a\\\\http://`, `a\\\\ http://`}, // a\\

		// closing delimiter
		{`**\`, `**\\**`},             // I(\)
		{`**\*`, `**\***`},            // I(*)
		{`**\**`, `**\*\***`},         // I(**)
		{`**\*\**`, `**\*\*\***`},     // I(***)
		{`**\*\*\**`, `**\*\*\*\***`}, // I(****)

		{`***a`, `***a**`},     // I(*a)
		{`***\*a`, `**\**a**`}, // I(**a)

		// left/right closing delimiter
		{`{{\`, `{{\\}}`},             // I(\)
		{`{{\}`, `{{\}}}`},            // I(})
		{`{{\}}`, `{{\}\}}}`},         // I(}})
		{`{{\}\}}`, `{{\}\}\}}}`},     // I(}}})
		{`{{\}\}\}}`, `{{\}\}\}\}}}`}, // I(}}}})

		{`{{**\`, `{{**\\**}}`},            // I1(I2(\))
		{`{{**\}`, `{{**}**}}`},            // I1(I2(}))
		{`{{**\}}`, `{{**\}}**}}`},         // I1(I2(}}))
		{`{{**\}\}}`, `{{**\}\}}**}}`},     // I1(I2(}}}))
		{`{{**\}\}\}}`, `{{**\}\}\}}**}}`}, // I1(I2(}}}}))

		{`{{**\\}}`, `{{**\\**}}`}, // I1(I2(\))

		// nested
		{">*a", "> * a"},      // B1(B2(a))
		{`>\*`, `> \*`},       // B(*)
		{`>\\*`, `> \\*`},     // B(\*)
		{`>\\\*`, `> \\*`},    // B(\*)
		{`>\\\\*`, `> \\\\*`}, // B(\\*)

		// nested closing delimiter
		{`>**\`, `> **\\**`},       // B(I(\))
		{`>{{**\`, `> {{**\\**}}`}, // B(I1(I2(BR)))

		// in verbatim
		{"`\n\\\\", "`\n\\\\\n`"}, // B(\n\\)
		{"``a\\\\", "``a\\\\``"},  // I(a\\)
	}

	elements := config.Elements{
		"A": {
			Type:      node.TypeHanging.String(),
			Delimiter: "*",
		},
		"B": {
			Type:      node.TypeWalled.String(),
			Delimiter: ">",
		},
		"C": {
			Type:      node.TypeFenced.String(),
			Delimiter: "`",
		},
		"MA": {
			Type:      node.TypeUniform.String(),
			Delimiter: "*",
		},
		"MB": {
			Type:      node.TypeEscaped.String(),
			Delimiter: "`",
		},
		// use "{" as it doesn't need escaping
		// in -run test regex as "(" or "["
		"MC": {
			Type:      node.TypeUniform.String(),
			Delimiter: "{",
		},
		"MD": {
			Type:      node.TypePrefixed.String(),
			Delimiter: "http://",
			Matcher:   "url",
		},
	}
	for _, c := range cases {
		name := fmt.Sprintf("%q", strings.ReplaceAll(c.in, "/", "2F")) // %2F is URL-escaped slash
		t.Run(name, func(t *testing.T) {
			test(t, elements, nil, c.in, c.out, 0)
		})
	}
}

// TODO: Might want to parametrize TestEscape* tests and define non-standard
// outs only when they are different. It would be easier to compare the
// differences.
func TestEscapeWithClash(t *testing.T) {
	// registered line break with delimiter "\" -> escape clash

	cases := []struct {
		in  string
		out string
	}{
		{`\`, `\`},       // BR
		{`\\`, `\\`},     // \
		{`\\\`, `\\ \`},  // \BR
		{`\\\\`, `\\\\`}, // \\

		{"a", "a"},         // a
		{`\a`, `\ a`},      // BRa
		{`\\a`, `\\a`},     // \a
		{`\\\a`, `\\ \ a`}, // \BRa
		{`\\\\a`, `\\\\a`}, // \\a

		// punctuation
		{"!", "!"},         // !
		{`\!`, `!`},        // !
		{`\\!`, `\\!`},     // \!
		{`\\\!`, `\\!`},    // \!
		{`\\\\!`, `\\\\!`}, // \\!

		{`a\`, `a \`},      // aBR
		{`a\\`, `a\\`},     // a\
		{`a\\\`, `a\\ \`},  // a\BR
		{`a\\\\`, `a\\\\`}, // a\\

		{"*a", "* a"},      // B(a)
		{`\*`, `\*`},       // *
		{`\\*`, `\\*`},     // \*
		{`\\\*`, `\\*`},    // \*
		{`\\\\*`, `\\\\*`}, // \\*

		// text block
		{"\\*\n\\*", "\\*\n\\*"}, // * *

		{"**a", "**a**"},          // I(a)
		{`\**`, `\**`},            // **
		{`\\**`, `\\ ****`},       // \
		{`\\**a`, `\\ **a**`},     // \I(a)
		{`\\\**`, `\\\**`},        // \**
		{`\\\**a`, `\\\**a`},      // \**a
		{`\\\\**a`, `\\\\ **a**`}, // \\I(A)

		{`a\**`, `a\**`},          // a**
		{`a\\**`, `a\\ ****`},     // a\
		{`a\\\**`, `a\\\**`},      // a\**
		{`a\\\\**`, `a\\\\ ****`}, // a\\

		{`a\***`, `a\* ****`},     // a*
		{`a\***b`, `a\* **b**`},   // a*I(b)
		{`a\*\**`, `a\*\**`},      // a***
		{`a\*\*\*`, `a\*\**`},     // a***
		{`a\*\*\**`, `a\*\*\**`},  // a****
		{`a\*\*\*\*`, `a\*\*\**`}, // a****

		// prefixed, non-punctuation delimiter
		{"http://a", "http://a"},          // I(a)
		{`\http://`, `\http://`},          // http://
		{`\\http://`, `\\ http://`},       // \
		{`\\http://a`, `\\ http://a`},     // \I(a)
		{`\\\http://`, `\\\http://`},      // \http://
		{`\\\http://a`, `\\\http://a`},    // \http://a
		{`\\\\http://a`, `\\\\ http://a`}, // \\I(A)

		{`a\http://`, `a\http://`},        // ahttp://
		{`a\\http://`, `a\\ http://`},     // a\
		{`a\\\http://`, `a\\\http://`},    // a\http://
		{`a\\\\http://`, `a\\\\ http://`}, // a\\

		// closing delimiter
		{`**\`, `**\ **`},             // I(BR)
		{`**\*`, `**\***`},            // I(*)
		{`**\**`, `**\*\***`},         // I(**)
		{`**\*\**`, `**\*\*\***`},     // I(***)
		{`**\*\*\**`, `**\*\*\*\***`}, // I(****)

		{`***a`, `***a**`},     // I(*a)
		{`***\*a`, `**\**a**`}, // I(**a)

		// left/right closing delimiter
		{`{{\`, `{{\ }}`},             // I(BR)
		{`{{\}`, `{{\}}}`},            // I(})
		{`{{\}}`, `{{\}\}}}`},         // I(}})
		{`{{\}\}}`, `{{\}\}\}}}`},     // I(}}})
		{`{{\}\}\}}`, `{{\}\}\}\}}}`}, // I(}}}})

		{`{{**\`, `{{**\ **}}`},            // I1(I2(BR))
		{`{{**\}`, `{{**}**}}`},            // I1(I2(}))
		{`{{**\}}`, `{{**\}}**}}`},         // I1(I2(}}))
		{`{{**\}\}}`, `{{**\}\}}**}}`},     // I1(I2(}}}))
		{`{{**\}\}\}}`, `{{**\}\}\}}**}}`}, // I1(I2(}}}}))

		{`{{**\\}}`, `{{**\\**}}`}, // I1(I2(\))

		// nested
		{">*a", "> * a"},      // B1(B2(a))
		{`>\*`, `> \*`},       // B(*)
		{`>\\*`, `> \\*`},     // B(\*)
		{`>\\\*`, `> \\*`},    // B(\*)
		{`>\\\\*`, `> \\\\*`}, // B(\\*)

		// nested closing delimiter
		{`>{{\`, `> {{\ }}`},       // B(I(BR))
		{`>{{**\`, `> {{**\ **}}`}, // B(I1(I2(BR)))

		// in verbatim
		{"`\n\\\\", "`\n\\\\\n`"}, // B(\n\\)
		{"``a\\\\", "``a\\\\``"},  // I(a\\)
	}

	elements := config.Elements{
		"A": {
			Type:      node.TypeHanging.String(),
			Delimiter: "*",
		},
		"B": {
			Type:      node.TypeWalled.String(),
			Delimiter: ">",
		},
		"C": {
			Type:      node.TypeFenced.String(),
			Delimiter: "`",
		},
		"MA": {
			Type:      node.TypeUniform.String(),
			Delimiter: "*",
		},
		"MB": {
			Type:      node.TypeEscaped.String(),
			Delimiter: "`",
		},
		"MC": {
			Type:      node.TypePrefixed.String(),
			Delimiter: `\`,
		},
		// use "{" as it doesn't need escaping
		// in -run test regex as "(" or "["
		"MD": {
			Type:      node.TypeUniform.String(),
			Delimiter: "{",
		},
		"ME": {
			Type:      node.TypePrefixed.String(),
			Delimiter: "http://",
			Matcher:   "url",
		},
	}
	for _, c := range cases {
		name := fmt.Sprintf("%q", strings.ReplaceAll(c.in, "/", "2F")) // %2F is URL-escaped slash
		t.Run(name, func(t *testing.T) {
			test(t, elements, nil, c.in, c.out, 0)
		})
	}
}

func TestLineLength(t *testing.T) {
	cases := []struct {
		in  string
		out string
	}{
		{"abcdefgh", "abcdefgh"},
		{"abcdefgha", "abcdefgha"},
		{"*abcdefgh", "* abcdefgh"},
		{"*\n abcdefgh", "* abcdefgh"},
		{">abcdefgh", "> abcdefgh"},

		// newline in the valid line length range
		{"abcd\nefgh", "abcd\nefgh"},

		// spacing
		{"a  b", "a b"},

		// tab
		{"abcd efg", "abcd efg"},
		{"abcd\tefg", "abcd efg"},
		{"abcd efgh", "abcd\nefgh"},
		{"abcd\tefgh", "abcd\nefgh"},

		// multiple words
		{"abcdef g", "abcdef g"},
		{"abcdef gh", "abcdef\ngh"},
		{"abcd e f", "abcd e f"},
		{"abcd e fg", "abcd e\nfg"},
		{"ab c d e", "ab c d e"},
		{"abc d e f", "abc d e\nf"},
		{"a b c d e", "a b c d\ne"},

		{"abcdefgh a", "abcdefgh\na"},
		{"abcdefgha b", "abcdefgha\nb"},
		{"abcdefgh\nabcdefgha b", "abcdefgh\nabcdefgha\nb"},
		{"abcd efgh abcd\nabcd efgh abcd", "abcd\nefgh\nabcd\nabcd\nefgh\nabcd"},
		{"abcdef\na", "abcdef a"},
		{"abcdef\na\nb", "abcdef a\nb"},

		{"*abcd e", "* abcd e"},
		{"*abcde f", "* abcde\n  f"},
		{"*\n abcde f", "* abcde\n  f"},
		{">abcde f", "> abcde\n> f"},

		// verbatim - shouldn't wrap
		{".abcdefgh", ".abcdefgh"},
		{"!abcdefgh", "!abcdefgh"},
		{"`abcdefgh", "`abcdefgh\n`"},
		{"`\nabcdefgh", "`\nabcdefgh\n`"},

		// nested verbatim
		{">.abcdefgh", "> .abcdefgh"},
		{">!abcdefgh", "> !abcdefgh"},
		{">`abcdefgh", "> `abcdefgh\n> `"},
		{">`\n>abcdefgh", "> `\n> abcdefgh\n> `"},

		// with inlines
		{"____a", "____ a"},
		{"__ab c", "__ab c__"},
		{"__abc d", "__abc d__"},
		{"a__", "a ____"},
		{"a __b__ c", "a __b__\nc"},
		{"a__bc", "a __bc__"},
		{"a__bcd", "a __bcd\n__"},
		{"a__bc d", "a __bc d\n__"},
		{"a__bcd e", "a __bcd\ne__"},
		{"abc__", "abc ____"},
		{"abc__d", "abc __d\n__"},
		{"abcd__", "abcd __\n__"},
		{"abcd__e", "abcd __e\n__"},
		{"abcde__f", "abcde __\nf__"},
		{"abcdef__", "abcdef\n____"},

		// escaped
		{"````a", "```` a"},
		{"``ab c", "``ab c``"},
		{"``abc d", "``abc d``"},
		{"a``", "a ````"},
		{"a ``b`` c", "a ``b``\nc"},
		{"a``bc", "a ``bc``"},
		{"a``bcd", "a\n``bcd``"},
		{"a``bc d", "a\n``bc d``"},
		{"a``bcd e", "a\n``bcd e``"},
		{"abc``", "abc ````"},
		{"abc``d", "abc\n``d``"},
		{"abcd``", "abcd\n````"},

		// prefixed
		{"@a b", "@a b"},
		{"@abcde f", "@abcde f"},
		{"@abcdef g", "@abcdef\ng"},
		{"a@b", "a @b"},
		{"abcde @f", "abcde @f"},
		{"abcdef @g", "abcdef\n@g"},
		{"ab @cd ef", "ab @cd\nef"},

		// inline sticky
		{"[[a]]__b__c", "[[a]]__b__\nc"},
		{"a[[b]]__c__", "a [[b]]__\nc__"},

		{"[[a]]__b__,", "[[a]]__b__,"},
		{"[[a]]__b__, a", "[[a]]__b__,\na"},

		// inline sticky with escaped
		{"__a__``b``c", "__a__``b``\nc"},
		{"a____````", "a\n____````"},
		{"a__b__``c``", "a\n__b__``c``"},
		{"ab__c__``d``", "ab\n__c__``d``"},
		{"abc__d__``e``", "abc __d\n__``e``"},
		{"abcd__e__``f``", "abcd __e\n__``f``"},

		{"``a``__b__c", "``a``__b__\nc"},
		{"a````____", "a ````__\n__"},
		{"a``b``__c__", "a\n``b``__c\n__"},
		{"ab``c``__d__", "ab\n``c``__d\n__"},
		{"abc``d``__e__", "abc\n``d``__e\n__"},
		{"abcd``e``__f__", "abcd\n``e``__f\n__"},

		// escape
		{"abcdefgh >a", "abcdefgh\n\\>a"},
		{"abcdefgh >a b", "abcdefgh\n\\>a b"},
		{">abcdef >a", "> abcdef\n> \\>a"},
		{">abcdef >a b", "> abcdef\n> \\>a b"},

		// no space/newline before puncutation
		{"____,", "____,"},
		{",____", ", ____"},
		{"____…", "____…"}, // UTF-8
		{"… ____", "… ____"},

		// UTF-8
		{"→ abcdef", "→ abcdef"},
		{"→→ abcdef", "→→\nabcdef"},
		{"abcdef →", "abcdef →"},
		{"abcdef →→", "abcdef\n→→"},
		// uniform
		{"★★★★abc", "★★★★ abc"}, // want bytes=16, utf-8=8
		{"abc★★", "abc ★★★★"},
		{"ab→★★", "ab→ ★★★★"},
		{"abcd★★", "abcd ★★\n★★"},
		{"abcde★★", "abcde ★★\n★★"},
		{"abcdef★★", "abcdef\n★★★★"},
		// escaped
		{"∑∑∑∑abc", "∑∑∑∑ abc"}, // want bytes=16, utf-8=8
		{"ab→∑∑", "ab→ ∑∑∑∑"},
		{"abc∑∑", "abc ∑∑∑∑"},
		{"abcd∑∑", "abcd\n∑∑∑∑"},
		{"abcde∑∑", "abcde\n∑∑∑∑"},
		{"abcdef∑∑", "abcdef\n∑∑∑∑"},
		{"ab∑∑c", "ab ∑∑c∑∑"},
		{"ab∑∑→", "ab ∑∑→∑∑"},
		// prefixed
		{"¡a abcde", "¡a abcde"}, // want bytes=9, utf-8=8
		{"abcde ¡a", "abcde ¡a"},
		{"abcd→ ¡a", "abcd→ ¡a"},
		{"abcde¡→", "abcde ¡→"},
	}

	elements := config.Elements{
		"A": {
			Type:      node.TypeHanging.String(),
			Delimiter: "*",
		},
		"B": {
			Type:      node.TypeWalled.String(),
			Delimiter: ">",
		},
		"C": {
			Type:      node.TypeVerbatimLine.String(),
			Delimiter: ".",
		},
		"D": {
			Type:      node.TypeVerbatimWalled.String(),
			Delimiter: "!",
		},
		"E": {
			Type:      node.TypeFenced.String(),
			Delimiter: "`",
		},
		"MA": {
			Type:      node.TypeUniform.String(),
			Delimiter: "_",
		},
		"MB": {
			Type:      node.TypeUniform.String(),
			Delimiter: "[",
		},
		"MC": {
			Type:      node.TypeEscaped.String(),
			Delimiter: "`",
		},
		"MD": {
			Type:      node.TypePrefixed.String(),
			Delimiter: "@",
			Matcher:   "url",
		},
		"ME": {
			Type:      node.TypeUniform.String(),
			Delimiter: "★",
		},
		"MF": {
			Type:      node.TypeEscaped.String(),
			Delimiter: "∑",
		},
		"MG": {
			Type:      node.TypePrefixed.String(),
			Delimiter: "¡",
			Matcher:   "url",
		},
	}
	transformers := []transformer.Transformer{
		sticky.Transformer{sticky.Map{
			"SA": sticky.Sticky{
				Element: "MB",
				Target:  "MA",
			},
			"SB": sticky.Sticky{
				Element: "MA",
				Target:  "MC",
			},
			"SC": sticky.Sticky{
				Element: "MC",
				Target:  "MA",
			},
		}},
	}
	for _, c := range cases {
		name := fmt.Sprintf("%q", c.in)
		t.Run(name, func(t *testing.T) {
			test(t, elements, transformers, c.in, c.out, 8)
		})
	}
}

func test(t *testing.T, elements config.Elements, transformers []transformer.Transformer, in, out string, lineLength int) {
	t.Helper()

	if elements == nil {
		elements = config.Elements{}
	}

	printed := runPrint(t, elements, transformers, in, *printTree, lineLength)
	if printed != out {
		t.Errorf("got %q, want %q", printed, out)
	}

	if !*noReprint {
		previousPrint := printed
		for i := 0; ; i++ {
			if i > 2 {
				t.Errorf("too many reprints, skipping")
				break
			}

			reprinted := runPrint(t, elements, transformers, previousPrint, *printTree, lineLength)
			if reprinted == previousPrint {
				break
			}

			// test that printing the output returns the same output, if it
			// doesn't it is not canonical
			t.Errorf("reprint %d got %q, want %q", i+1, reprinted, previousPrint)
			previousPrint = reprinted
		}

		hasLeaf, hasText := hasLeafOrText(elements)
		if !hasLeaf || !hasText {
			// undefined leaf or text element:
			// test that result are the same whether the leaf or text
			// elements are defined or not
			if !hasLeaf {
				if _, ok := elements["_T"]; ok {
					t.Fatal("element _T already exists")
				}
				elements["_T"] = config.Element{
					Type: node.TypeLeaf.String(),
				}
			}
			if !hasText {
				if _, ok := elements["_MT"]; ok {
					t.Fatal("element _MT already exists")
				}
				elements["_MT"] = config.Element{
					Type: node.TypeText.String(),
				}
			}
			printedDefined := runPrint(t, elements, transformers, in, false, lineLength)
			if printedDefined != printed {
				t.Errorf("with defined text got %q, with undefined %q", printedDefined, printed)
			}
		}
	}
}

func runPrint(t *testing.T, elements config.Elements, transformers []transformer.Transformer, in string, printTree bool, lineLength int) string {
	t.Helper()

	p := parser.Parser{
		Elements: elements.ParserElements(),
		Matchers: matcher.Defaults(),
		TabWidth: 8,
	}
	root, err := p.Parse(nil, []byte(in))
	if err != nil {
		t.Fatal(err)
	}
	root = transformer.Group(transformers).Transform(root)

	if printTree {
		var b strings.Builder
		if err := (node.Printer{node.PrintData}).Fprint(&b, root); err != nil {
			t.Fatal(err)
		}
		fmt.Println(b.String())
	}

	var b strings.Builder
	if err := (printer.Printer{Elements: elements.ParserElements(), LineLength: lineLength}).Fprint(&b, root); err != nil {
		t.Fatal(err)
	}
	return b.String()
}

func hasLeafOrText(elements config.Elements) (bool, bool) {
	var hasLeaf, hasText bool
	for _, e := range elements {
		if hasLeaf && hasText {
			break
		}
		if e.Type == node.TypeLeaf.String() {
			hasLeaf = true
		} else if e.Type == node.TypeText.String() {
			hasText = true
		}
	}
	return hasLeaf, hasText
}
