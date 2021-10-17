package template

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/touchmarine/to/aggregator/seqnum"
)

func TestGroupBySequentialNumber(t *testing.T) {
	cases := []struct {
		name string
		in   seqnum.Aggregate
		out  sequentialNumberGroup
	}{
		{
			"single",
			seqnum.Aggregate{
				{
					Element:           "A",
					SequentialNumbers: []int{1},
				},
			},
			sequentialNumberGroup{
				sequentialNumberParticle{
					Element:           "A",
					SequentialNumbers: []int{1},
				},
			},
		},
		{
			"1 depth",
			seqnum.Aggregate{
				{
					Element:           "A",
					SequentialNumbers: []int{1},
				},
				{
					Element:           "A",
					SequentialNumbers: []int{2},
				},
			},
			sequentialNumberGroup{
				sequentialNumberParticle{
					Element:           "A",
					SequentialNumbers: []int{1},
				},
				sequentialNumberParticle{
					Element:           "A",
					SequentialNumbers: []int{2},
				},
			},
		},
		{
			"2 depths",
			seqnum.Aggregate{
				{
					Element:           "A",
					SequentialNumbers: []int{1},
				},
				{
					Element:           "A",
					SequentialNumbers: []int{1, 1},
				},
			},
			sequentialNumberGroup{
				sequentialNumberParticle{
					Element:           "A",
					SequentialNumbers: []int{1},
				},
				sequentialNumberGroup{
					sequentialNumberParticle{
						Element:           "A",
						SequentialNumbers: []int{1, 1},
					},
				},
			},
		},
		{
			"3 depths",
			seqnum.Aggregate{
				{
					Element:           "A",
					SequentialNumbers: []int{1},
				},
				{
					Element:           "A",
					SequentialNumbers: []int{1, 1},
				},
				{
					Element:           "A",
					SequentialNumbers: []int{1, 1, 1},
				},
			},
			sequentialNumberGroup{
				sequentialNumberParticle{
					Element:           "A",
					SequentialNumbers: []int{1},
				},
				sequentialNumberGroup{
					sequentialNumberParticle{
						Element:           "A",
						SequentialNumbers: []int{1, 1},
					},
					sequentialNumberGroup{
						sequentialNumberParticle{
							Element:           "A",
							SequentialNumbers: []int{1, 1, 1},
						},
					},
				},
			},
		},
		{
			"decrease depth",
			seqnum.Aggregate{
				{
					Element:           "A",
					SequentialNumbers: []int{1, 1},
				},
				{
					Element:           "A",
					SequentialNumbers: []int{1},
				},
			},
			sequentialNumberGroup{
				sequentialNumberGroup{
					sequentialNumberParticle{
						Element:           "A",
						SequentialNumbers: []int{1, 1},
					},
				},
				sequentialNumberParticle{
					Element:           "A",
					SequentialNumbers: []int{1},
				},
			},
		},
		{
			"decrease depth 1",
			seqnum.Aggregate{
				{
					Element:           "A",
					SequentialNumbers: []int{1, 1},
				},
				{
					Element:           "A",
					SequentialNumbers: []int{1},
				},
				{
					Element:           "A",
					SequentialNumbers: []int{2},
				},
			},
			sequentialNumberGroup{
				sequentialNumberGroup{
					sequentialNumberParticle{
						Element:           "A",
						SequentialNumbers: []int{1, 1},
					},
				},
				sequentialNumberParticle{
					Element:           "A",
					SequentialNumbers: []int{1},
				},
				sequentialNumberParticle{
					Element:           "A",
					SequentialNumbers: []int{2},
				},
			},
		},
		{
			"decrease depth 2",
			seqnum.Aggregate{
				{
					Element:           "A",
					SequentialNumbers: []int{1, 1},
				},
				{
					Element:           "A",
					SequentialNumbers: []int{1},
				},
				{
					Element:           "A",
					SequentialNumbers: []int{2, 1},
				},
			},
			sequentialNumberGroup{
				sequentialNumberGroup{
					sequentialNumberParticle{
						Element:           "A",
						SequentialNumbers: []int{1, 1},
					},
				},
				sequentialNumberParticle{
					Element:           "A",
					SequentialNumbers: []int{1},
				},
				sequentialNumberGroup{
					sequentialNumberParticle{
						Element:           "A",
						SequentialNumbers: []int{2, 1},
					},
				},
			},
		},
		{
			"increase depth",
			seqnum.Aggregate{
				{
					Element:           "A",
					SequentialNumbers: []int{1},
				},
				{
					Element:           "A",
					SequentialNumbers: []int{1, 1},
				},
				{
					Element:           "A",
					SequentialNumbers: []int{2},
				},
			},
			sequentialNumberGroup{
				sequentialNumberParticle{
					Element:           "A",
					SequentialNumbers: []int{1},
				},
				sequentialNumberGroup{
					sequentialNumberParticle{
						Element:           "A",
						SequentialNumbers: []int{1, 1},
					},
				},
				sequentialNumberParticle{
					Element:           "A",
					SequentialNumbers: []int{2},
				},
			},
		},
		{
			"increase depth",
			seqnum.Aggregate{
				{
					Element:           "A",
					SequentialNumbers: []int{1},
				},
				{
					Element:           "A",
					SequentialNumbers: []int{1, 1},
				},
				{
					Element:           "A",
					SequentialNumbers: []int{2},
				},
				{
					Element:           "A",
					SequentialNumbers: []int{2, 1},
				},
			},
			sequentialNumberGroup{
				sequentialNumberParticle{
					Element:           "A",
					SequentialNumbers: []int{1},
				},
				sequentialNumberGroup{
					sequentialNumberParticle{
						Element:           "A",
						SequentialNumbers: []int{1, 1},
					},
				},
				sequentialNumberParticle{
					Element:           "A",
					SequentialNumbers: []int{2},
				},
				sequentialNumberGroup{
					sequentialNumberParticle{
						Element:           "A",
						SequentialNumbers: []int{2, 1},
					},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			a := groupBySequentialNumber(c.in)

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

func jsonMarshal(t *testing.T, v interface{}) string {
	t.Helper()

	json, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		t.Fatal(err)
	}

	return string(json)
}
