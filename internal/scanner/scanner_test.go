package scanner_test

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"text/tabwriter"
	"to/internal/scanner"
	"to/internal/token"
)

func ExampleScanner_Scan() {
	input := "| paragraph __emphasis__ // comment"
	s := scanner.New(input, scanner.ScanComments, func(err error, errCount uint) {})

	for {
		tok, lit := s.Scan()
		w := tabwriter.NewWriter(os.Stdout, 16, 8, 1, '\t', tabwriter.AlignRight)
		fmt.Fprintf(w, "%s\t\"%s\"\n", tok, lit)
		w.Flush()

		if tok == token.EOF {
			break
		}
	}
}

func TestScanner_Scan(t *testing.T) {
	cases := []struct {
		src string
		tok token.Token
		lit string
	}{
		{"\n", token.LINEFEED, "\n"},
		{"// comment", token.COMMENT, "// comment"},

		{"|", token.VLINE, "|"},
		{">", token.GT, ">"},
		{"a", token.TEXT, "a"},
	}

	for _, c := range cases {
		t.Run(literal(c.src), func(t *testing.T) {
			test(t, c.src, []tl{{c.tok, c.lit}})
		})
	}

}

func TestVerticalLine(t *testing.T) {
	cases := []struct {
		src           string
		tokenLiterals []tl
	}{
		{"|", []tl{{token.VLINE, "|"}}},
		{
			"||",
			[]tl{
				{token.VLINE, "|"},
				{token.VLINE, "|"},
			},
		},
		{
			"| |",
			[]tl{
				{token.VLINE, "|"},
				{token.INDENT, " "},
				{token.VLINE, "|"},
			},
		},
		{
			"|a|",
			[]tl{
				{token.VLINE, "|"},
				{token.TEXT, "a|"},
			},
		},
		{
			"| a |",
			[]tl{
				{token.VLINE, "|"},
				{token.INDENT, " "},
				{token.TEXT, "a |"},
			},
		},
	}

	for _, c := range cases {
		t.Run(literal(c.src), func(t *testing.T) {
			test(t, c.src, c.tokenLiterals)
		})
	}

}

func TestIndentation(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		cases := []struct {
			src string
			tok token.Token
			lit string
		}{
			{" ", token.INDENT, " "},
			{"  ", token.INDENT, "  "},
			{"        ", token.INDENT, "        "},
			{"\t", token.INDENT, "\t"},
			{"\t\t", token.INDENT, "\t\t"},
			{" \t", token.INDENT, " \t"},
			{"\t ", token.INDENT, "\t "},
		}

		for _, c := range cases {
			t.Run(literal(c.src), func(t *testing.T) {
				test(t, c.src, []tl{{c.tok, c.lit}})
			})
		}
	})

	t.Run("nested", func(t *testing.T) {
		cases := []struct {
			src        string
			tokenPairs []tl
		}{
			{
				"a\n b",
				[]tl{
					{token.TEXT, "a"},
					{token.LINEFEED, "\n"},
					{token.INDENT, " "},
					{token.TEXT, "b"},
				},
			},
			{
				" a\n  b",
				[]tl{
					{token.INDENT, " "},
					{token.TEXT, "a"},
					{token.LINEFEED, "\n"},
					{token.INDENT, "  "},
					{token.TEXT, "b"},
				},
			},
			{
				"  a\n   b",
				[]tl{
					{token.INDENT, "  "},
					{token.TEXT, "a"},
					{token.LINEFEED, "\n"},
					{token.INDENT, "   "},
					{token.TEXT, "b"},
				},
			},
			{
				"        a\n         b",
				[]tl{
					{token.INDENT, "        "},
					{token.TEXT, "a"},
					{token.LINEFEED, "\n"},
					{token.INDENT, "         "},
					{token.TEXT, "b"},
				},
			},
			{
				"a\n\tb",
				[]tl{
					{token.TEXT, "a"},
					{token.LINEFEED, "\n"},
					{token.INDENT, "\t"},
					{token.TEXT, "b"},
				},
			},
			{
				"\ta\n\t\tb",
				[]tl{
					{token.INDENT, "\t"},
					{token.TEXT, "a"},
					{token.LINEFEED, "\n"},
					{token.INDENT, "\t\t"},
					{token.TEXT, "b"},
				},
			},
			{
				"\t\ta\n\t\t\tb",
				[]tl{
					{token.INDENT, "\t\t"},
					{token.TEXT, "a"},
					{token.LINEFEED, "\n"},
					{token.INDENT, "\t\t\t"},
					{token.TEXT, "b"},
				},
			},
			{
				" \ta\n \t\tb",
				[]tl{
					{token.INDENT, " \t"},
					{token.TEXT, "a"},
					{token.LINEFEED, "\n"},
					{token.INDENT, " \t\t"},
					{token.TEXT, "b"},
				},
			},
			{
				"\t a\n\t\t b",
				[]tl{
					{token.INDENT, "\t "},
					{token.TEXT, "a"},
					{token.LINEFEED, "\n"},
					{token.INDENT, "\t\t "},
					{token.TEXT, "b"},
				},
			},
			{
				"\ta\n         b",
				[]tl{
					{token.INDENT, "\t"},
					{token.TEXT, "a"},
					{token.LINEFEED, "\n"},
					{token.INDENT, "         "},
					{token.TEXT, "b"},
				},
			},
			{
				"       a\n\tb",
				[]tl{
					{token.INDENT, "       "},
					{token.TEXT, "a"},
					{token.LINEFEED, "\n"},
					{token.INDENT, "\t"},
					{token.TEXT, "b"},
				},
			},
		}

		for _, c := range cases {
			t.Run(literal(c.src), func(t *testing.T) {
				test(t, c.src, c.tokenPairs)
			})
		}
	})
}

