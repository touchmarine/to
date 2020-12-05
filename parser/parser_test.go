package parser_test

import (
	"testing"
	"to/node"
	"to/parser"
)

func TestParseDocument(t *testing.T) {
	cases := []struct {
		name  string
		input string
		doc   *node.Document
	}{
		{
			name:  "emphasis",
			input: "Touch Markup is a _simple_ markup language.",
			doc: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Children: []node.Inline{
							&node.Text{
								Value: "Touch Markup is a ",
							},
							&node.Emphasis{
								Children: []node.Inline{
									&node.Text{
										Value: "simple",
									},
								},
							},
							&node.Text{
								Value: " markup language.",
							},
						},
					},
				},
			},
		},
		{
			name:  "strong",
			input: "Touch Markup is a *simple* markup language.",
			doc: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Children: []node.Inline{
							&node.Text{
								Value: "Touch Markup is a ",
							},
							&node.Strong{
								Children: []node.Inline{
									&node.Text{
										Value: "simple",
									},
								},
							},
							&node.Text{
								Value: " markup language.",
							},
						},
					},
				},
			},
		},
		{
			name:  "unterminated nested emphasis",
			input: "Touch Markup is a *_simple* markup language.",
			doc: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Children: []node.Inline{
							&node.Text{
								Value: "Touch Markup is a ",
							},
							&node.Strong{
								Children: []node.Inline{
									&node.Emphasis{
										Children: []node.Inline{
											&node.Text{
												Value: "simple",
											},
										},
									},
								},
							},
							&node.Text{
								Value: " markup language.",
							},
						},
					},
				},
			},
		},
		{
			name:  "nested emphasis in strong",
			input: "Touch Markup is a *_simple_* markup language.",
			doc: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Children: []node.Inline{
							&node.Text{
								Value: "Touch Markup is a ",
							},
							&node.Strong{
								Children: []node.Inline{
									&node.Emphasis{
										Children: []node.Inline{
											&node.Text{
												Value: "simple",
											},
										},
									},
								},
							},
							&node.Text{
								Value: " markup language.",
							},
						},
					},
				},
			},
		},
		{
			name:  "double nested strong",
			input: "Touch Markup is a *_*simple*_* markup language.",
			doc: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Children: []node.Inline{
							&node.Text{
								Value: "Touch Markup is a ",
							},
							&node.Strong{
								Children: []node.Inline{
									&node.Emphasis{},
								},
							},
							&node.Text{
								Value: "simple",
							},
							&node.Strong{
								Children: []node.Inline{
									&node.Emphasis{},
								},
							},
							&node.Text{
								Value: " markup language.",
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := parser.New(tc.input)

			doc := p.ParseDocument()
			if doc == nil {
				t.Fatalf("ParseDocument() returned nil")
			}

			if doc.String() != tc.doc.String() {
				t.Errorf(
					"document \"%s\" is incorrect, from input `%s` got:\n%s\nwant:\n%s",
					tc.name,
					tc.input,
					doc.String(),
					tc.doc.String(),
				)
			}
		})
	}
}
