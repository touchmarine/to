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
			name:  "underscore",
			input: "Tibsey is a _koala_.",
			doc: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Children: []node.Inline{
							&node.Text{
								Value: "Tibsey is a _koala_.",
							},
						},
					},
				},
			},
		},
		{
			name:  "emphasis",
			input: "Tibsey is a __koala__.",
			doc: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Children: []node.Inline{
							&node.Text{
								Value: "Tibsey is a ",
							},
							&node.Emphasis{
								Children: []node.Inline{
									&node.Text{
										Value: "koala",
									},
								},
							},
							&node.Text{
								Value: ".",
							},
						},
					},
				},
			},
		},
		{
			name:  "asterisk",
			input: "Climb *faster* Tibsey.",
			doc: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Children: []node.Inline{
							&node.Text{
								Value: "Climb *faster* Tibsey.",
							},
						},
					},
				},
			},
		},
		{
			name:  "strong",
			input: "Climb **faster** Tibsey.",
			doc: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Children: []node.Inline{
							&node.Text{
								Value: "Climb ",
							},
							&node.Strong{
								Children: []node.Inline{
									&node.Text{
										Value: "faster",
									},
								},
							},
							&node.Text{
								Value: " Tibsey.",
							},
						},
					},
				},
			},
		},
		{
			name:  "unterminated emphasis",
			input: "Tibsey is a __koala.",
			doc: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Children: []node.Inline{
							&node.Text{
								Value: "Tibsey is a ",
							},
							&node.Emphasis{
								Children: []node.Inline{
									&node.Text{
										Value: "koala.",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "unterminated emphasis in strong",
			input: "Tibsey is a **__koala**.",
			doc: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Children: []node.Inline{
							&node.Text{
								Value: "Tibsey is a ",
							},
							&node.Strong{
								Children: []node.Inline{
									&node.Emphasis{
										Children: []node.Inline{
											&node.Text{
												Value: "koala",
											},
										},
									},
								},
							},
							&node.Text{
								Value: ".",
							},
						},
					},
				},
			},
		},
		{
			name:  "nested emphasis in strong",
			input: "YEAH **__YEAH__** YEAH",
			doc: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Children: []node.Inline{
							&node.Text{
								Value: "YEAH ",
							},
							&node.Strong{
								Children: []node.Inline{
									&node.Emphasis{
										Children: []node.Inline{
											&node.Text{
												Value: "YEAH",
											},
										},
									},
								},
							},
							&node.Text{
								Value: " YEAH",
							},
						},
					},
				},
			},
		},
		{
			name:  "underscore in emphasis",
			input: "A __under_score__ inside emphasis.",
			doc: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Children: []node.Inline{
							&node.Text{
								Value: "A ",
							},
							&node.Emphasis{
								Children: []node.Inline{
									&node.Text{
										Value: "under_score",
									},
								},
							},
							&node.Text{
								Value: " inside emphasis.",
							},
						},
					},
				},
			},
		},
		{
			name:  "underscore in nested emphasis",
			input: "__Printer goes __brr_r__.__",
			doc: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Children: []node.Inline{
							&node.Emphasis{
								Children: []node.Inline{
									&node.Text{
										Value: "Printer goes ",
									},
								},
							},
							&node.Text{
								Value: "brr_r",
							},
							&node.Emphasis{
								Children: []node.Inline{
									&node.Text{
										Value: ".",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "intraword emphasis",
			input: "s__E__pt__E__mb__E__r",
			doc: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Children: []node.Inline{
							&node.Text{
								Value: "s",
							},
							&node.Emphasis{
								Children: []node.Inline{
									&node.Text{
										Value: "E",
									},
								},
							},
							&node.Text{
								Value: "pt",
							},
							&node.Emphasis{
								Children: []node.Inline{
									&node.Text{
										Value: "E",
									},
								},
							},
							&node.Text{
								Value: "mb",
							},
							&node.Emphasis{
								Children: []node.Inline{
									&node.Text{
										Value: "E",
									},
								},
							},
							&node.Text{
								Value: "r",
							},
						},
					},
				},
			},
		},
		{
			name: "not paragraphs",
			input: `
Tibsey is eating eucalyptus leaves.
Tibsey is going shopping.
Tibsey likes to sleep.
`,
			doc: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Children: []node.Inline{
							&node.Text{
								Value: "Tibsey is eating eucalyptus leaves.",
							},
							&node.Text{
								Value: "Tibsey is going shopping.",
							},
							&node.Text{
								Value: "Tibsey likes to sleep.",
							},
						},
					},
				},
			},
		},
		{
			name: "not paragraphs with strong",
			input: `
**Tibsey is eating eucalyptus leaves.
Tibsey is going shopping.**
Tibsey **likes** to sleep.
`,
			doc: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Children: []node.Inline{
							&node.Strong{
								Children: []node.Inline{
									&node.Text{
										Value: "Tibsey is eating eucalyptus leaves.",
									},
								},
							},
							&node.Text{
								Value: "Tibsey is going shopping.",
							},
							&node.Strong{},
							&node.Text{
								Value: "Tibsey ",
							},
							&node.Strong{
								Children: []node.Inline{
									&node.Text{
										Value: "likes",
									},
								},
							},
							&node.Text{
								Value: " to sleep.",
							},
						},
					},
				},
			},
		},
		{
			name: "paragraphs",
			input: `
Tibsey is eating eucalyptus leaves.

Tibsey is going shopping.

Tibsey likes to sleep.
`,
			doc: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Children: []node.Inline{
							&node.Text{
								Value: "Tibsey is eating eucalyptus leaves.",
							},
						},
					},
					&node.Paragraph{
						Children: []node.Inline{
							&node.Text{
								Value: "Tibsey is going shopping.",
							},
						},
					},
					&node.Paragraph{
						Children: []node.Inline{
							&node.Text{
								Value: "Tibsey likes to sleep.",
							},
						},
					},
				},
			},
		},
		{
			name: "paragraphs with strong",
			input: `
**Tibsey is eating eucalyptus leaves.

Tibsey is going shopping.**

Tibsey **likes** to sleep.
`,
			doc: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Children: []node.Inline{
							&node.Strong{
								Children: []node.Inline{
									&node.Text{
										Value: "Tibsey is eating eucalyptus leaves.",
									},
								},
							},
						},
					},
					&node.Paragraph{
						Children: []node.Inline{
							&node.Text{
								Value: "Tibsey is going shopping.",
							},
							&node.Strong{},
						},
					},
					&node.Paragraph{
						Children: []node.Inline{
							&node.Text{
								Value: "Tibsey ",
							},
							&node.Strong{
								Children: []node.Inline{
									&node.Text{
										Value: "likes",
									},
								},
							},
							&node.Text{
								Value: " to sleep.",
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
