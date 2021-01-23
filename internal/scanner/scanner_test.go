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

func TestIndent(t *testing.T) {
	cases := []struct {
		src        string
		tokenPairs []tl
	}{
		{
			" a",
			[]tl{
				{token.INDENT, " "},
				{token.TEXT, "a"},
			},
		},
		{
			"a\n b",
			[]tl{
				{token.TEXT, "a"},
				{token.LINEFEED, "\n"},
				{token.INDENT, " "},
				{token.TEXT, "b"},
			},
		},
	}

	for _, c := range cases {
		t.Run(literal(c.src), func(t *testing.T) {
			test(t, c.src, c.tokenPairs)
		})
	}
}

func TestDedent(t *testing.T) {
	cases := []struct {
		src        string
		tokenPairs []tl
	}{
		{
			" a\nb",
			[]tl{
				{token.INDENT, " "},
				{token.TEXT, "a"},
				{token.LINEFEED, "\n"},
				{token.DEDENT, ""},
				{token.TEXT, "b"},
			},
		},
		{
			"  a\n b\nc",
			[]tl{
				{token.INDENT, "  "},
				{token.TEXT, "a"},
				{token.LINEFEED, "\n"},
				{token.DEDENT, ""},
				{token.TEXT, "b"},
				{token.LINEFEED, "\n"},
				{token.DEDENT, ""},
				{token.TEXT, "c"},
			},
		},
	}

	for _, c := range cases {
		t.Run(literal(c.src), func(t *testing.T) {
			test(t, c.src, c.tokenPairs)
		})
	}
}

func TestEqualIndentation(t *testing.T) {
	cases := []struct {
		src        string
		tokenPairs []tl
	}{
		{
			"a\nb",
			[]tl{
				{token.TEXT, "a"},
				{token.LINEFEED, "\n"},
				{token.TEXT, "b"},
			},
		},
		{
			" a\n b",
			[]tl{
				{token.INDENT, "  "},
				{token.TEXT, "a"},
				{token.LINEFEED, "\n"},
				{token.TEXT, "b"},
			},
		},
	}

	for _, c := range cases {
		t.Run(literal(c.src), func(t *testing.T) {
			test(t, c.src, c.tokenPairs)
		})
	}
}

