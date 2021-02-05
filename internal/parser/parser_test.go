package parser_test

import (
	"strings"
	"testing"
	"to/internal/node"
	"to/internal/parser"
	"to/internal/stringifier"
	"unicode"
)

func TestLine(t *testing.T) {
	cases := []struct {
		in  string
		out []node.Node
	}{
		{
			"a",
			[]node.Node{node.Text("a")},
		},
	}

	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			nodes, errs := parser.Parse(strings.NewReader(c.in))
			for _, err := range errs {
				t.Errorf("got error %q", err)
			}

			got, want := stringifier.Stringify(nodes...), stringifier.Stringify(c.out...)
			if got != want {
				t.Errorf("\ngot\n%s\nwant\n%s", got, want)
			}
		})
	}
}

func TestInvalidUTF8Encoding(t *testing.T) {
	const fcb = "\x80" // first continuation byte

	cases := []struct {
		name string
		in   string
		out  []node.Node
	}{
		{
			"fcb at the beginning",
			fcb + "a",
			[]node.Node{node.Text(string(unicode.ReplacementChar) + "a")},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			nodes, errs := parser.Parse(strings.NewReader(c.in))

			var errCount uint
			for _, err := range errs {
				if err == parser.ErrInvalidUTF8Encoding {
					errCount++
					continue
				}
				t.Errorf("got error %q", err)
			}
			if errCount > 1 {
				t.Errorf(
					"got %d %s errors, want 1",
					errCount,
					parser.ErrInvalidUTF8Encoding,
				)
			}

			got, want := stringifier.Stringify(nodes...), stringifier.Stringify(c.out...)
			if got != want {
				t.Errorf("\ngot\n%s\nwant\n%s", got, want)
			}
		})
	}
}
