// package printer_test contines tests the printer package.
//
// out cases containing "$$" denote the wanted output for the printer mode
// printer.KeepNewlines; the output is placed after the "$$". If there is no
// "$$" section, the wanted output is assumed to be the regular out string.
package printer_test

import (
	"flag"
	"fmt"
	"strings"
	"testing"

	"github.com/touchmarine/to/config"
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
		{"a\nb", "a\nb$$a\nb"},
		{"a\n b", "a\nb$$a\nb"},
		{"a\n\n b", "a\n\nb"},
		{"ab\n c", "ab\nc$$ab\nc"},
		{"ab\n\n c", "ab\n\nc"},

		// interrupted by empty blocks
		{"a\n>\n*\nb", "a\nb"},
		{"a\n>b\n*\nc", "a\n\n> b\n\nc"},
		{"a\n>\n*b\nc", "a\n\n* b\n\nc"},
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
	for _, c := range cases {
		name := fmt.Sprintf("%q", c.in)
		t.Run(name, func(t *testing.T) {
			test(t, elements, nil, c.in, c.out, 0)
		})
		t.Run("KeepNewlines-"+name, func(t *testing.T) {
			test(t, elements, nil, c.in, c.out, printer.KeepNewlines)
		})
	}
}

func TestVerbatimLine(t *testing.T) {
	cases := []struct {
		in  string
		out string
	}{
		{".a", ""},
		{".A", ".A"},
		{".aa", ".a a"},
		{".aa ", ".a a"},
		// would be nested-but can only contain verbatim
		{".a>", ".a >"},
		{".a>b", ".a >b"},

		// nested
		{">.a", ""},
		{">.ab", "> .a b"},
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
		t.Run("KeepNewlines-"+name, func(t *testing.T) {
			test(t, elements, nil, c.in, c.out, printer.KeepNewlines)
		})
	}
}

func TestHanging(t *testing.T) {
	cases := []struct {
		in  string
		out string
	}{
		{"-", ""},
		{"-a", "- a"},
		{"-\n a", "- a"},
		{"-a\n b", "- a\n  b$$- a\n  b"},
		{"-a\n -b", "- a\n\n  - b"},
		{"-a\n\n -b", "- a\n\n  - b"},
		{"-a\n \n -b", "- a\n\n  - b"},
		{"-a\n\n\n -b", "- a\n\n  - b"},

		// nested
		{"->", ""},
		{"->a", "- > a"},
		{"-\n>", ""},
		{"-\n>a", "> a"},
		{"-\n >", ""},
		{"-\n >a", "- > a"},
		{">-", ""},
		{">-a", "> - a"},
		{">\n>-", ""},
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
		t.Run("KeepNewlines-"+name, func(t *testing.T) {
			test(t, elements, nil, c.in, c.out, printer.KeepNewlines)
		})
	}
}

func TestRankedHanging(t *testing.T) {
	cases := []struct {
		in  string
		out string
	}{

		{"==", ""},
		{"==a", "== a"},
		{"==\n  a", "== a"},
		{"==a\n  b", "== a\n   b$$== a\n   b"},
		{"==a", "== a"},
		{"==a\n  b", "== a\n   b$$== a\n   b"},
		{"==a\n\n  b", "== a\n\n   b"},
		{"==a\n \n  b", "== a\n\n   b"},
		{"==a\n\n\n  b", "== a\n\n   b"},

		// nested
		{"==>", ""},
		{"==>a", "== > a"},
		{"==\n>", ""},
		{"==\n>a", "> a"},
		{"==\n  >", ""},
		{"==\n  >a", "== > a"},
		{">==", ""},
		{">==a", "> == a"},
		{">\n>==", ""},
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
		t.Run("KeepNewlines-"+name, func(t *testing.T) {
			test(t, elements, nil, c.in, c.out, printer.KeepNewlines)
		})
	}
}