func TestNoIndentation(t *testing.T) {
	cases := []struct {
		src string
		tok token.Token
		lit string
	}{
		{"a ", token.TEXT, "a "},
		{"a  ", token.TEXT, "a  "},
		{"a        ", token.TEXT, "a        "},
		{"a\t", token.TEXT, "a\t"},
		{"a\t\t", token.TEXT, "a\t\t"},
		{"a \t", token.TEXT, "a \t"},
		{"a\t ", token.TEXT, "a\t "},
	}

	for _, c := range cases {
		t.Run(literal(c.src), func(t *testing.T) {
			test(t, c.src, []tl{{c.tok, c.lit}})
		})
	}

}

func TestDedent(t *testing.T) {
	cases := []struct {
		src        string
		tokenPairs []tl
	}{
		{
			" \na",
			[]tl{
				{token.INDENT, " "},
				{token.LINEFEED, "\n"},
				{token.TEXT, "a"},
			},
		},
		{
			" a\n  b\nc",
			[]tl{
				{token.INDENT, " "},
				{token.TEXT, "a"},
				{token.LINEFEED, "\n"},
				{token.INDENT, "  "},
				{token.TEXT, "b"},
				{token.LINEFEED, "\n"},
				{token.TEXT, "c"},
			},
		},
		{
			" a\n  b\n c\nd",
			[]tl{
				{token.INDENT, " "},
				{token.TEXT, "a"},
				{token.LINEFEED, "\n"},
				{token.INDENT, "  "},
				{token.TEXT, "b"},
				{token.LINEFEED, "\n"},
				{token.INDENT, " "},
				{token.TEXT, "c"},
				{token.LINEFEED, "\n"},
				{token.TEXT, "d"},
			},
		},
		{
			"\t\na",
			[]tl{
				{token.INDENT, "\t"},
				{token.LINEFEED, "\n"},
				{token.TEXT, "a"},
			},
		},
		{
			"\ta\n\t\tb\nc",
			[]tl{
				{token.INDENT, "\t"},
				{token.TEXT, "a"},
				{token.LINEFEED, "\n"},
				{token.INDENT, "\t\t"},
				{token.TEXT, "b"},
				{token.LINEFEED, "\n"},
				{token.TEXT, "c"},
			},
		},
		{
			"\ta\n\t\tb\n\tc\nd",
			[]tl{
				{token.INDENT, "\t"},
				{token.TEXT, "a"},
				{token.LINEFEED, "\n"},
				{token.INDENT, "\t\t"},
				{token.TEXT, "b"},
				{token.LINEFEED, "\n"},
				{token.INDENT, "\t"},
				{token.TEXT, "c"},
				{token.LINEFEED, "\n"},
				{token.TEXT, "d"},
			},
		},
		{
			// mixed tabs and spaces
			"\ta\n        \tb\n         c\nd",
			[]tl{
				{token.INDENT, "\t"},
				{token.TEXT, "a"},
				{token.LINEFEED, "\n"},
				{token.INDENT, "        \t"},
				{token.TEXT, "b"},
				{token.LINEFEED, "\n"},
				{token.INDENT, "         "},
				{token.TEXT, "c"},
				{token.LINEFEED, "\n"},
				{token.TEXT, "d"},
			},
		},
	}

	for _, c := range cases {
		t.Run(literal(c.src), func(t *testing.T) {
			test(t, c.src, c.tokenPairs)
		})
	}
}

