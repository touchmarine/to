package template_test

import (
	"fmt"
	"testing"

	"github.com/touchmarine/to/template"
)

func TestDot(t *testing.T) {
	m := map[string]interface{}{
		"a": "a",
		"b": map[string]interface{}{
			"bb": map[string]interface{}{
				"bbb": "bbb",
			},
		},
		"c": map[int]interface{}{
			0: "0",
		},
	}

	cases := []struct {
		in  string
		out interface{}
	}{
		{"a", m["a"]},
		{"a.a", ""},
		{"b", m["b"]},
		{"b.b", ""},
		{"b.bb", m["b"].(map[string]interface{})["bb"]},
		{"b.bb.bbb", m["b"].(map[string]interface{})["bb"].(map[string]interface{})["bbb"]},
		{"c", m["c"]},
		{"c.0", ""},
		{"c.a", ""},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			got := fmt.Sprintf("%q", template.Dot(m, c.in))
			want := fmt.Sprintf("%q", c.out)
			if got != want {
				t.Errorf("got %s, want %s", got, want)
			}
		})
	}
}
