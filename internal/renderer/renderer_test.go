package renderer

import (
	"sort"
	"strings"
	"testing"
	"to/internal/node"
)

func TestOnlyLineComment(t *testing.T) {
	cases := []struct {
		name string
		in   []node.Inline
		out  bool
	}{
		{
			"empty",
			[]node.Inline{},
			false,
		},
		{
			"text",
			[]node.Inline{
				node.Text("a"),
			},
			false,
		},
		{
			"text and comment",
			[]node.Inline{
				node.Text("a"),
				node.LineComment("b"),
			},
			false,
		},
		{
			"comment",
			[]node.Inline{
				node.LineComment("b"),
			},
			true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := onlyLineComment(c.in); got != c.out {
				t.Errorf("got %t, want %t", got, c.out)
			}
		})
	}
}

func TestHead(t *testing.T) {
	cases := []struct {
		in  []string
		out string
	}{
		{
			[]string{},
			"",
		},
		{
			[]string{"a"},
			"a",
		},
		{
			[]string{"a", "b"},
			"a",
		},
	}

	for _, c := range cases {
		t.Run(strings.Join(c.in, " "), func(t *testing.T) {
			if got := head(c.in); got != c.out {
				t.Errorf("got %s, want %s", got, c.out)
			}
		})
	}
}

func TestBody(t *testing.T) {
	cases := []struct {
		in  []string
		out string
	}{
		{
			[]string{},
			"",
		},
		{
			[]string{"a"},
			"",
		},
		{
			[]string{"a", "b"},
			"b",
		},
		{
			[]string{"a", "b", "c"},
			"b\nc",
		},
	}

	for _, c := range cases {
		t.Run(strings.Join(c.in, " "), func(t *testing.T) {
			if got := body(c.in); got != c.out {
				t.Errorf("got %q, want %q", got, c.out)
			}
		})
	}
}

func TestPrimarySecondary(t *testing.T) {
	cases := []struct {
		in  []string
		out primarySecondary
	}{
		{
			[]string{""},
			primarySecondary{},
		},
		{
			[]string{"a"},
			primarySecondary{"a", ""},
		},
		{
			[]string{"a b"},
			primarySecondary{"a", "b"},
		},
		{
			[]string{"a\tb"},
			primarySecondary{"a", "b"},
		},
		{
			[]string{"a", "b"},
			primarySecondary{"a", "b"},
		},
		{
			[]string{"", "b"},
			primarySecondary{"", "b"},
		},
	}

	for _, c := range cases {
		t.Run(strings.Join(c.in, " "), func(t *testing.T) {
			got := parsePrimarySecondary(c.in)

			if got.Primary != c.out.Primary {
				t.Errorf("got Primary %q, want %q", got.Primary, c.out.Primary)
			}
			if got.Secondary != c.out.Secondary {
				t.Errorf("got Secondary %q, want %q", got.Secondary, c.out.Secondary)
			}
		})
	}
}

func TestNamedUnnamed(t *testing.T) {
	cases := []struct {
		in  []string
		out namedUnnamed
	}{
		{
			[]string{},
			namedUnnamed{},
		},

		// unnamed
		{
			[]string{"a"},
			namedUnnamed{[]string{"a"}, nil},
		},
		{
			[]string{"a;"},
			namedUnnamed{[]string{"a"}, nil},
		},
		{
			[]string{"a b;"},
			namedUnnamed{[]string{"a", "b"}, nil},
		},
		{
			[]string{" a b;"},
			namedUnnamed{[]string{"a", "b"}, nil},
		},
		{
			[]string{"a b ;"},
			namedUnnamed{[]string{"a", "b"}, nil},
		},
		{
			[]string{"a\tb;"},
			namedUnnamed{[]string{"a", "b"}, nil},
		},
		{
			[]string{"a", "b"},
			namedUnnamed{[]string{"a", "b"}, nil},
		},
		{
			[]string{"a ", " b"},
			namedUnnamed{[]string{"a", "b"}, nil},
		},

		// named
		{
			[]string{"a:"},
			namedUnnamed{nil, map[string]string{
				"a": "",
			}},
		},
		{
			[]string{"a b:"},
			namedUnnamed{nil, map[string]string{
				"a b": "",
			}},
		},
		{
			[]string{"a:b"},
			namedUnnamed{nil, map[string]string{
				"a": "b",
			}},
		},
		{
			[]string{"a:b c"},
			namedUnnamed{nil, map[string]string{
				"a": "b c",
			}},
		},

		// named spacing
		{
			[]string{" a:b"},
			namedUnnamed{nil, map[string]string{
				"a": "b",
			}},
		},
		{
			[]string{"a :b"},
			namedUnnamed{nil, map[string]string{
				"a": "b",
			}},
		},
		{
			[]string{"a: b"},
			namedUnnamed{nil, map[string]string{
				"a": "b",
			}},
		},
		{
			[]string{"a:b "},
			namedUnnamed{nil, map[string]string{
				"a": "b",
			}},
		},

		// multiple named
		{
			[]string{"a:b;c:d"},
			namedUnnamed{nil, map[string]string{
				"a": "b",
				"c": "d",
			}},
		},
		{
			[]string{"a:b", "c:d"},
			namedUnnamed{nil, map[string]string{
				"a": "b",
				"c": "d",
			}},
		},

		// redundant ";:
		{
			[]string{";a:b", "c:d"},
			namedUnnamed{nil, map[string]string{
				"a": "b",
				"c": "d",
			}},
		},
		{
			[]string{"a:b;", "c:d"},
			namedUnnamed{nil, map[string]string{
				"a": "b",
				"c": "d",
			}},
		},
		{
			[]string{"a:b", ";c:d"},
			namedUnnamed{nil, map[string]string{
				"a": "b",
				"c": "d",
			}},
		},
		{
			[]string{"a:b", "c:d;"},
			namedUnnamed{nil, map[string]string{
				"a": "b",
				"c": "d",
			}},
		},

		// duplcate keys
		{
			[]string{"a:b", "a:c"},
			namedUnnamed{nil, map[string]string{
				"a": "c",
			}},
		},
	}

	for _, c := range cases {
		t.Run(strings.Join(c.in, "\n"), func(t *testing.T) {
			got := parseNamedUnnamed(c.in)
			want := c.out

			if g, w := strings.Join(got.Unnamed, "; "), strings.Join(want.Unnamed, "; "); g != w {
				t.Errorf("got Unnamed %q, want %q", g, w)
			}

			if g, w := namedStr(got.Named), namedStr(want.Named); g != w {
				t.Errorf("got Named %q, want %q", g, w)
			}
		})
	}
}

func namedStr(m map[string]string) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	var i int
	for _, k := range keys {
		if i > 0 {
			b.WriteString("; ")
		}

		b.WriteString(k + ": " + m[k])
		i++
	}

	return b.String()
}