func TestWalled(t *testing.T) {
	cases := []struct {
		in  string
		out string
	}{
		{"+", ""},
		{"+a", "+ a"},
		{"+\n+a", "+ a"},
		{"+a\n+b", "+ a\n+ b$$+ a\n+ b"},
		{"+a\n++b", "+ a\n+\n+ + b"},
		{"+a\n+\n++b", "+ a\n+\n+ + b"},
		{"+a\n++\n++b", "+ a\n+\n+ + b"},
		{"+a\n+\n+\n++b", "+ a\n+\n+ + b"},

		// nested
		{"+>", ""},
		{"+>a", "+ > a"},
		{"+\n>", ""},
		{"+\n>a", "> a"},
		{"+\n+>", ""},
		{"+\n+>a", "+ > a"},
		{">+", ""},
		{">+a", "> + a"},
		{">\n>+", ""},
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
		t.Run("KeepNewlines-"+name, func(t *testing.T) {
			test(t, elements, nil, c.in, c.out, printer.KeepNewlines)
		})
	}
}

func TestVerbatimWalled(t *testing.T) {
	cases := []struct {
		in  string
		out string
	}{
		{"!", ""},
		{"!a", "! a"},
		{"!\n!a", "! a"},
		{"!a\n!b", "! a b$$! a\n! b"},
		{"!a\n!\n!b", "! a b$$! a\n! b"},
		{"!a\n!\n!\n!b", "! a b$$! a\n! b"},
		// would be nested-but can only contain verbatim
		{"!>", "! >"},
		{"!>a", "! >a"},
		{"!\n>", ""},
		{"!\n>a", "> a"},
		{"!\n!>", "! >"},
		{"!\n!>a", "! >a"},

		// nested
		{">!", ""},
		{">!a", "> ! a"},
		{">\n>!", ""},
		{">\n>!a", "> ! a"},
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
		t.Run("KeepNewlines-"+name, func(t *testing.T) {
			test(t, elements, nil, c.in, c.out, printer.KeepNewlines)
		})
	}
}

