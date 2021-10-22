package printer_test

import (
	"bytes"
	"flag"
	"fmt"
	"strings"
	"testing"

	"github.com/touchmarine/to/config"
	"github.com/touchmarine/to/node"
	"github.com/touchmarine/to/parser"
	"github.com/touchmarine/to/printer"
	"github.com/touchmarine/to/transformer"
)

var stringify = flag.Bool("stringify", false, "dump node")

/*
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

		{"`^", ""},
		{"`^a", ""},
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
		{"`a", ""},
		{"`a`", ""},
		{"`a\nb", "`a\nb\n`"},
		{"`a\n ", ""},
		{"`a\n\nb", "`a\n\nb\n`"},
		{"`a\n \nb", "`a\n \nb\n`"},
		{"`a\n\n\nb", "`a\n\n\nb\n`"},

		{"`\\a", ""},
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
			test(t, cfg, []byte(c.in), c.out)
		})
	}
}
*/

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
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("%q", c.in), func(t *testing.T) {
			test(t, nil, nil, []byte(c.in), c.out)
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
			test(t, elements, nil, []byte(c.in), c.out)
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
			test(t, elements, nil, []byte(c.in), c.out)
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
			test(t, elements, nil, []byte(c.in), c.out)
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
			test(t, elements, nil, []byte(c.in), c.out)
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
			test(t, elements, nil, []byte(c.in), c.out)
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

		// escape delimiter in text
		{"``", ""},
		{"``\nb", "``\nb\n`"},
		{"`a\n`", ""},
		{"`a\nb`", "`\\a\nb`\n\\`"},

		// unnecessary escape
		{"`\\a\nb", "`a\nb\n`"},
		{"`a\nb\n\\`", "`\\a\nb\n\\`\n\\`"},

		// nested
		{">`", ""},
		{">`a\n>b", "> `a\n> b\n> `"},
		{">\n>`", ""},
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
			test(t, elements, nil, []byte(c.in), c.out)
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
			elements := config.Elements{
				"T": {
					Type: node.TypeLeaf,
				},
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
				"MT": {
					Type: node.TypeText,
				},
			}

			test(t, elements, nil, []byte(c.in), c.out)
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
				"MT": {
					Type: node.TypeText,
				},
			}
			test(t, elements, nil, []byte(c.in), c.out)
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
			test(t, elements, nil, []byte(c.in), c.out)
		})
	}
}

func test(t *testing.T, elements config.Elements, transformers []transformer.Transformer, in []byte, out string) {
	t.Helper()

	if elements == nil {
		elements = config.Elements{}
	}

	printed := runPrint(t, elements, transformers, in, *stringify)
	if printed != out {
		t.Errorf("got %q, want %q", printed, out)
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

func runPrint(t *testing.T, elements config.Elements, transformers []transformer.Transformer, in []byte, stringify bool) string {
	t.Helper()

	r := bytes.NewReader(in)
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
