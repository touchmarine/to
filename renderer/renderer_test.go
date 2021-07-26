package renderer

import (
	"encoding/json"
	"fmt"
	"github.com/touchmarine/to/aggregator"
	"github.com/touchmarine/to/node"
	"sort"
	"strings"
	"testing"
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

func TestGroupBySeqNum(t *testing.T) {
	cases := []struct {
		name string
		in   []aggregator.Item
		out  seqNumGroup
	}{
		{
			"1 depth",
			[]aggregator.Item{
				{
					Element: "NumberedHeading",
					SeqNums: []uint{1},
					SeqNum:  "1",
				},
				{
					Element: "NumberedHeading",
					SeqNums: []uint{2},
					SeqNum:  "2",
				},
			},
			seqNumGroup{
				seqNumItem{
					Element: "NumberedHeading",
					SeqNums: []uint{1},
					SeqNum:  "1",
				},
				seqNumItem{
					Element: "NumberedHeading",
					SeqNums: []uint{2},
					SeqNum:  "2",
				},
			},
		},
		{
			"2 depths",
			[]aggregator.Item{
				{
					Element: "NumberedHeading",
					SeqNums: []uint{1},
					SeqNum:  "1",
				},
				{
					Element: "NumberedHeading",
					SeqNums: []uint{1, 1},
					SeqNum:  "1.1",
				},
			},
			seqNumGroup{
				seqNumItem{
					Element: "NumberedHeading",
					SeqNums: []uint{1},
					SeqNum:  "1",
				},
				seqNumGroup{
					seqNumItem{
						Element: "NumberedHeading",
						SeqNums: []uint{1, 1},
						SeqNum:  "1.1",
					},
				},
			},
		},
		{
			"3 depths",
			[]aggregator.Item{
				{
					Element: "NumberedHeading",
					SeqNums: []uint{1},
					SeqNum:  "1",
				},
				{
					Element: "NumberedHeading",
					SeqNums: []uint{1, 1},
					SeqNum:  "1.1",
				},
				{
					Element: "NumberedHeading",
					SeqNums: []uint{1, 1, 1},
					SeqNum:  "1.1.1",
				},
			},
			seqNumGroup{
				seqNumItem{
					Element: "NumberedHeading",
					SeqNums: []uint{1},
					SeqNum:  "1",
				},
				seqNumGroup{
					seqNumItem{
						Element: "NumberedHeading",
						SeqNums: []uint{1, 1},
						SeqNum:  "1.1",
					},
					seqNumGroup{
						seqNumItem{
							Element: "NumberedHeading",
							SeqNums: []uint{1, 1, 1},
							SeqNum:  "1.1.1",
						},
					},
				},
			},
		},
		{
			"decrease depth",
			[]aggregator.Item{
				{
					Element: "NumberedHeading",
					SeqNums: []uint{1, 1},
					SeqNum:  "1.1",
				},
				{
					Element: "NumberedHeading",
					SeqNums: []uint{1},
					SeqNum:  "1",
				},
			},
			seqNumGroup{
				seqNumGroup{
					seqNumItem{
						Element: "NumberedHeading",
						SeqNums: []uint{1, 1},
						SeqNum:  "1.1",
					},
				},
				seqNumItem{
					Element: "NumberedHeading",
					SeqNums: []uint{1},
					SeqNum:  "1",
				},
			},
		},
		{
			"decrease depth 1",
			[]aggregator.Item{
				{
					Element: "NumberedHeading",
					SeqNums: []uint{1, 1},
					SeqNum:  "1.1",
				},
				{
					Element: "NumberedHeading",
					SeqNums: []uint{1},
					SeqNum:  "1",
				},
				{
					Element: "NumberedHeading",
					SeqNums: []uint{2},
					SeqNum:  "2",
				},
			},
			seqNumGroup{
				seqNumGroup{
					seqNumItem{
						Element: "NumberedHeading",
						SeqNums: []uint{1, 1},
						SeqNum:  "1.1",
					},
				},
				seqNumItem{
					Element: "NumberedHeading",
					SeqNums: []uint{1},
					SeqNum:  "1",
				},
				seqNumItem{
					Element: "NumberedHeading",
					SeqNums: []uint{2},
					SeqNum:  "2",
				},
			},
		},
		{
			"decrease depth 2",
			[]aggregator.Item{
				{
					Element: "NumberedHeading",
					SeqNums: []uint{1, 1},
					SeqNum:  "1.1",
				},
				{
					Element: "NumberedHeading",
					SeqNums: []uint{1},
					SeqNum:  "1",
				},
				{
					Element: "NumberedHeading",
					SeqNums: []uint{2, 1},
					SeqNum:  "2.1",
				},
			},
			seqNumGroup{
				seqNumGroup{
					seqNumItem{
						Element: "NumberedHeading",
						SeqNums: []uint{1, 1},
						SeqNum:  "1.1",
					},
				},
				seqNumItem{
					Element: "NumberedHeading",
					SeqNums: []uint{1},
					SeqNum:  "1",
				},
				seqNumGroup{
					seqNumItem{
						Element: "NumberedHeading",
						SeqNums: []uint{2, 1},
						SeqNum:  "2.1",
					},
				},
			},
		},
		{
			"decrease depth 3",
			[]aggregator.Item{
				{
					Element: "NumberedHeading",
					SeqNums: []uint{1},
					SeqNum:  "1",
				},
				{
					Element: "NumberedHeading",
					SeqNums: []uint{1, 1},
					SeqNum:  "1.1",
				},
				{
					Element: "NumberedHeading",
					SeqNums: []uint{2},
					SeqNum:  "2",
				},
			},
			seqNumGroup{
				seqNumItem{
					Element: "NumberedHeading",
					SeqNums: []uint{1},
					SeqNum:  "1",
				},
				seqNumGroup{
					seqNumItem{
						Element: "NumberedHeading",
						SeqNums: []uint{1, 1},
						SeqNum:  "1.1",
					},
				},
				seqNumItem{
					Element: "NumberedHeading",
					SeqNums: []uint{2},
					SeqNum:  "2",
				},
			},
		},
		{
			"decrease depth 4",
			[]aggregator.Item{
				{
					Element: "NumberedHeading",
					SeqNums: []uint{1},
					SeqNum:  "1",
				},
				{
					Element: "NumberedHeading",
					SeqNums: []uint{1, 1},
					SeqNum:  "1.1",
				},
				{
					Element: "NumberedHeading",
					SeqNums: []uint{2},
					SeqNum:  "2",
				},
				{
					Element: "NumberedHeading",
					SeqNums: []uint{2, 1},
					SeqNum:  "2.1",
				},
			},
			seqNumGroup{
				seqNumItem{
					Element: "NumberedHeading",
					SeqNums: []uint{1},
					SeqNum:  "1",
				},
				seqNumGroup{
					seqNumItem{
						Element: "NumberedHeading",
						SeqNums: []uint{1, 1},
						SeqNum:  "1.1",
					},
				},
				seqNumItem{
					Element: "NumberedHeading",
					SeqNums: []uint{2},
					SeqNum:  "2",
				},
				seqNumGroup{
					seqNumItem{
						Element: "NumberedHeading",
						SeqNums: []uint{2, 1},
						SeqNum:  "2.1",
					},
				},
			},
		},

		/*
			{
				"single",
				[]aggregator.Item{
					{
						Element: "NumberedHeading",
						SeqNums: []uint{1},
						SeqNum:  "1",
					},
				},
				[][]aggregator.Item{
					{
						{
							Element: "NumberedHeading",
							SeqNums: []uint{1},
							SeqNum:  "1",
						},
					},
				},
			},
			{
				"same depth",
				[]aggregator.Item{
					{
						Element: "NumberedHeading",
						SeqNums: []uint{1},
						SeqNum:  "1",
					},
					{
						Element: "NumberedHeading",
						SeqNums: []uint{2},
						SeqNum:  "2",
					},
				},
				[][]aggregator.Item{
					{
						{
							Element: "NumberedHeading",
							SeqNums: []uint{1},
							SeqNum:  "1",
						},
						{
							Element: "NumberedHeading",
							SeqNums: []uint{2},
							SeqNum:  "2",
						},
					},
				},
			},
			{
				"two depths",
				[]aggregator.Item{
					{
						Element: "NumberedHeading",
						SeqNums: []uint{1},
						SeqNum:  "1",
					},
					{
						Element: "NumberedHeading",
						SeqNums: []uint{1, 1},
						SeqNum:  "1.1",
					},
				},
				[][]aggregator.Item{
					{
						{
							Element: "NumberedHeading",
							SeqNums: []uint{1},
							SeqNum:  "1",
						},
					},
					{
						{
							Element: "NumberedHeading",
							SeqNums: []uint{1, 1},
							SeqNum:  "1.1",
						},
					},
				},
			},
			{
				"returning depth",
				[]aggregator.Item{
					{
						Element: "NumberedHeading",
						SeqNums: []uint{1, 1},
						SeqNum:  "1.1",
					},
					{
						Element: "NumberedHeading",
						SeqNums: []uint{1},
						SeqNum:  "1",
					},
				},
				[][]aggregator.Item{
					{
						{
							Element: "NumberedHeading",
							SeqNums: []uint{1, 1},
							SeqNum:  "1.1",
						},
					},
					{
						{
							Element: "NumberedHeading",
							SeqNums: []uint{1},
							SeqNum:  "1",
						},
					},
				},
			},
		*/
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			a := groupBySeqNum(c.in)

			got := jsonMarshal(t, a)
			want := jsonMarshal(t, c.out)

			if got != want {
				t.Errorf("\ngot\n%s\nwant\n%s", got, want)
			}
		})
	}
}

func TestParseAttr(t *testing.T) {
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
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("%q", c.in), func(t *testing.T) {
			var p attrParser
			p.init(strings.NewReader(c.in))
			p.parse()

			attrs := p.attrs

			got := fmt.Sprint(attrs)
			want := fmt.Sprint(c.out)

			if got != want {
				t.Errorf("got %s, want %s", got, want)
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

func jsonMarshal(t *testing.T, v interface{}) string {
	t.Helper()

	json, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		t.Fatal(err)
	}

	return string(json)
}