func TestFenced(t *testing.T) {
	cases := []struct {
		in  string
		out string
	}{
		{"`", ""},
		{"`a", ""},
		{"`a`", ""},
		{"`a\n", ""},
		{"`a\nb", "`a\nb\n`"},
		{"`a\n\nb", "`a\n\nb\n`"},
		{"`a\n \nb", "`a\n \nb\n`"},
		{"`a\n\n\nb", "`a\n\n\nb\n`"},

		// trailing text
		{"`a`b", ""},
		{"`a\nb\n`", "`a\nb\n`"},
		{"`a\nb\n`c", "`a\nb\n`"},
		{"`a\nb\n`\nc", "`a\nb\n`\n\nc"},

		// escape
		{"``", ""},
		{"``\nb", "``\nb\n`"},
		{"`a\n`", ""},
		{"`a\nb`", "`a\nb`\n`"},

		// unnecessary escape
		{"`\\a\nb", "`a\nb\n`"},
		{"`a\nb\n\\`", "`a\nb\n\\`\n`"},

		// nested
		{">`", ""},
		{">\n>`", ""},
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
		t.Run("KeepNewlines-"+name, func(t *testing.T) {
			test(t, elements, nil, c.in, c.out, printer.KeepNewlines)
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
			{"a\n>\n*\nb", "a\nb"},
			{"a\n>b\n*\nc", "a\n\n> b\n\nc"},
			{"a\n>\n*b\nc", "a\n\n* b\n\nc"},
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
				node.TypeLeaf: "PA",
			}},
		}
		for _, c := range cases {
			name := fmt.Sprintf("%q", c.in)
			t.Run(name, func(t *testing.T) {
				test(t, elements, transformers, c.in, c.out, 0)
			})
			t.Run("eepNewlines-"+name, func(t *testing.T) {
				test(t, elements, transformers, c.in, c.out, printer.KeepNewlines)
			})
		}
	})

	t.Run("list", func(t *testing.T) {
		cases := []struct {
			in  string
			out string
		}{
			{"-a\n-", "- a"},
			{"-a\n-b", "- a\n- b"},
			{"-a\n\n-b", "- a\n- b"},

			// nested
			{"-a\n-", "- a"},
			{"-a\n-b", "- a\n- b"},

			// interrupted by empty blocks
			{"-a\n>\n-b", "- a\n- b"},
			{"-a\n>\n\n>\n-b", "- a\n- b"},
			{"-a\n>b\n\n>\n-c", "- a\n\n> b\n\n- c"},
			{"-a\n>\n\n>b\n-c", "- a\n\n> b\n\n- c"},
			{"-a\n>\n*\n-b", "- a\n- b"},
			{"-a\n>b\n*\n-c", "- a\n\n> b\n\n- c"},
			{"-a\n>\n*b\n-c", "- a\n\n* b\n\n- c"},
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
				"A": "LA",
			}},
		}
		for _, c := range cases {
			name := fmt.Sprintf("%q", c.in)
			t.Run(name, func(t *testing.T) {
				test(t, elements, transformers, c.in, c.out, 0)
			})
			t.Run("KeepNewlines-"+name, func(t *testing.T) {
				test(t, elements, transformers, c.in, c.out, printer.KeepNewlines)
			})
		}
	})

	t.Run("sticky", func(t *testing.T) {
		cases := []struct {
			in  string
			out string
		}{
			// sticky before
			{"!\na", "a"},
			{"!a\nb", "! a\nb"},
			{"!a\n\nb", "! a\nb"},
			{"a\n!", "a"},
			{"a\n!b", "a\n\n! b"},
			{"a\n\n!b", "a\n\n! b"},

			// sticky after
			{"a\n+", "a"},
			{"a\n+b", "a\n+ b"},
			{"a\n\n+b", "a\n+ b"},
			{"+\na", "a"},
			{"+a\nb", "+ a\n\nb"},
			{"+a\n\nb", "+ a\n\nb"},

			// interrupted by empty blocks
			// note: semantics change-a side effect of consistently
			// removing empty blocks
			{"!a\n>\nb", "! a\n\nb"},
			//{"!a\n>\nb", "! a\n>\n\nb"},
			//{"!a\n>\nb", "! a\nb"},
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
				"A": sticky.Sticky{
					Name: "SA",
				},
				"B": sticky.Sticky{
					Name:  "SB",
					After: true,
				},
			}},
		}
		for _, c := range cases {
			name := fmt.Sprintf("%q", c.in)
			t.Run(name, func(t *testing.T) {
				test(t, elements, transformers, c.in, c.out, 0)
			})
			t.Run("KeepNewlines-"+name, func(t *testing.T) {
				test(t, elements, transformers, c.in, c.out, printer.KeepNewlines)
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
			{"(())****", ""},
			{"((a))****", "((a))"},
			{"(())**a**", "**a**"},
			{"((a))**b**", "((a))**b**"},
			{"((a)) **b**", "((a))**b**"},
			{"((a))b**c**", "((a))b**c**"},
			{"((a))\n**b**", "((a))**b**"},

			{"a\n((b))**c**", "a\n((b))**c**$$a\n((b))**c**"},
			{"((a))**b**((c))**d**", "((a))**b**((c))**d**"},
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
				"A": sticky.Sticky{
					Name:   "SA",
					Target: "B",
				},
			}},
		}
		for _, c := range cases {
			name := fmt.Sprintf("%q", c.in)
			t.Run(name, func(t *testing.T) {
				test(t, elements, transformers, c.in, c.out, 0)
			})
			t.Run("KeepNewlines", func(t *testing.T) {
				t.Run(name, func(t *testing.T) {
					test(t, elements, transformers, c.in, c.out, printer.KeepNewlines)
				})
			})
		}
	})
}

func TestUniform(t *testing.T) {
	cases := []struct {
		in  string
		out string
	}{
		{"**", ""},
		{"** ", ""},
		{"**a", "**a**"},
		{"**a**b", "**a**b"},
		{"**\n", ""},
		{"**\n ", ""},
		{"**\na", "**\na**$$**\na**"},
		{"**\na**", "**\na**$$**\na**"},
		{"**\na**b", "**\na**b$$**\na**b"},
		{"**\n**", ""},

		{"a**", "a"},

		// nested
		{"**__", ""},
		{"**a__b", "**a__b__**"},
		{"**a__b__", "**a__b__**"},
		{"**a__b**", "**a__b__**"},
		{"**a__b**__", "**a__b__**"},
		{"**a__b__**", "**a__b__**"},

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
		t.Run("KeepNewlines-"+name, func(t *testing.T) {
			test(t, elements, nil, c.in, c.out, printer.KeepNewlines)
		})
	}
}

