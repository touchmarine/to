package printer_test

import (
	"bytes"
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
		{"^", ""},
		{"^a", "^ a"},
		{"^\n a", "^ a"},
		{"^a\n b", "^ a b"},
		{"^a\n ^b", "^ a\n\n  ^ b"},
		{"^a\n\n ^b", "^ a\n\n  ^ b"},
		{"^a\n \n ^b", "^ a\n\n  ^ b"},
		{"^a\n\n\n ^b", "^ a\n\n  ^ b"},

		{"^!a\n !b", "^ !a\n  !b"},

		{"=^", ""},
		{"=^a", "= ^ a"},
		{"=\n ^", ""},
		{"=\n ^a", "= ^ a"},

		{".image^", ".image ^"},
		{".image^a", ".image ^a"},

		{">^", ""},
		{">^a", "> ^ a"},
		{">\n>^", ""},
		{">\n>^a", "> ^ a"},

		{"`^", "`^\n`"},
		{"`^a", "`^a\n`"},
		{"`\n ^", "`\n ^\n`"},
		{"`\n ^a", "`\n ^a\n`"},

		{"-^", ""},
		{"-^a", "- ^ a"},
		{"-\n ^", ""},
		{"-\n ^a", "- ^ a"},

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
		{"!", ""},
		{"!a", "!a"},
		{"! a", "! a"},
		{"!\n!", ""},
		{"!a\n!b", "!a\n!b"},
		{"! a\n! b", "! a\n! b"},
		{"!a\n!\n!b", "!a\n!\n!b"},

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
		{"!a\nb", "!a\nb"},
		{"!a\n!b\nc", "!a\n!b\nc"},
		{"a\n!b", "a\n\n!b"},
		{"a\n!b\n!c", "a\n\n!b\n!c"},

		{"a\n+b", "a\n+ b"},
		{"a\n+b\n+c", "a\n+ b c"},
		{"+a\nb", "+ a\n\nb"},
		{"+a\n+b\nc", "+ a b\n\nc"},

		// multiple blocks
		{"a\n\nb", "a\n\nb"},
		{"a\n\n\nb", "a\n\nb"},
		{"a\n \nb", "a\n\nb"},
		{"a\n^b", "a\n\n^ b"},
		{"^a\nb", "^ a\n\nb"},
		{"^a\n^b", "^ a\n^ b"},
		{"^a\n>b", "^ a\n\n> b"},
		{">a\n\n>b", "> a\n\n> b"},
		{">a\n \n>b", "> a\n\n> b"},
		{">a\n\n\n>b", "> a\n\n> b"},
		{">a\n^b", "> a\n\n^ b"},

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

		// prefixed
		{"ahttp://", "a"},
		{"ahttp://b", "ahttp://b"},

		// composite
		{"(())[[]]", ""},
		{"((a))[[b]]", "((a))[[b]]"},
		{"a((b))[[c]]", "a((b))[[c]]"},
		{"(( a ))[[b]]", "(( a ))[[b]]"},
		{"((a))[[ b ]]", "((a))[[ b ]]"},
		{"((((c))[[d]]a))[[b]]", "((((c))[[d]]a))[[b]]"}, // G(G(c)L(d)a)L(b)

		{"((``a))[[b]]", "((``a))[[b]]``))"},
		{"((``a``))[[b]]", "((``a``))[[b]]"},

		// escape common elements
		{"a\\``", "a\\``"},
		{`a\[[`, `a\[[`},
		{`a\//`, `a\//`},
		{`a\(())[[]]`, `a\(())`},
		{`a(())\[[]]`, `a\[[]]`},
		{`a\((a))[[b]]`, `a\((a))[[b]]`},
		{`a((a))\[[b]]`, `a((a))\[[b]]`},
		{`a\((a))\[[b]]`, `a\((a))\[[b]]`},
		{`a\http://`, "ahttp:"},
	}

	for _, c := range cases {
		name := strings.ReplaceAll(c.in, "/", "2F") // %2F is URL-escaped slash

		t.Run(fmt.Sprintf("%q", name), func(t *testing.T) {
			test(t, config.Default, []byte(c.in), c.out)
		})
	}
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
			conf := &config.Config{
				Elements: []config.Element{
					{
						Name: "T",
						Type: node.TypeLeaf,
					},
					{
						Name:      "A",
						Type:      node.TypeHanging,
						Delimiter: "*",
					},
					{
						Name:      "B",
						Type:      node.TypeWalled,
						Delimiter: ">",
					},
					{
						Name:      "C",
						Type:      node.TypeFenced,
						Delimiter: "`",
					},
					{
						Name:      "MA",
						Type:      node.TypeUniform,
						Delimiter: "*",
					},
					{
						Name:      "MB",
						Type:      node.TypeEscaped,
						Delimiter: "`",
					},
					// use "{" as it doesn't need escaping
					// in -run test regex as "(" or "["
					{
						Name:      "MC",
						Type:      node.TypeUniform,
						Delimiter: "{",
					},
					{
						Name:      "MD",
						Type:      node.TypePrefixed,
						Delimiter: "http://",
						Matcher:   "url",
					},
					{
						Name: "MT",
						Type: node.TypeText,
					},
				},
			}

			test(t, conf, []byte(c.in), c.out)
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
			conf := &config.Config{
				Elements: []config.Element{
					{
						Name:      "A",
						Type:      node.TypeHanging,
						Delimiter: "*",
					},
					{
						Name:      "B",
						Type:      node.TypeWalled,
						Delimiter: ">",
					},
					{
						Name:      "C",
						Type:      node.TypeFenced,
						Delimiter: "`",
					},
					{
						Name:      "MA",
						Type:      node.TypeUniform,
						Delimiter: "*",
					},
					{
						Name:      "MB",
						Type:      node.TypeEscaped,
						Delimiter: "`",
					},
					{
						Name:        "MC",
						Type:        node.TypePrefixed,
						Delimiter:   `\`,
						DoNotRemove: true,
					},
					// use "{" as it doesn't need escaping
					// in -run test regex as "(" or "["
					{
						Name:      "MD",
						Type:      node.TypeUniform,
						Delimiter: "{",
					},
					{
						Name:      "ME",
						Type:      node.TypePrefixed,
						Delimiter: "http://",
						Matcher:   "url",
					},
					{
						Name: "MT",
						Type: node.TypeText,
					},
				},
			}

			test(t, conf, []byte(c.in), c.out)
		})
	}
}

func TestDoNotRemove(t *testing.T) {
	cases := []struct {
		in  string
		out string
	}{
		{".toc", ".toc"},
		{".toc ", ".toc"},

		{">.toc ", ">.toc"},
		{">>.toc ", ">>.toc"},

		{".c", ""},
		{".c.toc ", ".c.toc"},

		{`\`, `\`},
		{`a\`, `a\`},

		{"**", ""},
		{`**\`, "**\\\n**"},
	}

	for _, c := range cases {
		name := strings.ReplaceAll(c.in, "/", "2F") // %2F is URL-escaped slash

		t.Run(fmt.Sprintf("%q", name), func(t *testing.T) {
			conf := &config.Config{
				Elements: []config.Element{
					{
						Name:        "A",
						Type:        node.TypeHanging,
						Delimiter:   ".toc",
						DoNotRemove: true,
					},
					{
						Name:      "B",
						Type:      node.TypeWalled,
						Delimiter: ">",
					},
					{
						Name:      "C",
						Type:      node.TypeHanging,
						Delimiter: ".c",
					},
					{
						Name:        "MA",
						Type:        node.TypePrefixed,
						Delimiter:   `\`,
						DoNotRemove: true,
					},
					{
						Name:      "MB",
						Type:      node.TypeUniform,
						Delimiter: "*",
					},
				},
			}

			test(t, conf, []byte(c.in), c.out)
		})
	}
}

func test(t *testing.T, conf *config.Config, in []byte, out string) {
	t.Helper()

	r := bytes.NewReader(in)
	blocks, errs := parser.Parse(r, conf.ParserElements())
	if errs != nil {
		t.Fatal(errs)
	}

	nodes := node.BlocksToNodes(blocks)
	nodes = transformer.Apply(nodes, conf.DefaultTransformers())

	var b strings.Builder
	printer.Fprint(&b, conf.PrinterElements(), nodes)

	printed := b.String()

	if printed != out {
		t.Errorf("got %q, want %q", printed, out)
	}
}
