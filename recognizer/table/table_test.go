package table_test

import (
	"github.com/touchmarine/to/config"
	"github.com/touchmarine/to/node"
	"github.com/touchmarine/to/recognizer/table"
	"strings"
	"testing"
)

func TestTable(t *testing.T) {
	cases := []struct {
		in  string
		out string
	}{
		// < = row  (hanging)
		// & = cell (hanging)

		{"a|b", "<&a\n &b"},
		{"a|b\n", "<&a\n &b\n"},
		{"a|b|c", "<&a\n &b\n &c"},

		{"a1|b1\na2|b2", "<&a1\n &b1\n<&a2\n &b2"},

		{"**a|b", "<&**a\n &b"},
		{"**a**|b", "<&**a**\n &b"},
		{"**a|**b", "<&**a\n &**b"},
		{"a|**b", "<&a\n &**b"},

		// escaped
		{"``a|b", "``a|b"},
		{"a``|b", "a``|b"},
		{"a|``b", "<&a\n &``b"},
	}

	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			var b strings.Builder

			table.Recognize(&b, []byte(c.in), []config.Element{
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
			})

			out := b.String()

			if out != c.out {
				t.Errorf("got %q, want %q", out, c.out)
			}
		})
	}
}
