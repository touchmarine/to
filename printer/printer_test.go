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

var stringify = flag.Bool("stringify", false, "dump node")

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
		{"a\nb", "a b"},
		{"a\n b", "a b"},
		{"a\n\n b", "a\n\nb"},
		{"ab\n c", "ab c"},
		{"ab\n\n c", "ab\n\nc"},

		// interrupted by empty blocks
		{"a\n>\n*\nb", "a\nb"},
		{"a\n>b\n*\nc", "a\n\n> b\n\nc"},
		{"a\n>\n*b\nc", "a\n\n* b\n\nc"},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("%q", c.in), func(t *testing.T) {
			elements := config.Elements{
				"A": {
					Type:      node.TypeWalled,
					Delimiter: ">",
				},
				"B": {
					Type:      node.TypeWalled,
					Delimiter: "*",
				},
			}
			test(t, elements, nil, c.in, c.out)
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

	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			elements := config.Elements{
				"A": {
					Type:      node.TypeVerbatimLine,
					Delimiter: ".a",
				},
				"B": {
					Type:      node.TypeWalled,
					Delimiter: ">",
				},
			}
			test(t, elements, nil, c.in, c.out)
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
		{"-a\n b", "- a b"},
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

	for _, c := range cases {
		t.Run(fmt.Sprintf("%q", c.in), func(t *testing.T) {
			elements := config.Elements{
				"A": {
					Type:      node.TypeHanging,
					Delimiter: "-",
				},
				"B": {
					Type:      node.TypeWalled,
					Delimiter: ">",
				},
			}
			test(t, elements, nil, c.in, c.out)
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
		{"==a\n  b", "== a b"},
		{"==a", "== a"},
		{"==a\n  b", "== a b"},
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

	for _, c := range cases {
		t.Run(fmt.Sprintf("%q", c.in), func(t *testing.T) {
			elements := config.Elements{
				"A": {
					Type:      node.TypeRankedHanging,
					Delimiter: "=",
				},
				"B": {
					Type:      node.TypeWalled,
					Delimiter: ">",
				},
			}
			test(t, elements, nil, c.in, c.out)
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
		{"+a\n+b", "+ a b"},
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

	for _, c := range cases {
		t.Run(fmt.Sprintf("%q", c.in), func(t *testing.T) {
			elements := config.Elements{
				"A": {
					Type:      node.TypeWalled,
					Delimiter: "+",
				},
				"B": {
					Type:      node.TypeWalled,
					Delimiter: ">",
				},
			}
			test(t, elements, nil, c.in, c.out)
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
		{"!a\n!b", "! a b"},
		{"!a\n!\n!b", "! a b"},
		{"!a\n!\n!\n!b", "! a b"},
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

	for _, c := range cases {
		t.Run(fmt.Sprintf("%q", c.in), func(t *testing.T) {
			elements := config.Elements{
				"A": {
					Type:      node.TypeVerbatimWalled,
					Delimiter: "!",
				},
				"B": {
					Type:      node.TypeWalled,
					Delimiter: ">",
				},
			}
			test(t, elements, nil, c.in, c.out)
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

	for _, c := range cases {
		t.Run(fmt.Sprintf("%q", c.in), func(t *testing.T) {
			elements := config.Elements{
				"A": {
					Type:      node.TypeFenced,
					Delimiter: "`",
				},
				"B": {
					Type:      node.TypeWalled,
					Delimiter: ">",
				},
			}
			test(t, elements, nil, c.in, c.out)
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

		for _, c := range cases {
			t.Run(fmt.Sprintf("%q", c.in), func(t *testing.T) {
				elements := config.Elements{
					"A": {
						Type:      node.TypeWalled,
						Delimiter: ">",
					},
					"B": {
						Type:      node.TypeWalled,
						Delimiter: "*",
					},
				}
				transformers := []transformer.Transformer{
					paragraph.Transformer{paragraph.Map{
						node.TypeLeaf: "PA",
					}},
				}
				test(t, elements, transformers, c.in, c.out)
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

		for _, c := range cases {
			t.Run(fmt.Sprintf("%q", c.in), func(t *testing.T) {
				elements := config.Elements{
					"A": {
						Type:      node.TypeHanging,
						Delimiter: "-",
					},
					"B": {
						Type:      node.TypeWalled,
						Delimiter: ">",
					},
					"C": {
						Type:      node.TypeWalled,
						Delimiter: "*",
					},
				}
				transformers := []transformer.Transformer{
					group.Transformer{group.Map{
						"A": "LA",
					}},
				}
				test(t, elements, transformers, c.in, c.out)
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

		for _, c := range cases {
			t.Run(fmt.Sprintf("%q", c.in), func(t *testing.T) {
				elements := config.Elements{
					"A": {
						Type:      node.TypeVerbatimWalled,
						Delimiter: "!",
					},
					"B": {
						Type:      node.TypeWalled,
						Delimiter: "+",
					},
					"C": {
						Type:      node.TypeWalled,
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
				test(t, elements, transformers, c.in, c.out)
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
		}

		for _, c := range cases {
			t.Run(fmt.Sprintf("%q", c.in), func(t *testing.T) {
				elements := config.Elements{
					"A": {
						Type:      node.TypeUniform,
						Delimiter: "(",
					},
					"B": {
						Type:      node.TypeUniform,
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
				test(t, elements, transformers, c.in, c.out)
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
		{"**\na", "** a**"},
		{"**\na**", "** a**"},
		{"**\na**b", "** a**b"},
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

	for _, c := range cases {
		t.Run(fmt.Sprintf("%q", c.in), func(t *testing.T) {
			elements := config.Elements{
				"MA": {
					Type:      node.TypeUniform,
					Delimiter: "*",
				},
				"MB": {
					Type:      node.TypeUniform,
					Delimiter: "_",
				},
				"MC": {
					Type:      node.TypeUniform,
					Delimiter: "(",
				},
			}
			test(t, elements, nil, c.in, c.out)
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
		{"``\na", "`` a``"},
		{"``\na``", "`` a``"},
		{"``\na``b", "`` a``b"},
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
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("%q", c.in), func(t *testing.T) {
			elements := config.Elements{
				"MA": {
					Type:      node.TypeEscaped,
					Delimiter: "`",
				},
				"MB": {
					Type:      node.TypeUniform,
					Delimiter: "_",
				},
				"MC": {
					Type:      node.TypeEscaped,
					Delimiter: "[",
				},
			}
			test(t, elements, nil, c.in, c.out)
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

	for _, c := range cases {
		t.Run(fmt.Sprintf("%q", c.in), func(t *testing.T) {
			elements := config.Elements{
				"MA": {
					Type:      node.TypePrefixed,
					Delimiter: `\`,
				},
			}
			test(t, elements, nil, c.in, c.out)
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

		for _, c := range cases {
			t.Run(fmt.Sprintf("%q", c.in), func(t *testing.T) {
				elements := config.Elements{
					"MA": {
						Type:        node.TypePrefixed,
						Delimiter:   `\`,
						DoNotRemove: true,
					},
				}
				test(t, elements, nil, c.in, c.out)
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

		for _, c := range cases {
			t.Run(fmt.Sprintf("%q", c.in), func(t *testing.T) {
				elements := config.Elements{
					"MA": {
						Type:      node.TypePrefixed,
						Delimiter: "a:",
						Matcher:   "url",
					},
				}
				test(t, elements, nil, c.in, c.out)
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
		{"\\*\n\\*", `\* *`}, // * *

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

	for _, c := range cases {
		name := strings.ReplaceAll(c.in, "/", "2F") // %2F is URL-escaped slash
		t.Run(fmt.Sprintf("%q", name), func(t *testing.T) {
			elements := config.Elements{
				"A": {
					Type:      node.TypeHanging,
					Delimiter: "*",
				},
				"B": {
					Type:      node.TypeWalled,
					Delimiter: ">",
				},
				"C": {
					Type:      node.TypeFenced,
					Delimiter: "`",
				},
				"MA": {
					Type:      node.TypeUniform,
					Delimiter: "*",
				},
				"MB": {
					Type:      node.TypeEscaped,
					Delimiter: "`",
				},
				// use "{" as it doesn't need escaping
				// in -run test regex as "(" or "["
				"MC": {
					Type:      node.TypeUniform,
					Delimiter: "{",
				},
				"MD": {
					Type:      node.TypePrefixed,
					Delimiter: "http://",
					Matcher:   "url",
				},
			}

			test(t, elements, nil, c.in, c.out)
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
		{"\\*\n\\*", `\* *`}, // * *

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

	for _, c := range cases {
		name := strings.ReplaceAll(c.in, "/", "2F") // %2F is URL-escaped slash
		t.Run(fmt.Sprintf("%q", name), func(t *testing.T) {
			elements := config.Elements{
				"A": {
					Type:      node.TypeHanging,
					Delimiter: "*",
				},
				"B": {
					Type:      node.TypeWalled,
					Delimiter: ">",
				},
				"C": {
					Type:      node.TypeFenced,
					Delimiter: "`",
				},
				"MA": {
					Type:      node.TypeUniform,
					Delimiter: "*",
				},
				"MB": {
					Type:      node.TypeEscaped,
					Delimiter: "`",
				},
				"MC": {
					Type:        node.TypePrefixed,
					Delimiter:   `\`,
					DoNotRemove: true,
				},
				// use "{" as it doesn't need escaping
				// in -run test regex as "(" or "["
				"MD": {
					Type:      node.TypeUniform,
					Delimiter: "{",
				},
				"ME": {
					Type:      node.TypePrefixed,
					Delimiter: "http://",
					Matcher:   "url",
				},
			}
			test(t, elements, nil, c.in, c.out)
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

	for _, c := range cases {
		name := strings.ReplaceAll(c.in, "/", "2F") // %2F is URL-escaped slash
		t.Run(fmt.Sprintf("%q", name), func(t *testing.T) {
			elements := config.Elements{
				"A": {
					Type:        node.TypeHanging,
					Delimiter:   ".a",
					DoNotRemove: true,
				},
				"B": {
					Type:      node.TypeHanging,
					Delimiter: ".b",
				},
				"C": {
					Type:      node.TypeWalled,
					Delimiter: ">",
				},
				"MA": {
					Type:        node.TypePrefixed,
					Delimiter:   `\`,
					DoNotRemove: true,
				},
				"MB": {
					Type:      node.TypeUniform,
					Delimiter: "*",
				},
			}
			test(t, elements, nil, c.in, c.out)
		})
	}
}

func test(t *testing.T, elements config.Elements, transformers []transformer.Transformer, in string, out string) {
	t.Helper()

	if elements == nil {
		elements = config.Elements{}
	}

	printed := runPrint(t, elements, transformers, in, *stringify)
	if printed != out {
		t.Errorf("got %q, want %q", printed, out)
	}

	previousPrint := printed
	for i := 0; ; i++ {
		if i > 2 {
			t.Errorf("too many reprints, skipping")
			break
		}

		reprinted := runPrint(t, elements, transformers, previousPrint, *stringify)
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
				Type: node.TypeLeaf,
			}
		}
		if !hasText {
			if _, ok := elements["_MT"]; ok {
				t.Fatal("element _MT already exists")
			}
			elements["_MT"] = config.Element{
				Type: node.TypeText,
			}
		}
		printedDefined := runPrint(t, elements, transformers, in, false)
		if printedDefined != printed {
			t.Errorf("with defined text got %q, with undefined %q", printedDefined, printed)
		}
	}
}

func runPrint(t *testing.T, elements config.Elements, transformers []transformer.Transformer, in string, stringify bool) string {
	t.Helper()

	r := strings.NewReader(in)
	root, err := parser.Parse(r, config.ToParserElements(elements))
	if err != nil {
		t.Fatal(err)
	}
	root = transformer.Apply(root, transformers)

	if stringify {
		s, err := node.StringifyDetailed(root)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(s)
	}

	var b strings.Builder
	if err := printer.Fprint(&b, config.ToPrinterElements(elements), root); err != nil {
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
		if e.Type == node.TypeLeaf {
			hasLeaf = true
		} else if e.Type == node.TypeText {
			hasText = true
		}
	}
	return hasLeaf, hasText
}