func TestEscaped(t *testing.T) {
	cases := []struct {
		in  string
		out string
	}{
		{"``", ""},
		{"`` ", ""},
		{"``a", "``a``"},
		{"``a``b", "``a``b"},
		{"``\n", ""},
		{"``\n ", ""},
		{"``\na", "``\na``"},
		{"``\na``", "``\na``"},
		{"``\na``b", "``\na``b"},
		{"``\n``", ""},
		{"`````", "`"},

		{"a``", "a"},

		// would be nested
		{"``__", "``__``"},
		{"``a__b", "``a__b``"},
		{"``a__b__", "``a__b__``"},
		{"``a__b``", "``a__b``"},
		{"``a__b``__", "``a__b``"},
		{"``a__b__``", "``a__b__``"},

		// escape
		{"```", "``\\`\\``"},
		{"``\\`", "``\\`\\``"},
		{"``\\``", "``\\``\\``"},
		{"``\\a``b", "``\\a``b\\``"},
		{"``\\a``b", "``\\a``b\\``"},

		// left-right delimiter
		{"[[a", "[[a]]"},

		{"a\n``b``", "a\n``b``$$a\n``b``"},
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
		t.Run("KeepNewlines-"+name, func(t *testing.T) {
			test(t, elements, nil, c.in, c.out, printer.KeepNewlines)
		})
	}
}