func TestNoDedent(t *testing.T) {
	cases := []struct {
		src        string
		tokenPairs []tl
	}{
		{
			" \n a",
			[]tl{
				{token.INDENT, " "},
				{token.LINEFEED, "\n"},
				{token.INDENT, " "},
				{token.TEXT, "a"},
			},
		},
		{
			"\t\n\ta",
			[]tl{
				{token.INDENT, "\t"},
				{token.LINEFEED, "\n"},
				{token.INDENT, "\t"},
				{token.TEXT, "a"},
			},
		},
		{
			"\t \n\t a",
			[]tl{
				{token.INDENT, "\t "},
				{token.LINEFEED, "\n"},
				{token.INDENT, "\t "},
				{token.TEXT, "a"},
			},
		},
		{
			" \t\n \ta",
			[]tl{
				{token.INDENT, " \t"},
				{token.LINEFEED, "\n"},
				{token.INDENT, " \t"},
				{token.TEXT, "a"},
			},
		},
		{
			"\t\n        a",
			[]tl{
				{token.INDENT, "\t"},
				{token.LINEFEED, "\n"},
				{token.INDENT, "        "},
				{token.TEXT, "a"},
			},
		},
		{
			"        \n\ta",
			[]tl{
				{token.INDENT, "        "},
				{token.LINEFEED, "\n"},
				{token.INDENT, "\t"},
				{token.TEXT, "a"},
			},
		},
		{
			"\t\t\n\t\ta",
			[]tl{
				{token.INDENT, "\t\t"},
				{token.LINEFEED, "\n"},
				{token.INDENT, "\t\t"},
				{token.TEXT, "a"},
			},
		},
		{
			"\t\t\n\t        a",
			[]tl{
				{token.INDENT, "\t\t"},
				{token.LINEFEED, "\n"},
				{token.INDENT, "\t        "},
				{token.TEXT, "a"},
			},
		},
		{
			"\t\t\n                a",
			[]tl{
				{token.INDENT, "\t\t"},
				{token.LINEFEED, "\n"},
				{token.INDENT, "                "},
				{token.TEXT, "a"},
			},
		},
	}

	for _, c := range cases {
		t.Run(literal(c.src), func(t *testing.T) {
			test(t, c.src, c.tokenPairs)
		})
	}
}

func TestGreaterThan(t *testing.T) {
	cases := []struct {
		src           string
		tokenLiterals []tl
	}{
		{">", []tl{{token.GT, ">"}}},
		{
			">>",
			[]tl{
				{token.GT, ">"},
				{token.GT, ">"},
			},
		},
		{
			"> >",
			[]tl{
				{token.GT, ">"},
				{token.INDENT, " "},
				{token.GT, ">"},
			},
		},
		{
			">a>",
			[]tl{
				{token.GT, ">"},
				{token.TEXT, "a>"},
			},
		},
		{
			"> a >",
			[]tl{
				{token.GT, ">"},
				{token.INDENT, " "},
				{token.TEXT, "a >"},
			},
		},
	}

	for _, c := range cases {
		t.Run(literal(c.src), func(t *testing.T) {
			test(t, c.src, c.tokenLiterals)
		})
	}

}

