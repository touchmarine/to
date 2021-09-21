package stringifier_test

import (
	"github.com/touchmarine/to/node"
	"github.com/touchmarine/to/stringifier"
	"testing"
)

func TestStringify(t *testing.T) {
	cases := []struct {
		name string
		in   *node.Node
		out  string
	}{
		{
			"a",
			&node.Node{
				Name: "MT",
				Type: node.TypeText,
				Data: "a",
			},
			`Text(MT)(
	a
)`,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := stringifier.Stringify(c.in)
			if got != c.out {
				t.Errorf("got %q, want %q", got, c.out)
			}
		})
	}
}
