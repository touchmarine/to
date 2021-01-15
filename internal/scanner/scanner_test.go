package scanner_test

import (
	"fmt"
	"os"
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
	}

	for _, c := range cases {
		t.Run(c.src, func(t *testing.T) {
			s := scanner.New(
				c.src,
				scanner.ScanComments,
				func(err error, errCount uint) {},
			)

			for {
				tok, lit := s.Scan()
				if tok == token.EOF {
					break
				}

				if tok != c.tok {
					t.Errorf("got token %s, want %s", tok, c.tok)
				}
				if lit != c.lit {
					t.Errorf("got literal %s, want %s", lit, c.lit)
				}
			}
		})
	}

}