func TestPrefixed(t *testing.T) {
	cases := []struct {
		in  string
		out string
	}{
		{`\`, ""},
		{`\a`, "a"},
		{`a\`, "a"},
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
		t.Run("KeepNewlines-"+name, func(t *testing.T) {
			test(t, elements, nil, c.in, c.out, printer.KeepNewlines)
		})
	}

	t.Run("do not remove", func(t *testing.T) {
		cases := []struct {
			in  string
			out string
		}{
			{`\`, `\`},
			{`\a`, `\a`},
			{`a\`, `a\`},
		}

		elements := config.Elements{
			"MA": {
				Type:        node.TypePrefixed.String(),
				Delimiter:   `\`,
				DoNotRemove: true,
			},
		}
		for _, c := range cases {
			name := fmt.Sprintf("%q", c.in)
			t.Run(name, func(t *testing.T) {
				test(t, elements, nil, c.in, c.out, 0)
			})
			t.Run("KeepNewlines-"+name, func(t *testing.T) {
				test(t, elements, nil, c.in, c.out, printer.KeepNewlines)
			})
		}
	})

	t.Run("with content", func(t *testing.T) {
		cases := []struct {
			in  string
			out string
		}{
			{"a:", ""},
			{"a:b", "a:b"},
			{"ba:", "b"},
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
			t.Run("KeepNewlines-"+name, func(t *testing.T) {
				test(t, elements, nil, c.in, c.out, printer.KeepNewlines)
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
		{"\\*\n\\*", "\\*\n\\*$$\\*\n\\*"}, // * *

		{"**a", "**a**"},         // I(a)
		{`\**`, `\**`},           // **
		{`\\**`, `\`},            // \
		{`\\**a`, `\\**a**`},     // \I(a)
		{`\\\**`, `\\\**`},       // \**
		{`\\\**a`, `\\\**a`},     // \**a
		{`\\\\**a`, `\\\\**a**`}, // \\I(A)

		{`a\**`, `a\**`},     // a**
		{`a\\**`, `a\`},      // a\
		{`a\\\**`, `a\\\**`}, // a\**
		{`a\\\\**`, `a\\\`},  // a\\

		{`a\***`, `a*`},           // a*
		{`a\***b`, `a\***b**`},    // a*I(b)
		{`a\*\**`, `a\*\**`},      // a***
		{`a\*\*\*`, `a\*\**`},     // a***
		{`a\*\*\**`, `a\*\*\**`},  // a****
		{`a\*\*\*\*`, `a\*\*\**`}, // a****

		// prefixed, non-punctuation delimiter
		{"http://a", "http://a"},         // I(a)
		{`\http://`, `\http://`},         // http://
		{`\\http://`, `\`},               // \
		{`\\http://a`, `\\http://a`},     // \I(a)
		{`\\\http://`, `\\\http://`},     // \http://
		{`\\\http://a`, `\\\http://a`},   // \http://a
		{`\\\\http://a`, `\\\\http://a`}, // \\I(A)

		{`a\http://`, `a\http://`},     // ahttp://
		{`a\\http://`, `a\`},           // a\
		{`a\\\http://`, `a\\\http://`}, // a\http://
		{`a\\\\http://`, `a\\\`},       // a\\

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
		{`>**\`, `> **\\**`}, // B(I(\))

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
		t.Run("KeepNewlines-"+name, func(t *testing.T) {
			test(t, elements, nil, c.in, c.out, printer.KeepNewlines)
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
		{`\\\`, `\\\`},   // \BR
		{`\\\\`, `\\\\`}, // \\

		{"a", "a"},         // a
		{`\a`, `\a`},       // BRa
		{`\\a`, `\\a`},     // \a
		{`\\\a`, `\\\a`},   // \BRa
		{`\\\\a`, `\\\\a`}, // \\a

		// punctuation
		{"!", "!"},         // !
		{`\!`, `!`},        // !
		{`\\!`, `\\!`},     // \!
		{`\\\!`, `\\!`},    // \!
		{`\\\\!`, `\\\\!`}, // \\!

		{`a\`, `a\`},       // aBR
		{`a\\`, `a\\`},     // a\
		{`a\\\`, `a\\\`},   // a\BR
		{`a\\\\`, `a\\\\`}, // a\\

		{"*a", "* a"},      // B(a)
		{`\*`, `\*`},       // *
		{`\\*`, `\\*`},     // \*
		{`\\\*`, `\\*`},    // \*
		{`\\\\*`, `\\\\*`}, // \\*

		// text block
		{"\\*\n\\*", "\\*\n\\*$$\\*\n\\*"}, // * *

		{"**a", "**a**"},         // I(a)
		{`\**`, `\**`},           // **
		{`\\**`, `\\`},           // \
		{`\\**a`, `\\**a**`},     // \I(a)
		{`\\\**`, `\\\**`},       // \**
		{`\\\**a`, `\\\**a`},     // \**a
		{`\\\\**a`, `\\\\**a**`}, // \\I(A)

		{`a\**`, `a\**`},     // a**
		{`a\\**`, `a\\`},     // a\
		{`a\\\**`, `a\\\**`}, // a\**
		{`a\\\\**`, `a\\\\`}, // a\\

		{`a\***`, `a*`},           // a*
		{`a\***b`, `a\***b**`},    // a*I(b)
		{`a\*\**`, `a\*\**`},      // a***
		{`a\*\*\*`, `a\*\**`},     // a***
		{`a\*\*\**`, `a\*\*\**`},  // a****
		{`a\*\*\*\*`, `a\*\*\**`}, // a****

		// prefixed, non-punctuation delimiter
		{"http://a", "http://a"},         // I(a)
		{`\http://`, `\http://`},         // http://
		{`\\http://`, `\\`},              // \
		{`\\http://a`, `\\http://a`},     // \I(a)
		{`\\\http://`, `\\\http://`},     // \http://
		{`\\\http://a`, `\\\http://a`},   // \http://a
		{`\\\\http://a`, `\\\\http://a`}, // \\I(A)

		{`a\http://`, `a\http://`},     // ahttp://
		{`a\\http://`, `a\\`},          // a\
		{`a\\\http://`, `a\\\http://`}, // a\http://
		{`a\\\\http://`, `a\\\\`},      // a\\

		// closing delimiter
		{`**\`, "**\\\n**"},           // I(BR)
		{`**\*`, `**\***`},            // I(*)
		{`**\**`, `**\*\***`},         // I(**)
		{`**\*\**`, `**\*\*\***`},     // I(***)
		{`**\*\*\**`, `**\*\*\*\***`}, // I(****)

		{`***a`, `***a**`},     // I(*a)
		{`***\*a`, `**\**a**`}, // I(**a)

		// left/right closing delimiter
		{`{{\`, "{{\\\n}}"},           // I(BR)
		{`{{\}`, `{{\}}}`},            // I(})
		{`{{\}}`, `{{\}\}}}`},         // I(}})
		{`{{\}\}}`, `{{\}\}\}}}`},     // I(}}})
		{`{{\}\}\}}`, `{{\}\}\}\}}}`}, // I(}}}})

		{`{{**\`, "{{**\\\n**}}"},          // I1(I2(BR))
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
		{`>{{\`, "> {{\\\n> }}"},       // B(I(BR))
		{`>{{**\`, "> {{**\\\n> **}}"}, // B(I1(I2(BR)))

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
			Type:        node.TypePrefixed.String(),
			Delimiter:   `\`,
			DoNotRemove: true,
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
		t.Run("KeepNewlines-"+name, func(t *testing.T) {
			test(t, elements, nil, c.in, c.out, printer.KeepNewlines)
		})
	}
}

func TestDoNotRemove(t *testing.T) {
	cases := []struct {
		in  string
		out string
	}{
		{".a", ".a"},
		{".a ", ".a"},

		{">.a ", "> .a"},
		{">>.a ", "> > .a"},

		{".b", ""},
		{".b.a ", ".b .a"},

		{`\`, `\`},
		{`a\`, `a\`},

		{"**", ""},
		{`**\`, "**\\\n**"},
	}

	elements := config.Elements{
		"A": {
			Type:        node.TypeHanging.String(),
			Delimiter:   ".a",
			DoNotRemove: true,
		},
		"B": {
			Type:      node.TypeHanging.String(),
			Delimiter: ".b",
		},
		"C": {
			Type:      node.TypeWalled.String(),
			Delimiter: ">",
		},
		"MA": {
			Type:        node.TypePrefixed.String(),
			Delimiter:   `\`,
			DoNotRemove: true,
		},
		"MB": {
			Type:      node.TypeUniform.String(),
			Delimiter: "*",
		},
	}
	for _, c := range cases {
		name := fmt.Sprintf("%q", strings.ReplaceAll(c.in, "/", "2F")) // %2F is URL-escaped slash
		t.Run(name, func(t *testing.T) {
			test(t, elements, nil, c.in, c.out, 0)
		})
		t.Run("KeepNewlines-"+name, func(t *testing.T) {
			test(t, elements, nil, c.in, c.out, printer.KeepNewlines)
		})
	}
}

func test(t *testing.T, elements config.Elements, transformers []transformer.Transformer, in, out string, mode printer.Mode) {
	t.Helper()

	if elements == nil {
		elements = config.Elements{}
	}

	var want string
	if i := strings.Index(out, "$$"); i > -1 {
		if mode&printer.KeepNewlines != 0 {
			want = out[i+2:]
		} else {
			want = out[:i]
		}
	} else {
		want = out
	}

	printed := runPrint(t, elements, transformers, in, mode, *printTree)
	if printed != want {
		t.Errorf("got %q, want %q", printed, want)
	}

	if !*noReprint {
		previousPrint := printed
		for i := 0; ; i++ {
			if i > 2 {
				t.Errorf("too many reprints, skipping")
				break
			}

			reprinted := runPrint(t, elements, transformers, previousPrint, mode, *printTree)
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
			printedDefined := runPrint(t, elements, transformers, in, mode, false)
			if printedDefined != printed {
				t.Errorf("with defined text got %q, with undefined %q", printedDefined, printed)
			}
		}
	}
}

func runPrint(t *testing.T, elements config.Elements, transformers []transformer.Transformer, in string, mode printer.Mode, printTree bool) string {
	t.Helper()

	r := strings.NewReader(in)
	root, err := parser.Parse(r, elements.ParserElements())
	if err != nil {
		t.Fatal(err)
	}
	root = transformer.Apply(root, transformers)

	if printTree {
		var b strings.Builder
		if err := (node.Printer{node.PrintData}).Fprint(&b, root); err != nil {
			t.Fatal(err)
		}
		fmt.Println(b.String())
	}

	var b strings.Builder
	if err := (printer.Printer{Elements: elements.PrinterElements(), Mode: mode}).Fprint(&b, root); err != nil {
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
