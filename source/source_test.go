package source_test

import (
	"strings"
	"testing"

	"github.com/touchmarine/to/node"
	"github.com/touchmarine/to/parser"
	"github.com/touchmarine/to/source"
)

func TestNodeRanges(t *testing.T) {
	type rang struct {
		start, end int
		ranges     []source.Range
	}
	cases := []struct {
		in     string
		ranges []rang
	}{
		{
			"a\n>\n>+",
			[]rang{
				{
					start: 0,
					end:   1,
					ranges: []source.Range{
						{
							Start: source.Position{
								Offset: 0,
								Line:   0,
								Column: 0,
							},
							End: source.Position{
								Offset: 1,
								Line:   0,
								Column: 1,
							},
						},
					},
				},
				{
					start: 2,
					end:   6,
					ranges: []source.Range{
						{
							Start: source.Position{
								Offset: 2,
								Line:   1,
								Column: 0,
							},
							End: source.Position{
								Offset: 3,
								Line:   1,
								Column: 1,
							},
						},
						{
							Start: source.Position{
								Offset: 4,
								Line:   2,
								Column: 0,
							},
							End: source.Position{
								Offset: 6,
								Line:   2,
								Column: 2,
							},
						},
					},
				},
				{
					start: 4,
					end:   6,
					ranges: []source.Range{
						{
							Start: source.Position{
								Offset: 4,
								Line:   2,
								Column: 0,
							},
							End: source.Position{
								Offset: 6,
								Line:   2,
								Column: 2,
							},
						},
					},
				},
			},
		},
		{
			">\n >",
			[]rang{
				{
					start: 0,
					end:   4,
					ranges: []source.Range{
						{
							Start: source.Position{
								Offset: 0,
								Line:   0,
								Column: 0,
							},
							End: source.Position{
								Offset: 1,
								Line:   0,
								Column: 1,
							},
						},
						{
							Start: source.Position{
								Offset: 3,
								Line:   1,
								Column: 1,
							},
							End: source.Position{
								Offset: 4,
								Line:   1,
								Column: 2,
							},
						},
					},
				},
			},
		},
		{
			">**\n>**",
			[]rang{
				{
					start: 0,
					end:   7,
					ranges: []source.Range{
						{
							Start: source.Position{
								Offset: 0,
								Line:   0,
								Column: 0,
							},
							End: source.Position{
								Offset: 3,
								Line:   0,
								Column: 3,
							},
						},
						{
							Start: source.Position{
								Offset: 4,
								Line:   1,
								Column: 0,
							},
							End: source.Position{
								Offset: 7,
								Line:   1,
								Column: 3,
							},
						},
					},
				},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			in := []byte(c.in)
			sm := source.NewMap(in)
			p := parser.Parser{
				Elements: parser.Elements{
					"A": {
						Type:      node.TypeWalled,
						Delimiter: ">",
					},
					"B": {
						Type:      node.TypeWalled,
						Delimiter: "+",
					},
					"MA": {
						Type:      node.TypeUniform,
						Delimiter: "*",
					},
				},
				TabWidth: 8,
			}
			_, err := p.Parse(sm, in)
			if err != nil {
				t.Fatal(err)
			}
			for _, r := range c.ranges {
				got := rangesString(sm.NodeRanges(r.start, r.end))
				want := rangesString(r.ranges)
				if got != want {
					t.Errorf(
						"\nfor input (%d-%d):\n%s\ngot:\n%s\nwant:\n%s",
						r.start, r.end, tab(c.in), tab(got), tab(want),
					)
				}
			}
		})
	}
}

func rangesString(ranges []source.Range) string {
	var b strings.Builder
	for i, r := range ranges {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(r.String())
	}
	return b.String()
}

func tab(s string) string {
	var o string
	if s != "" && s[0] != '\n' {
		o += "\t"
	}
	o += strings.Replace(s, "\n", "\n\t", -1)
	return o
}