func TestGraveAccents(t *testing.T) {
	cases := []struct {
		src           string
		tokenLiterals []tl
	}{
		{"``", []tl{{token.GRAVEACCENTS, "``"}}},
		{"```", []tl{{token.GRAVEACCENTS, "```"}}},
		{"``````````", []tl{{token.GRAVEACCENTS, "``````````"}}},
		{
			"`` ``",
			[]tl{
				{token.GRAVEACCENTS, "``"},
				{token.INDENT, " "},
				{token.GRAVEACCENTS, "``"},
			},
		},
		{
			"``\n``",
			[]tl{
				{token.GRAVEACCENTS, "``"},
				{token.LINEFEED, "\n"},
				{token.GRAVEACCENTS, "``"},
			},
		},
	}

	for _, c := range cases {
		t.Run(literal(c.src), func(t *testing.T) {
			test(t, c.src, c.tokenLiterals)
		})
	}

}

func TestNoGraveAccents(t *testing.T) {
	cases := []struct {
		src           string
		tokenLiterals []tl
	}{
		{"`", []tl{{token.TEXT, "`"}}},
	}

	for _, c := range cases {
		t.Run(literal(c.src), func(t *testing.T) {
			test(t, c.src, c.tokenLiterals)
		})
	}

}

func TestUnderscores(t *testing.T) {
	cases := []struct {
		src           string
		tokenLiterals []tl
	}{
		{"__", []tl{{token.UNDERSCORES, "__"}}},
		{
			"__a",
			[]tl{
				{token.UNDERSCORES, "__"},
				{token.TEXT, "a"},
			},
		},
		{
			"a__",
			[]tl{
				{token.TEXT, "a"},
				{token.UNDERSCORES, "__"},
			},
		},
		{
			"_a_",
			[]tl{
				{token.TEXT, "_a_"},
			},
		},
		{
			"|__",
			[]tl{
				{token.VLINE, "|"},
				{token.UNDERSCORES, "__"},
			},
		},
		{
			"|a__",
			[]tl{
				{token.VLINE, "|"},
				{token.TEXT, "a"},
				{token.UNDERSCORES, "__"},
			},
		},
	}

	for _, c := range cases {
		t.Run(literal(c.src), func(t *testing.T) {
			test(t, c.src, c.tokenLiterals)
		})
	}

}

// token-literal pair struct
type tl struct {
	tok token.Token
	lit string
}

func test(t *testing.T, src string, tokenPairs []tl) {
	t.Helper()

	s := scanner.New(
		src,
		scanner.ScanComments,
		func(err error, errCount uint) {
			if errCount > 0 {
				t.Fatalf("scanning error: %s", err)
			}
		},
	)

	var errCount int
	var scanned []tl

	var count int
	for {
		tok, lit := s.Scan()
		if tok == token.EOF {
			break
		}

		scanned = append(scanned, tl{tok, lit})

		if count >= len(tokenPairs) {
			errCount++
			count++
			continue
		}

		tp := tokenPairs[count] // token-pair

		if tok != tp.tok || lit != tp.lit {
			errCount++
		}

		count++
	}

	if count < len(tokenPairs) {
		errCount += len(tokenPairs) - count
	}

	if errCount > 0 {
		var bg strings.Builder
		wg := tabwriter.NewWriter(&bg, 0, 0, 8, ' ', 0)
		for i, tp := range scanned {
			fmt.Fprintf(wg, "\t%d\t%s\t%s\n", i, tp.tok, strconv.Quote(tp.lit))
		}
		wg.Flush()

		var bw strings.Builder
		ww := tabwriter.NewWriter(&bw, 0, 0, 8, ' ', 0)
		for i, tp := range tokenPairs {
			fmt.Fprintf(ww, "\t%d\t%s\t%s\n", i, tp.tok, strconv.Quote(tp.lit))
		}
		ww.Flush()

		t.Errorf(
			"%d errors:\ngot %d tokens:\n%swant %d:\n%s",
			errCount,
			count,
			bg.String(),
			len(tokenPairs),
			bw.String(),
		)
	}
}

func literal(s string) string {
	q := strconv.Quote(s)
	return q[1 : len(q)-1]
}
