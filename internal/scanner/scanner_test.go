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
		{"\n", token.Newline, "\n"},
		{"// comment", token.Comment, "// comment"},

		{"|", token.BlockDelim, "|"},
		{">", token.BlockDelim, ">"},
		{"-", token.BlockDelim, "-"},

		{"__", token.InlineDelim, "__"},
		{"**", token.InlineDelim, "**"},
		{"``", token.InlineDelim, "``"},
		{"`*", token.InlineDelim, "`*"},
		{"`_", token.InlineDelim, "`_"},
		{"<<", token.InlineDelim, "<<"},
		{"<`", token.InlineDelim, "<`"},
		{"<*", token.InlineDelim, "<*"},
		{"<_", token.InlineDelim, "<_"},

		/*
			{">>", token.InlineDelim, ">>"},
			{"`>", token.InlineDelim, "`>"},
			{"*>", token.InlineDelim, "*>"},
			{"_>", token.InlineDelim, "_>"},
		*/
	}

	for _, c := range cases {
		t.Run(strconv.Quote(c.src), func(t *testing.T) {
			test(t, c.src, []tl{{c.tok, c.lit}})
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
					{token.Newline, "\n"},
					{token.Indent, "  "},
					{token.Text, "b"},
				},
			},
			{
				"  a\n   b",
				[]tl{
					{token.Indent, "  "},
					{token.Text, "a"},
					{token.Newline, "\n"},
					{token.Indent, "   "},
					{token.Text, "b"},
				},
			},
			{
				"        a\n         b",
				[]tl{
					{token.Indent, "        "},
					{token.Text, "a"},
					{token.Newline, "\n"},
					{token.Indent, "         "},
					{token.Text, "b"},
				},
			},
			{
				"\ta\n\t\tb",
				[]tl{
					{token.Indent, "\t"},
					{token.Text, "a"},
					{token.Newline, "\n"},
					{token.Indent, "\t\t"},
					{token.Text, "b"},
				},
			},
			{
				"\t\ta\n\t\t\tb",
				[]tl{
					{token.Indent, "\t\t"},
					{token.Text, "a"},
					{token.Newline, "\n"},
					{token.Indent, "\t\t\t"},
					{token.Text, "b"},
				},
			},
			{
				" \ta\n \t\tb",
				[]tl{
					{token.Indent, " \t"},
					{token.Text, "a"},
					{token.Newline, "\n"},
					{token.Indent, " \t\t"},
					{token.Text, "b"},
				},
			},
			{
				"\t a\n\t\t b",
				[]tl{
					{token.Indent, "\t "},
					{token.Text, "a"},
					{token.Newline, "\n"},
					{token.Indent, "\t\t "},
					{token.Text, "b"},
				},
			},
			{
				"\ta\n         b",
				[]tl{
					{token.Indent, "\t"},
					{token.Text, "a"},
					{token.Newline, "\n"},
					{token.Indent, "         "},
					{token.Text, "b"},
				},
			},
			{
				"       a\n\tb",
				[]tl{
					{token.Indent, "       "},
					{token.Text, "a"},
					{token.Newline, "\n"},
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
				{token.Newline, "\n"},
				{token.Dedent, ""},
				{token.Text, "a"},
			},
		},
		{
			" a\n  b\nc",
			[]tl{
				{token.Indent, " "},
				{token.Text, "a"},
				{token.Newline, "\n"},
				{token.Indent, "  "},
				{token.Text, "b"},
				{token.Newline, "\n"},
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
				{token.Newline, "\n"},
				{token.Indent, "  "},
				{token.Text, "b"},
				{token.Newline, "\n"},
				{token.Dedent, ""},
				{token.Text, "c"},
				{token.Newline, "\n"},
				{token.Dedent, ""},
				{token.Text, "d"},
			},
		},
		{
			"\t\na",
			[]tl{
				{token.Indent, "\t"},
				{token.Newline, "\n"},
				{token.Dedent, ""},
				{token.Text, "a"},
			},
		},
		{
			"\ta\n\t\tb\nc",
			[]tl{
				{token.Indent, "\t"},
				{token.Text, "a"},
				{token.Newline, "\n"},
				{token.Indent, "\t\t"},
				{token.Text, "b"},
				{token.Newline, "\n"},
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
				{token.Newline, "\n"},
				{token.Indent, "\t\t"},
				{token.Text, "b"},
				{token.Newline, "\n"},
				{token.Dedent, ""},
				{token.Text, "c"},
				{token.Newline, "\n"},
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
				{token.Newline, "\n"},
				{token.Indent, "        \t"},
				{token.Text, "b"},
				{token.Newline, "\n"},
				{token.Dedent, ""},
				{token.Text, "c"},
				{token.Newline, "\n"},
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
				{token.Newline, "\n"},
				{token.Text, "a"},
			},
		},
		{
			" \n a",
			[]tl{
				{token.Indent, " "},
				{token.Newline, "\n"},
				{token.Text, "a"},
			},
		},
		{
			"\t\n\ta",
			[]tl{
				{token.Indent, "\t"},
				{token.Newline, "\n"},
				{token.Text, "a"},
			},
		},
		{
			"\t \n\t a",
			[]tl{
				{token.Indent, "\t "},
				{token.Newline, "\n"},
				{token.Text, "a"},
			},
		},
		{
			" \t\n \ta",
			[]tl{
				{token.Indent, " \t"},
				{token.Newline, "\n"},
				{token.Text, "a"},
			},
		},
		{
			"\t\n        a",
			[]tl{
				{token.Indent, "\t"},
				{token.Newline, "\n"},
				{token.Text, "a"},
			},
		},
		{
			"        \n\ta",
			[]tl{
				{token.Indent, "        "},
				{token.Newline, "\n"},
				{token.Text, "a"},
			},
		},
		{
			"\t\t\n\t\ta",
			[]tl{
				{token.Indent, "\t\t"},
				{token.Newline, "\n"},
				{token.Text, "a"},
			},
		},
		{
			"\t\t\n\t        a",
			[]tl{
				{token.Indent, "\t\t"},
				{token.Newline, "\n"},
				{token.Text, "a"},
			},
		},
		{
			"\t\t\n                a",
			[]tl{
				{token.Indent, "\t\t"},
				{token.Newline, "\n"},
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
