package autolink_test

import (
	"github.com/touchmarine/to/autolink"
	"testing"
)

func TestAutoLink(t *testing.T) {
	cases := []struct {
		in  string
		out string
	}{
		{"a", "a"},
		{"abc", "abc"},
		{"1", "1"},
		{"-", "-"}, // permitted domain punctuation

		// spacing
		{"a b", "a"},
		{"a\tb", "a"},
		{"a\nb", "a"},

		// domain punctuation (not permitted)
		{"?", ""},
		{"!", ""},
		{",", ""},
		{":", ""},
		{"*", ""},

		// domain underscores (not permitted in last two segments)
		{"_", ""},

		{"_._", ""},
		{"_.b", ""},
		{"a._", ""},
		{"a.b", "a.b"},

		{"_._._", ""},
		{"_.b.c", "_.b.c"},
		{"a._.c", ""},
		{"a.b._", ""},
		{"a._._", ""},
		{"a.b.c", "a.b.c"},

		// trailing puncutation (not permitted)
		{"m/.", "m/"},
		{"m/a.", "m/a"},
		{"m/a..", "m/a"},
		{"m/a.b.", "m/a.b"},
		{"m/a...", "m/a"},
		{"m/a.b.c.", "m/a.b.c"},

		// unmatched trailing parentheses (not permitted)
		{"m/()", "m/()"},
		{"m/(a", "m/(a"},
		{"m/a)", "m/a"},
		{"m/(a)", "m/(a)"},

		{"m/((a", "m/((a"},
		{"m/((a)", "m/((a)"},
		{"m/((a))", "m/((a))"},
		{"m/((a)))", "m/((a))"},

		{"m/a))", "m/a"},
		{"m/(a))", "m/(a)"},
		{"m/(((a))", "m/(((a))"},

		// unmatched non-trailing parentheses (permitted)
		{"m/()n", "m/()n"},
		{"m/(n", "m/(n"},
		{"m/)n", "m/)n"},

		{"m/(())n", "m/(())n"},
		{"m/(()n", "m/(()n"},
		{"m/())n", "m/())n"},
	}

	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			size := autolink.Match([]byte(c.in))

			got := c.in[:size]
			want := c.out

			if got != want {
				t.Errorf("got %q, want %q", got, want)
			}
		})
	}
}
