package sequentialnumber

import (
	"encoding/json"
	"testing"
)

func TestGroup(t *testing.T) {
	cases := []struct {
		name string
		in   aggregate
		out  group
	}{
		{
			"single",
			aggregate{
				{
					Element:          "A",
					SequentialNumber: "1",
				},
			},
			group{
				particle{
					Element:          "A",
					SequentialNumber: "1",
				},
			},
		},
		{
			"1 depth",
			aggregate{
				{
					Element:          "A",
					SequentialNumber: "1",
				},
				{
					Element:          "A",
					SequentialNumber: "2",
				},
			},
			group{
				particle{
					Element:          "A",
					SequentialNumber: "1",
				},
				particle{
					Element:          "A",
					SequentialNumber: "2",
				},
			},
		},
		{
			"2 depths",
			aggregate{
				{
					Element:          "A",
					SequentialNumber: "1",
				},
				{
					Element:          "A",
					SequentialNumber: "1.1",
				},
			},
			group{
				particle{
					Element:          "A",
					SequentialNumber: "1",
				},
				group{
					particle{
						Element:          "A",
						SequentialNumber: "1.1",
					},
				},
			},
		},
		{
			"3 depths",
			aggregate{
				{
					Element:          "A",
					SequentialNumber: "1",
				},
				{
					Element:          "A",
					SequentialNumber: "1.1",
				},
				{
					Element:          "A",
					SequentialNumber: "1.1.1",
				},
			},
			group{
				particle{
					Element:          "A",
					SequentialNumber: "1",
				},
				group{
					particle{
						Element:          "A",
						SequentialNumber: "1.1",
					},
					group{
						particle{
							Element:          "A",
							SequentialNumber: "1.1.1",
						},
					},
				},
			},
		},
		{
			"decrease depth",
			aggregate{
				{
					Element:          "A",
					SequentialNumber: "1.1",
				},
				{
					Element:          "A",
					SequentialNumber: "1",
				},
			},
			group{
				group{
					particle{
						Element:          "A",
						SequentialNumber: "1.1",
					},
				},
				particle{
					Element:          "A",
					SequentialNumber: "1",
				},
			},
		},
		{
			"decrease depth 1",
			aggregate{
				{
					Element:          "A",
					SequentialNumber: "1.1",
				},
				{
					Element:          "A",
					SequentialNumber: "1",
				},
				{
					Element:          "A",
					SequentialNumber: "2",
				},
			},
			group{
				group{
					particle{
						Element:          "A",
						SequentialNumber: "1.1",
					},
				},
				particle{
					Element:          "A",
					SequentialNumber: "1",
				},
				particle{
					Element:          "A",
					SequentialNumber: "2",
				},
			},
		},
		{
			"decrease depth 2",
			aggregate{
				{
					Element:          "A",
					SequentialNumber: "1.1",
				},
				{
					Element:          "A",
					SequentialNumber: "1",
				},
				{
					Element:          "A",
					SequentialNumber: "2.1",
				},
			},
			group{
				group{
					particle{
						Element:          "A",
						SequentialNumber: "1.1",
					},
				},
				particle{
					Element:          "A",
					SequentialNumber: "1",
				},
				group{
					particle{
						Element:          "A",
						SequentialNumber: "2.1",
					},
				},
			},
		},
		{
			"increase depth",
			aggregate{
				{
					Element:          "A",
					SequentialNumber: "1",
				},
				{
					Element:          "A",
					SequentialNumber: "1.1",
				},
				{
					Element:          "A",
					SequentialNumber: "2",
				},
			},
			group{
				particle{
					Element:          "A",
					SequentialNumber: "1",
				},
				group{
					particle{
						Element:          "A",
						SequentialNumber: "1.1",
					},
				},
				particle{
					Element:          "A",
					SequentialNumber: "2",
				},
			},
		},
		{
			"increase depth",
			aggregate{
				{
					Element:          "A",
					SequentialNumber: "1",
				},
				{
					Element:          "A",
					SequentialNumber: "1.1",
				},
				{
					Element:          "A",
					SequentialNumber: "2",
				},
				{
					Element:          "A",
					SequentialNumber: "2.1",
				},
			},
			group{
				particle{
					Element:          "A",
					SequentialNumber: "1",
				},
				group{
					particle{
						Element:          "A",
						SequentialNumber: "1.1",
					},
				},
				particle{
					Element:          "A",
					SequentialNumber: "2",
				},
				group{
					particle{
						Element:          "A",
						SequentialNumber: "2.1",
					},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			gr := c.in.Group()
			got := indentedJSON(t, gr)
			want := indentedJSON(t, c.out)
			if got != want {
				t.Errorf("\ngot\n%s\nwant\n%s", got, want)
			}
		})
	}
}

func indentedJSON(t *testing.T, v interface{}) string {
	t.Helper()
	json, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		t.Fatal(err)
	}
	return string(json)
}