/*
func TestScanner_Scan(t *testing.T) {
	cases := []struct {
		src string
		tok token.Token
		lit string
	}{
		{"\n", token.LINEFEED, "\n"},
		{"// comment", token.Comment, "// comment"},

		{"|", token.Pipeline, "|"},
		{">", token.GreaterThan, ">"},
	}

	for _, c := range cases {
		t.Run(strconv.Quote(c.src), func(t *testing.T) {
			test(t, c.src, []tl{{c.tok, c.lit}})
		})
	}

}

func TestPipeline(t *testing.T) {
	cases := []struct {
		src           string
		tokenLiterals []tl
	}{
		{"|", []tl{{token.Pipeline, "|"}}},
		{
			"||",
			[]tl{
				{token.Pipeline, "|"},
				{token.Pipeline, "|"},
			},
		},
		{
			"| |",
			[]tl{
				{token.Pipeline, "|"},
				{token.Pipeline, "|"},
			},
		},
		{
			"|a|",
			[]tl{
				{token.Pipeline, "|"},
				{token.Text, "a|"},
			},
		},
		{
			"| a |",
			[]tl{
				{token.Pipeline, "|"},
				{token.Text, "a |"},
			},
		},
	}

	for _, c := range cases {
		t.Run(strconv.Quote(c.src), func(t *testing.T) {
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
			{" ", token.Indent, " "},
			{"  ", token.Indent, "  "},
			{"        ", token.Indent, "        "},
			{"\t", token.Indent, "\t"},
			{"\t\t", token.Indent, "\t\t"},
			{" \t", token.Indent, " \t"},
			{"\t ", token.Indent, "\t "},
		}

		for _, c := range cases {
			t.Run(strconv.Quote(c.src), func(t *testing.T) {
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
				" a\n  b",
				[]tl{
					{token.Indent, " "},
					{token.Text, "a"},
					{token.LINEFEED, "\n"},
					{token.Indent, "  "},
					{token.Text, "b"},
				},
			},
			{
				"  a\n   b",
				[]tl{
					{token.Indent, "  "},
					{token.Text, "a"},
					{token.LINEFEED, "\n"},
					{token.Indent, "   "},
					{token.Text, "b"},
				},
			},
			{
				"        a\n         b",
				[]tl{
					{token.Indent, "        "},
					{token.Text, "a"},
					{token.LINEFEED, "\n"},
					{token.Indent, "         "},
					{token.Text, "b"},
				},
			},
			{
				"\ta\n\t\tb",
				[]tl{
					{token.Indent, "\t"},
					{token.Text, "a"},
					{token.LINEFEED, "\n"},
					{token.Indent, "\t\t"},
					{token.Text, "b"},
				},
			},
			{
				"\t\ta\n\t\t\tb",
				[]tl{
					{token.Indent, "\t\t"},
					{token.Text, "a"},
					{token.LINEFEED, "\n"},
					{token.Indent, "\t\t\t"},
					{token.Text, "b"},
				},
			},
			{
				" \ta\n \t\tb",
				[]tl{
					{token.Indent, " \t"},
					{token.Text, "a"},
					{token.LINEFEED, "\n"},
					{token.Indent, " \t\t"},
					{token.Text, "b"},
				},
			},
			{
				"\t a\n\t\t b",
				[]tl{
					{token.Indent, "\t "},
					{token.Text, "a"},
					{token.LINEFEED, "\n"},
					{token.Indent, "\t\t "},
					{token.Text, "b"},
				},
			},
			{
				"\ta\n         b",
				[]tl{
					{token.Indent, "\t"},
					{token.Text, "a"},
					{token.LINEFEED, "\n"},
					{token.Indent, "         "},
					{token.Text, "b"},
				},
			},
			{
				"       a\n\tb",
				[]tl{
					{token.Indent, "       "},
					{token.Text, "a"},
					{token.LINEFEED, "\n"},
					{token.Indent, "\t"},
					{token.Text, "b"},
				},
			},
		}

		for _, c := range cases {
			t.Run(strconv.Quote(c.src), func(t *testing.T) {
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
		{"a ", token.Text, "a "},
		{"a  ", token.Text, "a  "},
		{"a        ", token.Text, "a        "},
		{"a\t", token.Text, "a\t"},
		{"a\t\t", token.Text, "a\t\t"},
		{"a \t", token.Text, "a \t"},
		{"a\t ", token.Text, "a\t "},
	}

	for _, c := range cases {
		t.Run(strconv.Quote(c.src), func(t *testing.T) {
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
				{token.Indent, " "},
				{token.LINEFEED, "\n"},
				{token.Dedent, ""},
				{token.Text, "a"},
			},
		},
		{
			" a\n  b\nc",
			[]tl{
				{token.Indent, " "},
				{token.Text, "a"},
				{token.LINEFEED, "\n"},
				{token.Indent, "  "},
				{token.Text, "b"},
				{token.LINEFEED, "\n"},
				{token.Dedent, ""},
				{token.Dedent, ""},
				{token.Text, "c"},
			},
		},
		{
			" a\n  b\n c\nd",
			[]tl{
				{token.Indent, " "},
				{token.Text, "a"},
				{token.LINEFEED, "\n"},
				{token.Indent, "  "},
				{token.Text, "b"},
				{token.LINEFEED, "\n"},
				{token.Dedent, ""},
				{token.Text, "c"},
				{token.LINEFEED, "\n"},
				{token.Dedent, ""},
				{token.Text, "d"},
			},
		},
		{
			"\t\na",
			[]tl{
				{token.Indent, "\t"},
				{token.LINEFEED, "\n"},
				{token.Dedent, ""},
				{token.Text, "a"},
			},
		},
		{
			"\ta\n\t\tb\nc",
			[]tl{
				{token.Indent, "\t"},
				{token.Text, "a"},
				{token.LINEFEED, "\n"},
				{token.Indent, "\t\t"},
				{token.Text, "b"},
				{token.LINEFEED, "\n"},
				{token.Dedent, ""},
				{token.Dedent, ""},
				{token.Text, "c"},
			},
		},
		{
			"\ta\n\t\tb\n\tc\nd",
			[]tl{
				{token.Indent, "\t"},
				{token.Text, "a"},
				{token.LINEFEED, "\n"},
				{token.Indent, "\t\t"},
				{token.Text, "b"},
				{token.LINEFEED, "\n"},
				{token.Dedent, ""},
				{token.Text, "c"},
				{token.LINEFEED, "\n"},
				{token.Dedent, ""},
				{token.Text, "d"},
			},
		},
		{
			// mixed tabs and spaces
			"\ta\n        \tb\n         c\nd",
			[]tl{
				{token.Indent, "\t"},
				{token.Text, "a"},
				{token.LINEFEED, "\n"},
				{token.Indent, "        \t"},
				{token.Text, "b"},
				{token.LINEFEED, "\n"},
				{token.Dedent, ""},
				{token.Text, "c"},
				{token.LINEFEED, "\n"},
				{token.Dedent, ""},
				{token.Text, "d"},
			},
		},
	}

	for _, c := range cases {
		t.Run(strconv.Quote(c.src), func(t *testing.T) {
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
				{token.Indent, " "},
				{token.LINEFEED, "\n"},
				{token.Text, "a"},
			},
		},
		{
			" \n a",
			[]tl{
				{token.Indent, " "},
				{token.LINEFEED, "\n"},
				{token.Text, "a"},
			},
		},
		{
			"\t\n\ta",
			[]tl{
				{token.Indent, "\t"},
				{token.LINEFEED, "\n"},
				{token.Text, "a"},
			},
		},
		{
			"\t \n\t a",
			[]tl{
				{token.Indent, "\t "},
				{token.LINEFEED, "\n"},
				{token.Text, "a"},
			},
		},
		{
			" \t\n \ta",
			[]tl{
				{token.Indent, " \t"},
				{token.LINEFEED, "\n"},
				{token.Text, "a"},
			},
		},
		{
			"\t\n        a",
			[]tl{
				{token.Indent, "\t"},
				{token.LINEFEED, "\n"},
				{token.Text, "a"},
			},
		},
		{
			"        \n\ta",
			[]tl{
				{token.Indent, "        "},
				{token.LINEFEED, "\n"},
				{token.Text, "a"},
			},
		},
		{
			"\t\t\n\t\ta",
			[]tl{
				{token.Indent, "\t\t"},
				{token.LINEFEED, "\n"},
				{token.Text, "a"},
			},
		},
		{
			"\t\t\n\t        a",
			[]tl{
				{token.Indent, "\t\t"},
				{token.LINEFEED, "\n"},
				{token.Text, "a"},
			},
		},
		{
			"\t\t\n                a",
			[]tl{
				{token.Indent, "\t\t"},
				{token.LINEFEED, "\n"},
				{token.Text, "a"},
			},
		},
	}

	for _, c := range cases {
		t.Run(strconv.Quote(c.src), func(t *testing.T) {
			test(t, c.src, c.tokenPairs)
		})
	}
}
*/

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
