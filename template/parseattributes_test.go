package template_test

import (
	"fmt"
	"testing"

	"github.com/touchmarine/to/template"
)

func TestParseAttributes(t *testing.T) {
	cases := []struct {
		in  string
		out map[string]string
	}{
		{
			"",
			nil,
		},
		{
			" ",
			nil,
		},
		{
			"=",
			nil,
		},

		{
			"a",
			map[string]string{"a": ""},
		},
		{
			"a=",
			map[string]string{"a": ""},
		},
		{
			" a=",
			map[string]string{"a": ""},
		},
		{
			"abc",
			map[string]string{"abc": ""},
		},
		{
			"-_",
			map[string]string{"-_": ""},
		},

		{
			"a=b",
			map[string]string{"a": "b"},
		},
		{
			"=b",
			nil,
		},

		{
			`a=b"`,
			map[string]string{"a": `b"`},
		},
		{
			"a=b'",
			map[string]string{"a": "b'"},
		},

		// spacing
		{
			" a",
			map[string]string{"a": ""},
		},
		{
			"a ",
			map[string]string{"a": ""},
		},

		// newline
		{
			"\na",
			map[string]string{"a": ""},
		},
		{
			"a\n",
			map[string]string{"a": ""},
		},
		{
			"a=\nb",
			map[string]string{"a": "", "b": ""},
		},
		{
			"a=b\n",
			map[string]string{"a": "b"},
		},
		{
			"a=\"\nb\"",
			map[string]string{"a": "\nb"},
		},
		{
			"a=\"b\n\"",
			map[string]string{"a": "b\n"},
		},
		{
			"a='\nb'",
			map[string]string{"a": "\nb"},
		},
		{
			"a='b\n'",
			map[string]string{"a": "b\n"},
		},

		// double quote
		{
			`a="b"`,
			map[string]string{"a": "b"},
		},
		{
			`a=" b "`,
			map[string]string{"a": " b "},
		},
		{
			`a="'"`,
			map[string]string{"a": "'"},
		},
		{
			`a="''"`,
			map[string]string{"a": "''"},
		},

		// escape
		{
			`a="\\"`,
			map[string]string{"a": `\`},
		},
		{
			`a="\""`,
			map[string]string{"a": `"`},
		},

		// single quote (raw content)
		{
			"a='b'",
			map[string]string{"a": "b"},
		},
		{
			"a=' b '",
			map[string]string{"a": " b "},
		},
		{
			`a='"'`,
			map[string]string{"a": `"`},
		},
		{
			`a='""'`,
			map[string]string{"a": `""`},
		},

		// no escape
		{
			`a='\\'`,
			map[string]string{"a": `\\`},
		},
		{
			`a='\"'`,
			map[string]string{"a": `\"`},
		},

		// multiple
		{
			"a b",
			map[string]string{"a": "", "b": ""},
		},
		{
			`a="1" b="2"`,
			map[string]string{"a": "1", "b": "2"},
		},
		{
			`a="1"b="2"`,
			map[string]string{"a": "1", "b": "2"},
		},
		{
			"a='1'b='2'",
			map[string]string{"a": "1", "b": "2"},
		},
		{
			"a=1b=2",
			map[string]string{"a": "1b=2"},
		},

		// duplicates
		{
			"a a",
			map[string]string{"a": ""},
		},
		{
			"a=a a=b",
			map[string]string{"a": "b"},
		},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("%q", c.in), func(t *testing.T) {
			got := fmt.Sprint(template.ParseAttributes(c.in))
			want := fmt.Sprint(c.out)
			if got != want {
				t.Errorf("got %s, want %s", got, want)
			}
		})
	}
}
