package parser_test

import (
	"errors"
	"testing"
	"to/node"
	"to/parser"
	"to/printer"
	"unicode/utf8"
)

func TestParseDocument(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  *node.Document
	}{
		{
			name:  "underscore",
			input: "Tibsey is a _koala_.",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Text{
										Value: "Tibsey is a _koala_.",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "emphasis",
			input: "Tibsey is a __koala__.",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
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
			},
		},
		{
			name:  "asterisk",
			input: "Climb *faster* Tibsey.",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Text{
										Value: "Climb *faster* Tibsey.",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "strong",
			input: "Climb **faster** Tibsey.",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
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
			},
		},
		{
			name:  "unterminated emphasis",
			input: "Tibsey is a __koala.",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
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
			},
		},
		{
			name:  "unterminated emphasis in strong",
			input: "Tibsey is a **__koala**.",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
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
			},
		},
		{
			name:  "nested emphasis in strong",
			input: "YEAH **__YEAH__** YEAH",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
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
			},
		},
		{
			name:  "underscore in emphasis",
			input: "A __under_score__ inside emphasis.",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
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
			},
		},
		{
			name:  "underscore in nested emphasis",
			input: "__Printer goes __brr_r__.__",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
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
			},
		},
		{
			name:  "intraword emphasis",
			input: "s__E__pt__E__mb__E__r",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
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
			},
		},
		{
			name: "not paragraphs",
			input: `
Tibsey is eating eucalyptus leaves.
Tibsey is going shopping.
Tibsey likes to sleep.
`,
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Text{
										Value: "Tibsey is eating eucalyptus leaves.",
									},
								},
							},
							{
								Children: []node.Inline{
									&node.Text{
										Value: "Tibsey is going shopping.",
									},
								},
							},
							{
								Children: []node.Inline{
									&node.Text{
										Value: "Tibsey likes to sleep.",
									},
								},
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
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
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
							{
								Children: []node.Inline{
									&node.Text{
										Value: "Tibsey is going shopping.",
									},
									&node.Strong{},
								},
							},
							{
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
			},
		},
		{
			name: "paragraphs",
			input: `
Tibsey is eating eucalyptus leaves.

Tibsey is going shopping.

Tibsey likes to sleep.
`,
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Text{
										Value: "Tibsey is eating eucalyptus leaves.",
									},
								},
							},
						},
					},
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Text{
										Value: "Tibsey is going shopping.",
									},
								},
							},
						},
					},
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Text{
										Value: "Tibsey likes to sleep.",
									},
								},
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
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
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
						},
					},
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Text{
										Value: "Tibsey is going shopping.",
									},
									&node.Strong{},
								},
							},
						},
					},
					&node.Paragraph{
						Lines: node.Lines{
							{
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
			},
		},
		{
			name: "paragraph lines",
			input: `
Tibsey is eating eucalyptus leaves.
Tibsey is going shopping.
Tibsey likes to sleep.
`,
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Text{
										Value: "Tibsey is eating eucalyptus leaves.",
									},
								},
							},
							{
								Children: []node.Inline{
									&node.Text{
										Value: "Tibsey is going shopping.",
									},
								},
							},
							{
								Children: []node.Inline{
									&node.Text{
										Value: "Tibsey likes to sleep.",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "heading 1",
			input: "= Koala",
			want: &node.Document{
				Children: []node.Node{
					&node.Heading{
						Level: 1,
						Children: []node.Inline{
							&node.Text{
								Value: "Koala",
							},
						},
					},
				},
			},
		},
		{
			name:  "heading 3",
			input: "=== Australia",
			want: &node.Document{
				Children: []node.Node{
					&node.Heading{
						Level: 3,
						Children: []node.Inline{
							&node.Text{
								Value: "Australia",
							},
						},
					},
				},
			},
		},
		{
			name:  "heading 30",
			input: "============================== Uh oh",
			want: &node.Document{
				Children: []node.Node{
					&node.Heading{
						Level: 30,
						Children: []node.Inline{
							&node.Text{
								Value: "Uh oh",
							},
						},
					},
				},
			},
		},
		{
			name:  "heading no space after =",
			input: "==Still a heading",
			want: &node.Document{
				Children: []node.Node{
					&node.Heading{
						Level: 2,
						Children: []node.Inline{
							&node.Text{
								Value: "Still a heading",
							},
						},
					},
				},
			},
		},
		{
			name:  "heading with sprinkled =",
			input: "== ======",
			want: &node.Document{
				Children: []node.Node{
					&node.Heading{
						Level: 2,
						Children: []node.Inline{
							&node.Text{
								Value: "======",
							},
						},
					},
				},
			},
		},
		{
			name: "consecutive headings",
			input: `
= Koalas
== Habitat
=== Australia
`,
			want: &node.Document{
				Children: []node.Node{
					&node.Heading{
						Level: 1,
						Children: []node.Inline{
							&node.Text{
								Value: "Koalas",
							},
						},
					},
					&node.Heading{
						Level: 2,
						Children: []node.Inline{
							&node.Text{
								Value: "Habitat",
							},
						},
					},
					&node.Heading{
						Level: 3,
						Children: []node.Inline{
							&node.Text{
								Value: "Australia",
							},
						},
					},
				},
			},
		},
		{
			name:  "heading emphasis and strong",
			input: "== __**Yee Haw**__",
			want: &node.Document{
				Children: []node.Node{
					&node.Heading{
						Level: 2,
						Children: []node.Inline{
							&node.Emphasis{
								Children: []node.Inline{
									&node.Strong{
										Children: []node.Inline{
											&node.Text{
												Value: "Yee Haw",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "numbered heading 1",
			input: "# Koala",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Text{
										Value: "# Koala",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "numbered heading 3",
			input: "### Australia",
			want: &node.Document{
				Children: []node.Node{
					&node.Heading{
						Level:      3,
						IsNumbered: true,
						Children: []node.Inline{
							&node.Text{
								Value: "Australia",
							},
						},
					},
				},
			},
		},
		{
			name:  "numbered heading 30",
			input: "############################## Uh oh",
			want: &node.Document{
				Children: []node.Node{
					&node.Heading{
						Level:      30,
						IsNumbered: true,
						Children: []node.Inline{
							&node.Text{
								Value: "Uh oh",
							},
						},
					},
				},
			},
		},
		{
			name:  "numbered heading no space after #",
			input: "##Still a heading",
			want: &node.Document{
				Children: []node.Node{
					&node.Heading{
						Level:      2,
						IsNumbered: true,
						Children: []node.Inline{
							&node.Text{
								Value: "Still a heading",
							},
						},
					},
				},
			},
		},
		{
			name:  "numbered heading with sprinkled #",
			input: "## ######",
			want: &node.Document{
				Children: []node.Node{
					&node.Heading{
						Level:      2,
						IsNumbered: true,
						Children: []node.Inline{
							&node.Text{
								Value: "######",
							},
						},
					},
				},
			},
		},
		{
			name: "consecutive numbered headings",
			input: `
= Koalas
## Habitat
### Australia
`,
			want: &node.Document{
				Children: []node.Node{
					&node.Heading{
						Level: 1,
						Children: []node.Inline{
							&node.Text{
								Value: "Koalas",
							},
						},
					},
					&node.Heading{
						Level:      2,
						IsNumbered: true,
						Children: []node.Inline{
							&node.Text{
								Value: "Habitat",
							},
						},
					},
					&node.Heading{
						Level:      3,
						IsNumbered: true,
						Children: []node.Inline{
							&node.Text{
								Value: "Australia",
							},
						},
					},
				},
			},
		},
		{
			name:  "numbered heading emphasis and strong",
			input: "## __**Yee Haw**__",
			want: &node.Document{
				Children: []node.Node{
					&node.Heading{
						Level:      2,
						IsNumbered: true,
						Children: []node.Inline{
							&node.Emphasis{
								Children: []node.Inline{
									&node.Strong{
										Children: []node.Inline{
											&node.Text{
												Value: "Yee Haw",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "heading and paragraph",
			input: `
= Koala

The koala is an iconic Australian animal. Often called...
`,
			want: &node.Document{
				Children: []node.Node{
					&node.Heading{
						Level: 1,
						Children: []node.Inline{
							&node.Text{
								Value: "Koala",
							},
						},
					},
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Text{
										Value: "The koala is an iconic Australian animal. Often called...",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "heading and paragraph no blank line",
			input: `
= Koala
The koala is an iconic Australian animal. Often called...
`,
			want: &node.Document{
				Children: []node.Node{
					&node.Heading{
						Level: 1,
						Children: []node.Inline{
							&node.Text{
								Value: "Koala",
							},
						},
					},
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Text{
										Value: "The koala is an iconic Australian animal. Often called...",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "numbered heading and paragraph",
			input: `
## Habitat

Koala lives in the...
`,
			want: &node.Document{
				Children: []node.Node{
					&node.Heading{
						Level:      2,
						IsNumbered: true,
						Children: []node.Inline{
							&node.Text{
								Value: "Habitat",
							},
						},
					},
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Text{
										Value: "Koala lives in the...",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "numbered heading and paragraph no blank line",
			input: `
## Habitat
Koala lives in the...
`,
			want: &node.Document{
				Children: []node.Node{
					&node.Heading{
						Level:      2,
						IsNumbered: true,
						Children: []node.Inline{
							&node.Text{
								Value: "Habitat",
							},
						},
					},
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Text{
										Value: "Koala lives in the...",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "paragraph and heading",
			input: `
The koala is an iconic Australian animal. Often called...

== Habitat
`,
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Text{
										Value: "The koala is an iconic Australian animal. Often called...",
									},
								},
							},
						},
					},
					&node.Heading{
						Level: 2,
						Children: []node.Inline{
							&node.Text{
								Value: "Habitat",
							},
						},
					},
				},
			},
		},
		{
			name: "paragraph and heading no blank line",
			input: `
The koala is an iconic Australian animal. Often called...
== Habitat
`,
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Text{
										Value: "The koala is an iconic Australian animal. Often called...",
									},
								},
							},
						},
					},
					&node.Heading{
						Level: 2,
						Children: []node.Inline{
							&node.Text{
								Value: "Habitat",
							},
						},
					},
				},
			},
		},
		{
			name: "paragraph and numbered heading",
			input: `
The koala is an iconic Australian animal. Often called...

## Habitat
`,
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Text{
										Value: "The koala is an iconic Australian animal. Often called...",
									},
								},
							},
						},
					},
					&node.Heading{
						Level:      2,
						IsNumbered: true,
						Children: []node.Inline{
							&node.Text{
								Value: "Habitat",
							},
						},
					},
				},
			},
		},
		{
			name: "paragraph and numbered heading no blank line",
			input: `
The koala is an iconic Australian animal. Often called...
## Habitat
`,
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Text{
										Value: "The koala is an iconic Australian animal. Often called...",
									},
								},
							},
						},
					},
					&node.Heading{
						Level:      2,
						IsNumbered: true,
						Children: []node.Inline{
							&node.Text{
								Value: "Habitat",
							},
						},
					},
				},
			},
		},
		{
			name:  "heading leading space",
			input: "=   \tKoala",
			want: &node.Document{
				Children: []node.Node{
					&node.Heading{
						Level: 1,
						Children: []node.Inline{
							&node.Text{
								Value: "Koala",
							},
						},
					},
				},
			},
		},
		{
			name:  "heading trailing space",
			input: "= Koala  \t",
			want: &node.Document{
				Children: []node.Node{
					&node.Heading{
						Level: 1,
						Children: []node.Inline{
							&node.Text{
								Value: "Koala  \t",
							},
						},
					},
				},
			},
		},
		{
			name:  "link",
			input: "<https://koala.test>",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Link{
										Destination: "https://koala.test",
										Children: []node.Inline{
											&node.Text{
												Value: "https://koala.test",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "nested link",
			input: "<https://</koala>.test>",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Link{
										Destination: "https://</koala",
										Children: []node.Inline{
											&node.Text{
												Value: "https://</koala",
											},
										},
									},
									&node.Text{
										Value: ".test>",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "two-part link nested in link",
			input: "<https://<koala></koala>.test>",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Link{
										Destination: "/koala",
										Children: []node.Inline{
											&node.Text{
												Value: "https://<koala",
											},
										},
									},
									&node.Text{
										Value: ".test>",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "alone >",
			input: "1 > 0",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Text{
										Value: "1 > 0",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "relative link",
			input: "</koalas>",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Link{
										Destination: "/koalas",
										Children: []node.Inline{
											&node.Text{
												Value: "/koalas",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "reference link",
			input: "<#habitat>",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Link{
										Destination: "#habitat",
										Children: []node.Inline{
											&node.Text{
												Value: "#habitat",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "email link",
			input: "<mailto:https://koala.test>",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Link{
										Destination: "mailto:https://koala.test",
										Children: []node.Inline{
											&node.Text{
												Value: "mailto:https://koala.test",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "link after unterminated strong",
			input: "**<https://koala.test> koalas are awesome",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Strong{
										Children: []node.Inline{
											&node.Link{
												Destination: "https://koala.test",
												Children: []node.Inline{
													&node.Text{
														Value: "https://koala.test",
													},
												},
											},
											&node.Text{
												Value: " koalas are awesome",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "link with strong",
			input: "Look at <https://**koala**.test> website",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Text{
										Value: "Look at ",
									},
									&node.Link{
										Destination: "https://**koala**.test",
										Children: []node.Inline{
											&node.Text{
												Value: "https://**koala**.test",
											},
										},
									},
									&node.Text{
										Value: " website",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "link with unterminated strong",
			input: "Look at <https://**koala.test> website",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Text{
										Value: "Look at ",
									},
									&node.Link{
										Destination: "https://**koala.test",
										Children: []node.Inline{
											&node.Text{
												Value: "https://**koala.test",
											},
										},
									},
									&node.Text{
										Value: " website",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "unterminated link",
			input: "Look at <https://koala.test website",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Text{
										Value: "Look at ",
									},
									&node.Link{
										Destination: "https://koala.test website",
										Children: []node.Inline{
											&node.Text{
												Value: "https://koala.test website",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "unterminated link another link",
			input: "Look at <https://koala.test website <#habitat>",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Text{
										Value: "Look at ",
									},
									&node.Link{
										Destination: "https://koala.test website <#habitat",
										Children: []node.Inline{
											&node.Text{
												Value: "https://koala.test website <#habitat",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "two-part link",
			input: "<Koala bears><https://koala.test>",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Link{
										Destination: "https://koala.test",
										Children: []node.Inline{
											&node.Text{
												Value: "Koala bears",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "link nested in link text",
			input: "<Koala </bears><https://koala.test>",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Link{
										Destination: "https://koala.test",
										Children: []node.Inline{
											&node.Text{
												Value: "Koala </bears",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "link nested in link destination",
			input: "<Koala><https://</koala>.test>",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Link{
										Destination: "https://</koala",
										Children: []node.Inline{
											&node.Text{
												Value: "Koala",
											},
										},
									},
									&node.Text{
										Value: ".test>",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "two-part link nested in link text",
			input: "<Koala <Bears></bears><https://koala.test>",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Link{
										Destination: "/bears",
										Children: []node.Inline{
											&node.Text{
												Value: "Koala <Bears",
											},
										},
									},
									&node.Link{
										Destination: "https://koala.test",
										Children: []node.Inline{
											&node.Text{
												Value: "https://koala.test",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "two-part link nested in link destination",
			input: "<Koala><https://<koala></koala>.test>",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Link{
										Destination: "https://<koala",
										Children: []node.Inline{
											&node.Text{
												Value: "Koala",
											},
										},
									},
									&node.Link{
										Destination: "/koala",
										Children: []node.Inline{
											&node.Text{
												Value: "/koala",
											},
										},
									},
									&node.Text{
										Value: ".test>",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "two-part link empahsized",
			input: "__<Koala bears><https://koala.test>__",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Emphasis{
										Children: []node.Inline{
											&node.Link{
												Destination: "https://koala.test",
												Children: []node.Inline{
													&node.Text{
														Value: "Koala bears",
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "two-part link strong text",
			input: "<**Koala bears**><https://koala.test>",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Link{
										Destination: "https://koala.test",
										Children: []node.Inline{
											&node.Strong{
												Children: []node.Inline{
													&node.Text{
														Value: "Koala bears",
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "two-part link unterminated strong text",
			input: "<**Koala bears><https://koala.test>",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Link{
										Destination: "https://koala.test",
										Children: []node.Inline{
											&node.Strong{
												Children: []node.Inline{
													&node.Text{
														Value: "Koala bears",
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "link with <",
			input: "<Koala bears<https://koala.test>",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Link{
										Destination: "Koala bears<https://koala.test",
										Children: []node.Inline{
											&node.Text{
												Value: "Koala bears<https://koala.test",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "link with < unterminated strong",
			input: "<**Koala bears<https://koala.test>",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Link{
										Destination: "**Koala bears<https://koala.test",
										Children: []node.Inline{
											&node.Text{
												Value: "**Koala bears<https://koala.test",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "link with < and strong",
			input: "<**Koala bears**<https://koala.test>",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Link{
										Destination: "**Koala bears**<https://koala.test",
										Children: []node.Inline{
											&node.Text{
												Value: "**Koala bears**<https://koala.test",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "two-part link strong and emphasis",
			input: "<**__Koala__ bears**><https://koala.test>",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Link{
										Destination: "https://koala.test",
										Children: []node.Inline{
											&node.Strong{
												Children: []node.Inline{
													&node.Emphasis{
														Children: []node.Inline{
															&node.Text{
																Value: "Koala",
															},
														},
													},
													&node.Text{
														Value: " bears",
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "two-part link plain destination",
			input: "<Koala bears><https://**koala**.test>",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Link{
										Destination: "https://**koala**.test",
										Children: []node.Inline{
											&node.Text{
												Value: "Koala bears",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "consecutive links",
			input: "<https://koala.test><https://eucalyptus.test>",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Link{
										Destination: "https://eucalyptus.test",
										Children: []node.Inline{
											&node.Text{
												Value: "https://koala.test",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "two links",
			input: "<https://koala.test> <https://eucalyptus.test>",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Link{
										Destination: "https://koala.test",
										Children: []node.Inline{
											&node.Text{
												Value: "https://koala.test",
											},
										},
									},
									&node.Text{
										Value: " ",
									},
									&node.Link{
										Destination: "https://eucalyptus.test",
										Children: []node.Inline{
											&node.Text{
												Value: "https://eucalyptus.test",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "consecutive two-part links",
			input: "<Koala bears><https://koala.test><Eucalyptus><https://eucalyptus.test>",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Link{
										Destination: "https://koala.test",
										Children: []node.Inline{
											&node.Text{
												Value: "Koala bears",
											},
										},
									},
									&node.Link{
										Destination: "https://eucalyptus.test",
										Children: []node.Inline{
											&node.Text{
												Value: "Eucalyptus",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "links in paragraphs",
			input: `<https://koala.test>
<https://eucalyptus.test>

<#habitat>`,
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Link{
										Destination: "https://koala.test",
										Children: []node.Inline{
											&node.Text{
												Value: "https://koala.test",
											},
										},
									},
								},
							},
							{
								Children: []node.Inline{
									&node.Link{
										Destination: "https://eucalyptus.test",
										Children: []node.Inline{
											&node.Text{
												Value: "https://eucalyptus.test",
											},
										},
									},
								},
							},
						},
					},
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Link{
										Destination: "#habitat",
										Children: []node.Inline{
											&node.Text{
												Value: "#habitat",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "code block",
			input: "``ts" + `
function displayButton(): void {
	const button = wantument.querySelector("button")
	button.style.display = "block"
	// ...
}
` + "``",
			want: &node.Document{
				Children: []node.Node{
					&node.CodeBlock{
						Language:    "ts",
						Filename:    "",
						MetadataRaw: "ts",
						Body: `function displayButton(): void {
	const button = wantument.querySelector("button")
	button.style.display = "block"
	// ...
}
`,
					},
				},
			},
		},
		{
			name: "code block no metadata",
			input: "``" + `
function displayButton(): void {
	const button = wantument.querySelector("button")
	button.style.display = "block"
	// ...
}
` + "``",
			want: &node.Document{
				Children: []node.Node{
					&node.CodeBlock{
						Language:    "",
						Filename:    "",
						MetadataRaw: "",
						Body: `function displayButton(): void {
	const button = wantument.querySelector("button")
	button.style.display = "block"
	// ...
}
`,
					},
				},
			},
		},
		{
			name: "code block with full metadata and whitespace",
			input: "``\tts  ,  button.ts  " + `
function displayButton(): void {
	const button = wantument.querySelector("button")
	button.style.display = "block"
	// ...
}
` + "``",
			want: &node.Document{
				Children: []node.Node{
					&node.CodeBlock{
						Language:    "ts",
						Filename:    "button.ts",
						MetadataRaw: "\tts  ,  button.ts  ",
						Body: `function displayButton(): void {
	const button = wantument.querySelector("button")
	button.style.display = "block"
	// ...
}
`,
					},
				},
			},
		},
		{
			name:  "code block no body",
			input: "``ts\n``",
			want: &node.Document{
				Children: []node.Node{
					&node.CodeBlock{
						Language:    "ts",
						Filename:    "",
						MetadataRaw: "ts",
						Body:        "",
					},
				},
			},
		},
		{
			name: "code block more than two delimiters",
			input: "````ts" + `
function displayButton(): void {
	const button = wantument.querySelector("button")
	button.style.display = "block"
	// ...
}
` + "````",
			want: &node.Document{
				Children: []node.Node{
					&node.CodeBlock{
						Language:    "ts",
						Filename:    "",
						MetadataRaw: "ts",
						Body: `function displayButton(): void {
	const button = wantument.querySelector("button")
	button.style.display = "block"
	// ...
}
`,
					},
				},
			},
		},
		{
			name: "code block more closing delimiters",
			input: "``ts" + `
function displayButton(): void {
	const button = wantument.querySelector("button")
	button.style.display = "block"
	// ...
}
` + "```",
			want: &node.Document{
				Children: []node.Node{
					&node.CodeBlock{
						Language:    "ts",
						Filename:    "",
						MetadataRaw: "ts",
						Body: `function displayButton(): void {
	const button = wantument.querySelector("button")
	button.style.display = "block"
	// ...
}
`,
					},
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Text{
										Value: "`",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "code block escaped delimiter inside body",
			input: "````ts\n```\n````",
			want: &node.Document{
				Children: []node.Node{
					&node.CodeBlock{
						Language:    "ts",
						Filename:    "",
						MetadataRaw: "ts",
						Body:        "```\n",
					},
				},
			},
		},
		{
			name:  "code block delimiter inside body",
			input: "``ts\n``\n``",
			want: &node.Document{
				Children: []node.Node{
					&node.CodeBlock{
						Language:    "ts",
						Filename:    "",
						MetadataRaw: "ts",
						Body:        "",
					},
					&node.CodeBlock{
						Language:    "",
						Filename:    "",
						MetadataRaw: "",
						Body:        "",
					},
				},
			},
		},
		{
			name:  "unterminated code block",
			input: "``ts\n",
			want: &node.Document{
				Children: []node.Node{
					&node.CodeBlock{
						Language:    "ts",
						Filename:    "",
						MetadataRaw: "ts",
						Body:        "",
					},
				},
			},
		},
		{
			name:  "code block inline",
			input: "``ts function displayButton(): void``",
			want: &node.Document{
				Children: []node.Node{
					&node.CodeBlock{
						Language:    "ts function displayButton(): void``",
						Filename:    "",
						MetadataRaw: "ts function displayButton(): void``",
						Body:        "",
					},
				},
			},
		},
		{
			name:  "code block two line",
			input: "``ts function displayButton(): void\n``",
			want: &node.Document{
				Children: []node.Node{
					&node.CodeBlock{
						Language:    "ts function displayButton(): void",
						Filename:    "",
						MetadataRaw: "ts function displayButton(): void",
						Body:        "",
					},
				},
			},
		},
		{
			name: "code block text on closing delimiter line",
			input: "``ts" + `
function displayButton(): void {
	const button = wantument.querySelector("button")
	button.style.display = "block"
	// ...
}
` + "`` ALREADY A PARAGRAPH",
			want: &node.Document{
				Children: []node.Node{
					&node.CodeBlock{
						Language:    "ts",
						Filename:    "",
						MetadataRaw: "ts",
						Body: `function displayButton(): void {
	const button = wantument.querySelector("button")
	button.style.display = "block"
	// ...
}
`,
					},
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Text{
										Value: "ALREADY A PARAGRAPH",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "code block after heading",
			input: "= Koala Language\n``koala\nEucalyptus, nom nom\n``",
			want: &node.Document{
				Children: []node.Node{
					&node.Heading{
						Level: 1,
						Children: []node.Inline{
							&node.Text{
								Value: "Koala Language",
							},
						},
					},
					&node.CodeBlock{
						Language:    "koala",
						MetadataRaw: "koala",
						Body:        "Eucalyptus, nom nom\n",
					},
				},
			},
		},
		{
			name:  "code block before heading",
			input: "``koala\nEucalyptus, nom nom\n``\n## Koala Language",
			want: &node.Document{
				Children: []node.Node{
					&node.CodeBlock{
						Language:    "koala",
						MetadataRaw: "koala",
						Body:        "Eucalyptus, nom nom\n",
					},
					&node.Heading{
						Level:      2,
						IsNumbered: true,
						Children: []node.Inline{
							&node.Text{
								Value: "Koala Language",
							},
						},
					},
				},
			},
		},
		{
			name:  "code block after paragraph",
			input: "Koala Language\n``koala\nEucalyptus, nom nom\n``",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Text{
										Value: "Koala Language",
									},
								},
							},
						},
					},
					&node.CodeBlock{
						Language:    "koala",
						MetadataRaw: "koala",
						Body:        "Eucalyptus, nom nom\n",
					},
				},
			},
		},
		{
			name:  "code block before paragraph",
			input: "``koala\nEucalyptus, nom nom\n``\nKoala Language",
			want: &node.Document{
				Children: []node.Node{
					&node.CodeBlock{
						Language:    "koala",
						MetadataRaw: "koala",
						Body:        "Eucalyptus, nom nom\n",
					},
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Text{
										Value: "Koala Language",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "unordered list",
			input: "- milk",
			want: &node.Document{
				Children: []node.Node{
					&node.List{
						Type: node.UnorderedList,
						ListItems: []*node.ListItem{
							{
								Children: []node.Node{
									node.Lines{
										&node.Line{
											Children: []node.Inline{
												&node.Text{
													Value: "milk",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "unordered list indented",
			input: `
 - milk`,
			want: &node.Document{
				Children: []node.Node{
					&node.List{
						Type: node.UnorderedList,
						ListItems: []*node.ListItem{
							{
								Children: []node.Node{
									node.Lines{
										&node.Line{
											Children: []node.Inline{
												&node.Text{
													Value: "milk",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "unordered list no whitespace after -",
			input: "-milk",
			want: &node.Document{
				Children: []node.Node{
					&node.List{
						Type: node.UnorderedList,
						ListItems: []*node.ListItem{
							{
								Children: []node.Node{
									node.Lines{
										&node.Line{
											Children: []node.Inline{
												&node.Text{
													Value: "milk",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "unordered list multiple items",
			input: `
- milk
- sugar
- bananas`,
			want: &node.Document{
				Children: []node.Node{
					&node.List{
						Type: node.UnorderedList,
						ListItems: []*node.ListItem{
							{
								Children: []node.Node{
									node.Lines{
										&node.Line{
											Children: []node.Inline{
												&node.Text{
													Value: "milk",
												},
											},
										},
									},
								},
							},
							{
								Children: []node.Node{
									node.Lines{
										&node.Line{
											Children: []node.Inline{
												&node.Text{
													Value: "sugar",
												},
											},
										},
									},
								},
							},
							{
								Children: []node.Node{
									node.Lines{
										&node.Line{
											Children: []node.Inline{
												&node.Text{
													Value: "bananas",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "unordered list multiple items - first indented",
			input: `
  - milk
  - sugar
  - bananas`,
			want: &node.Document{
				Children: []node.Node{
					&node.List{
						Type: node.UnorderedList,
						ListItems: []*node.ListItem{
							{
								Children: []node.Node{
									node.Lines{
										&node.Line{
											Children: []node.Inline{
												&node.Text{
													Value: "milk",
												},
											},
										},
									},
								},
							},
							{
								Children: []node.Node{
									node.Lines{
										&node.Line{
											Children: []node.Inline{
												&node.Text{
													Value: "sugar",
												},
											},
										},
									},
								},
							},
							{
								Children: []node.Node{
									node.Lines{
										&node.Line{
											Children: []node.Inline{
												&node.Text{
													Value: "bananas",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "unordered list nested",
			input: `
- Tuesday
 - milk`,
			want: &node.Document{
				Children: []node.Node{
					&node.List{
						Type: node.UnorderedList,
						ListItems: []*node.ListItem{
							{
								Children: []node.Node{
									node.Lines{
										&node.Line{
											Children: []node.Inline{
												&node.Text{
													Value: "Tuesday",
												},
											},
										},
									},
									&node.List{
										Type: node.UnorderedList,
										ListItems: []*node.ListItem{
											{
												Children: []node.Node{
													node.Lines{
														&node.Line{
															Children: []node.Inline{
																&node.Text{
																	Value: "milk",
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "unordered list tab nested",
			input: `
- Tuesday
	- milk`,
			want: &node.Document{
				Children: []node.Node{
					&node.List{
						Type: node.UnorderedList,
						ListItems: []*node.ListItem{
							{
								Children: []node.Node{
									node.Lines{
										&node.Line{
											Children: []node.Inline{
												&node.Text{
													Value: "Tuesday",
												},
											},
										},
									},
									&node.List{
										Type: node.UnorderedList,
										ListItems: []*node.ListItem{
											{
												Children: []node.Node{
													node.Lines{
														&node.Line{
															Children: []node.Inline{
																&node.Text{
																	Value: "milk",
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "unordered list - item with multiple children",
			input: `
- Tuesday
 Shopping list:
 - milk`,
			want: &node.Document{
				Children: []node.Node{
					&node.List{
						Type: node.UnorderedList,
						ListItems: []*node.ListItem{
							{
								Children: []node.Node{
									node.Lines{
										{
											Children: []node.Inline{
												&node.Text{
													Value: "Tuesday",
												},
											},
										},
										{
											Children: []node.Inline{
												&node.Text{
													Value: "Shopping list:",
												},
											},
										},
									},
									&node.List{
										Type: node.UnorderedList,
										ListItems: []*node.ListItem{
											{
												Children: []node.Node{
													node.Lines{
														&node.Line{
															Children: []node.Inline{
																&node.Text{
																	Value: "milk",
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "unordered list whitespace item nested",
			// whitespace after first '-'
			input: `
-
 - inner`,
			want: &node.Document{
				Children: []node.Node{
					&node.List{
						Type: node.UnorderedList,
						ListItems: []*node.ListItem{
							{
								Children: []node.Node{
									&node.List{
										Type: node.UnorderedList,
										ListItems: []*node.ListItem{
											{
												Children: []node.Node{
													node.Lines{
														&node.Line{
															Children: []node.Inline{
																&node.Text{
																	Value: "inner",
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "unordered list blank item nested",
			// no whitespace after first '-'
			input: `
-
 - inner`,
			want: &node.Document{
				Children: []node.Node{
					&node.List{
						Type: node.UnorderedList,
						ListItems: []*node.ListItem{
							{
								Children: []node.Node{
									&node.List{
										Type: node.UnorderedList,
										ListItems: []*node.ListItem{
											{
												Children: []node.Node{
													node.Lines{
														&node.Line{
															Children: []node.Inline{
																&node.Text{
																	Value: "inner",
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "unordered list deeply nested 1",
			input: `
- Tuesday
 - milk
  - low-fat`,
			want: &node.Document{
				Children: []node.Node{
					&node.List{
						Type: node.UnorderedList,
						ListItems: []*node.ListItem{
							{
								Children: []node.Node{
									node.Lines{
										&node.Line{
											Children: []node.Inline{
												&node.Text{
													Value: "Tuesday",
												},
											},
										},
									},
									&node.List{
										Type: node.UnorderedList,
										ListItems: []*node.ListItem{
											{
												Children: []node.Node{
													node.Lines{
														&node.Line{
															Children: []node.Inline{
																&node.Text{
																	Value: "milk",
																},
															},
														},
													},
													&node.List{
														Type: node.UnorderedList,
														ListItems: []*node.ListItem{
															{
																Children: []node.Node{
																	node.Lines{
																		&node.Line{
																			Children: []node.Inline{
																				&node.Text{
																					Value: "low-fat",
																				},
																			},
																		},
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "unordered list deeply nested 2",
			// list is indented
			input: `
 - Tuesday
  - milk
   - low-fat`,
			want: &node.Document{
				Children: []node.Node{
					&node.List{
						Type: node.UnorderedList,
						ListItems: []*node.ListItem{
							{
								Children: []node.Node{
									node.Lines{
										&node.Line{
											Children: []node.Inline{
												&node.Text{
													Value: "Tuesday",
												},
											},
										},
									},
									&node.List{
										Type: node.UnorderedList,
										ListItems: []*node.ListItem{
											{
												Children: []node.Node{
													node.Lines{
														&node.Line{
															Children: []node.Inline{
																&node.Text{
																	Value: "milk",
																},
															},
														},
													},
													&node.List{
														Type: node.UnorderedList,
														ListItems: []*node.ListItem{
															{
																Children: []node.Node{
																	node.Lines{
																		&node.Line{
																			Children: []node.Inline{
																				&node.Text{
																					Value: "low-fat",
																				},
																			},
																		},
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "unordered list deeply nested 3",
			input: `
- Tuesday
 - oatmeal
 - milk
  milk that is:
  - low-fat`,
			want: &node.Document{
				Children: []node.Node{
					&node.List{
						Type: node.UnorderedList,
						ListItems: []*node.ListItem{
							{
								Children: []node.Node{
									node.Lines{
										&node.Line{
											Children: []node.Inline{
												&node.Text{
													Value: "Tuesday",
												},
											},
										},
									},
									&node.List{
										Type: node.UnorderedList,
										ListItems: []*node.ListItem{
											{
												Children: []node.Node{
													node.Lines{
														&node.Line{
															Children: []node.Inline{
																&node.Text{
																	Value: "oatmeal",
																},
															},
														},
													},
												},
											},
											{
												Children: []node.Node{
													node.Lines{
														{
															Children: []node.Inline{
																&node.Text{
																	Value: "milk",
																},
															},
														},
														{
															Children: []node.Inline{
																&node.Text{
																	Value: "milk that is:",
																},
															},
														},
													},
													&node.List{
														Type: node.UnorderedList,
														ListItems: []*node.ListItem{
															{
																Children: []node.Node{
																	node.Lines{
																		&node.Line{
																			Children: []node.Inline{
																				&node.Text{
																					Value: "low-fat",
																				},
																			},
																		},
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "unordered list deeply nested 4",
			// no additional indent before "milk that is:"
			input: `
- Tuesday
 - oatmeal
 - milk
 milk that is:
  - low-fat`,
			want: &node.Document{
				Children: []node.Node{
					&node.List{
						Type: node.UnorderedList,
						ListItems: []*node.ListItem{
							{
								Children: []node.Node{
									node.Lines{
										&node.Line{
											Children: []node.Inline{
												&node.Text{
													Value: "Tuesday",
												},
											},
										},
									},
									&node.List{
										Type: node.UnorderedList,
										ListItems: []*node.ListItem{
											{
												Children: []node.Node{
													node.Lines{
														&node.Line{
															Children: []node.Inline{
																&node.Text{
																	Value: "oatmeal",
																},
															},
														},
													},
												},
											},
											{
												Children: []node.Node{
													node.Lines{
														&node.Line{
															Children: []node.Inline{
																&node.Text{
																	Value: "milk",
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Text{
										Value: "milk that is:",
									},
								},
							},
						},
					},
					&node.List{
						Type: node.UnorderedList,
						ListItems: []*node.ListItem{
							{
								Children: []node.Node{
									node.Lines{
										&node.Line{
											Children: []node.Inline{
												&node.Text{
													Value: "low-fat",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "unordered list deeply nested 5",
			input: `
- Tuesday
 - Shopping list:
  Go buy some milk and sugar
 - Work:
  Climb the bamboo and eat eucalyptus
- Wednesday
`,
			want: &node.Document{
				Children: []node.Node{
					&node.List{
						Type: node.UnorderedList,
						ListItems: []*node.ListItem{
							{
								Children: []node.Node{
									node.Lines{
										&node.Line{
											Children: []node.Inline{
												&node.Text{
													Value: "Tuesday",
												},
											},
										},
									},
									&node.List{
										Type: node.UnorderedList,
										ListItems: []*node.ListItem{
											{
												Children: []node.Node{
													node.Lines{
														{
															Children: []node.Inline{
																&node.Text{
																	Value: "Shopping list:",
																},
															},
														},
														{
															Children: []node.Inline{
																&node.Text{
																	Value: "Go buy some milk and sugar",
																},
															},
														},
													},
												},
											},
											{
												Children: []node.Node{
													node.Lines{
														{
															Children: []node.Inline{
																&node.Text{
																	Value: "Work:",
																},
															},
														},
														{
															Children: []node.Inline{
																&node.Text{
																	Value: "Climb the bamboo and eat eucalyptus",
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
							{
								Children: []node.Node{
									node.Lines{
										&node.Line{
											Children: []node.Inline{
												&node.Text{
													Value: "Wednesday",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "unordered list deeply nested 6",
			input: `
List
- 1
 - 2
  - 3
   - 4
  - 5
 - 6
- 7
End
			`,
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Text{
										Value: "List",
									},
								},
							},
						},
					},
					&node.List{
						Type: node.UnorderedList,
						ListItems: []*node.ListItem{
							{
								Children: []node.Node{
									node.Lines{
										&node.Line{
											Children: []node.Inline{
												&node.Text{
													Value: "1",
												},
											},
										},
									},
									&node.List{
										Type: node.UnorderedList,
										ListItems: []*node.ListItem{
											{
												Children: []node.Node{
													node.Lines{
														&node.Line{
															Children: []node.Inline{
																&node.Text{
																	Value: "2",
																},
															},
														},
													},
													&node.List{
														Type: node.UnorderedList,
														ListItems: []*node.ListItem{
															{
																Children: []node.Node{
																	node.Lines{
																		&node.Line{
																			Children: []node.Inline{
																				&node.Text{
																					Value: "3",
																				},
																			},
																		},
																	},
																	&node.List{
																		Type: node.UnorderedList,
																		ListItems: []*node.ListItem{
																			{
																				Children: []node.Node{
																					node.Lines{
																						&node.Line{
																							Children: []node.Inline{
																								&node.Text{
																									Value: "4",
																								},
																							},
																						},
																					},
																				},
																			},
																		},
																	},
																},
															},
															{
																Children: []node.Node{
																	node.Lines{
																		&node.Line{
																			Children: []node.Inline{
																				&node.Text{
																					Value: "5",
																				},
																			},
																		},
																	},
																},
															},
														},
													},
												},
											},
											{
												Children: []node.Node{
													node.Lines{
														&node.Line{
															Children: []node.Inline{
																&node.Text{
																	Value: "6",
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
							{
								Children: []node.Node{
									node.Lines{
										&node.Line{
											Children: []node.Inline{
												&node.Text{
													Value: "7",
												},
											},
										},
									},
								},
							},
						},
					},
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Text{
										Value: "End",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "unordered list deeply nested 7",
			// list is indented
			input: `
List
  - 1
   - 2
    - 3
     - 4
    - 5
   - 6
  - 7
End
			`,
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Text{
										Value: "List",
									},
								},
							},
						},
					},
					&node.List{
						Type: node.UnorderedList,
						ListItems: []*node.ListItem{
							{
								Children: []node.Node{
									node.Lines{
										&node.Line{
											Children: []node.Inline{
												&node.Text{
													Value: "1",
												},
											},
										},
									},
									&node.List{
										Type: node.UnorderedList,
										ListItems: []*node.ListItem{
											{
												Children: []node.Node{
													node.Lines{
														&node.Line{
															Children: []node.Inline{
																&node.Text{
																	Value: "2",
																},
															},
														},
													},
													&node.List{
														Type: node.UnorderedList,
														ListItems: []*node.ListItem{
															{
																Children: []node.Node{
																	node.Lines{
																		&node.Line{
																			Children: []node.Inline{
																				&node.Text{
																					Value: "3",
																				},
																			},
																		},
																	},
																	&node.List{
																		Type: node.UnorderedList,
																		ListItems: []*node.ListItem{
																			{
																				Children: []node.Node{
																					node.Lines{
																						&node.Line{
																							Children: []node.Inline{
																								&node.Text{
																									Value: "4",
																								},
																							},
																						},
																					},
																				},
																			},
																		},
																	},
																},
															},
															{
																Children: []node.Node{
																	node.Lines{
																		&node.Line{
																			Children: []node.Inline{
																				&node.Text{
																					Value: "5",
																				},
																			},
																		},
																	},
																},
															},
														},
													},
												},
											},
											{
												Children: []node.Node{
													node.Lines{
														&node.Line{
															Children: []node.Inline{
																&node.Text{
																	Value: "6",
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
							{
								Children: []node.Node{
									node.Lines{
										&node.Line{
											Children: []node.Inline{
												&node.Text{
													Value: "7",
												},
											},
										},
									},
								},
							},
						},
					},
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Text{
										Value: "End",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "unordered list deeply nested 8",
			input: `
List
- 1
 one
 - 2
  two
  - 3
   three
   - 4
    four
  - 5
   five
 - 6
  six
- 7
 seven
End
			`,
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Text{
										Value: "List",
									},
								},
							},
						},
					},
					&node.List{
						Type: node.UnorderedList,
						ListItems: []*node.ListItem{
							{
								Children: []node.Node{
									node.Lines{
										{
											Children: []node.Inline{
												&node.Text{
													Value: "1",
												},
											},
										},
										{
											Children: []node.Inline{
												&node.Text{
													Value: "one",
												},
											},
										},
									},
									&node.List{
										Type: node.UnorderedList,
										ListItems: []*node.ListItem{
											{
												Children: []node.Node{
													node.Lines{
														{
															Children: []node.Inline{
																&node.Text{
																	Value: "2",
																},
															},
														},
														{
															Children: []node.Inline{
																&node.Text{
																	Value: "two",
																},
															},
														},
													},
													&node.List{
														Type: node.UnorderedList,
														ListItems: []*node.ListItem{
															{
																Children: []node.Node{
																	node.Lines{
																		{
																			Children: []node.Inline{
																				&node.Text{
																					Value: "3",
																				},
																			},
																		},
																		{
																			Children: []node.Inline{
																				&node.Text{
																					Value: "three",
																				},
																			},
																		},
																	},
																	&node.List{
																		Type: node.UnorderedList,
																		ListItems: []*node.ListItem{
																			{
																				Children: []node.Node{
																					node.Lines{
																						&node.Line{
																							Children: []node.Inline{
																								&node.Text{
																									Value: "4",
																								},
																							},
																						},
																						&node.Line{
																							Children: []node.Inline{
																								&node.Text{
																									Value: "four",
																								},
																							},
																						},
																					},
																				},
																			},
																		},
																	},
																},
															},
															{
																Children: []node.Node{
																	node.Lines{
																		&node.Line{
																			Children: []node.Inline{
																				&node.Text{
																					Value: "5",
																				},
																			},
																		},
																		&node.Line{
																			Children: []node.Inline{
																				&node.Text{
																					Value: "five",
																				},
																			},
																		},
																	},
																},
															},
														},
													},
												},
											},
											{
												Children: []node.Node{
													node.Lines{
														&node.Line{
															Children: []node.Inline{
																&node.Text{
																	Value: "6",
																},
															},
														},
														&node.Line{
															Children: []node.Inline{
																&node.Text{
																	Value: "six",
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
							{
								Children: []node.Node{
									node.Lines{
										&node.Line{
											Children: []node.Inline{
												&node.Text{
													Value: "7",
												},
											},
										},
										&node.Line{
											Children: []node.Inline{
												&node.Text{
													Value: "seven",
												},
											},
										},
									},
								},
							},
						},
					},
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Text{
										Value: "End",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "list no identation",
			input: "-1\n-2",
			want: &node.Document{
				Children: []node.Node{
					&node.List{
						ListItems: []*node.ListItem{
							{
								Children: []node.Node{
									node.Lines{
										{
											Children: []node.Inline{
												&node.Text{
													Value: "1",
												},
											},
										},
									},
								},
							},
							{
								Children: []node.Node{
									node.Lines{
										{
											Children: []node.Inline{
												&node.Text{
													Value: "2",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "list space identation",
			input: "-1\n -2",
			want: &node.Document{
				Children: []node.Node{
					&node.List{
						ListItems: []*node.ListItem{
							{
								Children: []node.Node{
									node.Lines{
										{
											Children: []node.Inline{
												&node.Text{
													Value: "1",
												},
											},
										},
									},
									&node.List{
										ListItems: []*node.ListItem{
											{
												Children: []node.Node{
													node.Lines{
														{
															Children: []node.Inline{
																&node.Text{
																	Value: "2",
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "list tab identation",
			input: "-1\n\t-2",
			want: &node.Document{
				Children: []node.Node{
					&node.List{
						ListItems: []*node.ListItem{
							{
								Children: []node.Node{
									node.Lines{
										{
											Children: []node.Inline{
												&node.Text{
													Value: "1",
												},
											},
										},
									},
									&node.List{
										ListItems: []*node.ListItem{
											{
												Children: []node.Node{
													node.Lines{
														{
															Children: []node.Inline{
																&node.Text{
																	Value: "2",
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "list one space zero space identation",
			input: " -1\n-2",
			want: &node.Document{
				Children: []node.Node{
					&node.List{
						ListItems: []*node.ListItem{
							{
								Children: []node.Node{
									node.Lines{
										{
											Children: []node.Inline{
												&node.Text{
													Value: "1",
												},
											},
										},
									},
								},
							},
						},
					},
					&node.List{
						ListItems: []*node.ListItem{
							{
								Children: []node.Node{
									node.Lines{
										{
											Children: []node.Inline{
												&node.Text{
													Value: "2",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "list one space one space identation",
			input: " -1\n -2",
			want: &node.Document{
				Children: []node.Node{
					&node.List{
						ListItems: []*node.ListItem{
							{
								Children: []node.Node{
									node.Lines{
										{
											Children: []node.Inline{
												&node.Text{
													Value: "1",
												},
											},
										},
									},
								},
							},
							{
								Children: []node.Node{
									node.Lines{
										{
											Children: []node.Inline{
												&node.Text{
													Value: "2",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "list one space one tab identation",
			input: " -1\n\t-2",
			want: &node.Document{
				Children: []node.Node{
					&node.List{
						ListItems: []*node.ListItem{
							{
								Children: []node.Node{
									node.Lines{
										{
											Children: []node.Inline{
												&node.Text{
													Value: "1",
												},
											},
										},
									},
									&node.List{
										ListItems: []*node.ListItem{
											{
												Children: []node.Node{
													node.Lines{
														{
															Children: []node.Inline{
																&node.Text{
																	Value: "2",
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "list eight spaces one tab identation",
			input: "        -1\n\t-2",
			want: &node.Document{
				Children: []node.Node{
					&node.List{
						ListItems: []*node.ListItem{
							{
								Children: []node.Node{
									node.Lines{
										{
											Children: []node.Inline{
												&node.Text{
													Value: "1",
												},
											},
										},
									},
								},
							},
							{
								Children: []node.Node{
									node.Lines{
										{
											Children: []node.Inline{
												&node.Text{
													Value: "2",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "list nine spaces one tab identation",
			input: "         -1\n\t-2",
			want: &node.Document{
				Children: []node.Node{
					&node.List{
						ListItems: []*node.ListItem{
							{
								Children: []node.Node{
									node.Lines{
										{
											Children: []node.Inline{
												&node.Text{
													Value: "1",
												},
											},
										},
									},
								},
							},
						},
					},
					&node.List{
						ListItems: []*node.ListItem{
							{
								Children: []node.Node{
									node.Lines{
										{
											Children: []node.Inline{
												&node.Text{
													Value: "2",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "list nine spaces one tab + one space identation",
			input: "         -1\n\t -2",
			want: &node.Document{
				Children: []node.Node{
					&node.List{
						ListItems: []*node.ListItem{
							{
								Children: []node.Node{
									node.Lines{
										{
											Children: []node.Inline{
												&node.Text{
													Value: "1",
												},
											},
										},
									},
								},
							},
							{
								Children: []node.Node{
									node.Lines{
										{
											Children: []node.Inline{
												&node.Text{
													Value: "2",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "list one tab two spaces identation",
			input: "\t-1\n  -2",
			want: &node.Document{
				Children: []node.Node{
					&node.List{
						ListItems: []*node.ListItem{
							{
								Children: []node.Node{
									node.Lines{
										{
											Children: []node.Inline{
												&node.Text{
													Value: "1",
												},
											},
										},
									},
								},
							},
						},
					},
					&node.List{
						ListItems: []*node.ListItem{
							{
								Children: []node.Node{
									node.Lines{
										{
											Children: []node.Inline{
												&node.Text{
													Value: "2",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "escape sequences",
			input: "\\\\ \\< \\> \\_ \\* \\= \\# \\` \\-",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Text{
										Value: `\ < > _ * = # ` + "`" + ` -`,
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "escape backslash 1",
			input: `random \ text`,
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Text{
										Value: `random \ text`,
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "escape backslash 2",
			input: `random \\ text`,
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Text{
										Value: `random \ text`,
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "escape backslash 3",
			input: `<random \ text>`,
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Link{
										Destination: `random \ text`,
										Children: []node.Inline{
											&node.Text{
												Value: `random \ text`,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "escape backslash 4",
			input: `<random \\ text>`,
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Link{
										Destination: `random \ text`,
										Children: []node.Inline{
											&node.Text{
												Value: `random \ text`,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "escape backslash 5",
			input: `<random \\> text>`,
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Link{
										Destination: `random \`,
										Children: []node.Inline{
											&node.Text{
												Value: `random \`,
											},
										},
									},
									&node.Text{
										Value: " text>",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "escape inline backslash",
			input: `\_\\_a double underscore`,
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Text{
										Value: `_\_a double underscore`,
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "escape emphasis 1",
			input: `\__a double underscore`,
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Text{
										Value: "__a double underscore",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "escape emphasis 2",
			input: `\_\__a triple underscore`,
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Text{
										Value: "___a triple underscore",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "escape emphasis 3",
			input: `_\_a double underscore`,
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Text{
										Value: "__a double underscore",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "escape emphasis 4",
			input: `\_\_a double underscore`,
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Text{
										Value: "__a double underscore",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "escape emphasis intraword",
			input: `file\__name`,
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Text{
										Value: "file__name",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "escape emphasis at end",
			input: `filename\__`,
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Text{
										Value: "filename__",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "escape strong",
			input: `\**`,
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Text{
										Value: "**",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "escape link 1",
			input: `\<>`,
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Text{
										Value: "<>",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "escape link 2",
			input: `<\>`,
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Link{
										Destination: ">",
										Children: []node.Inline{
											&node.Text{
												Value: ">",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "escape link 3",
			input: `\>`,
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Text{
										Value: `>`,
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "escape link 4",
			input: `\<\>`,
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Text{
										Value: `<>`,
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "escape < in link text 1",
			input: `<<><>`,
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Link{
										Destination: "",
										Children: []node.Inline{
											&node.Text{
												Value: "<",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "escape < in link text 2",
			input: `<\<><>`,
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Link{
										Destination: "",
										Children: []node.Inline{
											&node.Text{
												Value: `\<`,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "escape > in link text 1",
			input: `<>><>`,
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Link{
										Destination: "",
										Children: []node.Inline{
											&node.Text{},
										},
									},
									&node.Text{
										Value: ">",
									},
									&node.Link{
										Destination: "",
										Children: []node.Inline{
											&node.Text{},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "escape > in link text 2",
			input: `<\>><>`,
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Link{
										Destination: "",
										Children: []node.Inline{
											&node.Text{
												Value: ">",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "escape < in link destination 1",
			input: `<><<>`,
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Link{
										Destination: "<",
										Children:    []node.Inline{},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "escape < in link destination 2",
			input: `<><\<>`,
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Link{
										Destination: `\<`,
										Children:    []node.Inline{},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "escape > in link destination 1",
			input: `<><>>`,
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Link{
										Destination: "",
										Children:    []node.Inline{},
									},
									&node.Text{
										Value: ">",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "escape > in link destination 2",
			input: `<><\>>`,
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Link{
										Destination: ">",
										Children:    []node.Inline{},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "escape backslash in link 1",
			input: `<\>`,
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Link{
										Destination: ">",
										Children: []node.Inline{
											&node.Text{
												Value: ">",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "escape backslash in link 2",
			input: `<\\>`,
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Link{
										Destination: `\`,
										Children: []node.Inline{
											&node.Text{
												Value: `\`,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "escape backslash in link 3",
			input: `<\\\>`,
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Link{
										Destination: `\>`,
										Children: []node.Inline{
											&node.Text{
												Value: `\>`,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "escape backslash in link 4",
			input: `<in bet\ween>`,
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Link{
										Destination: `in bet\ween`,
										Children: []node.Inline{
											&node.Text{
												Value: `in bet\ween`,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "escape backslash in link 5",
			input: `<in bet\\ween>`,
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Link{
										Destination: `in bet\ween`,
										Children: []node.Inline{
											&node.Text{
												Value: `in bet\ween`,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "escape backslash in link text 1",
			input: `<\>><>`,
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Link{
										Destination: "",
										Children: []node.Inline{
											&node.Text{
												Value: ">",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "escape backslash in link text 2",
			input: `<\\><>`,
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Link{
										Destination: "",
										Children: []node.Inline{
											&node.Text{
												Value: `\`,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "escape backslash in link text 3",
			input: `<\\\>><>`,
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Link{
										Destination: "",
										Children: []node.Inline{
											&node.Text{
												Value: `\>`,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "escape backslash in link text 4",
			input: `<in bet\ween><>`,
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Link{
										Destination: "",
										Children: []node.Inline{
											&node.Text{
												Value: `in bet\ween`,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "escape backslash in link text 5",
			input: `<in bet\\ween><>`,
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Link{
										Destination: "",
										Children: []node.Inline{
											&node.Text{
												Value: `in bet\ween`,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "escape backslash in link destination",
			input: `<><\\\>>`,
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Link{
										Destination: `\>`,
										Children:    []node.Inline{},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "escape block backslash",
			input: `\=\\=`,
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Text{
										Value: `=\=`,
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "escape heading 1",
			input: `\=`,
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Text{
										Value: "=",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "escape heading 2",
			input: `\==`,
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Text{
										Value: "==",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "escape heading 3", // not actually escaped
			input: `=\=`,
			want: &node.Document{
				Children: []node.Node{
					&node.Heading{
						Level: 1,
						Children: []node.Inline{
							&node.Text{
								Value: "=",
							},
						},
					},
				},
			},
		},
		{
			name:  "escape heading 4",
			input: `\=\=`,
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Text{
										Value: "==",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "escape heading after paragraph",
			input: `paragraph
\=`,
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Text{
										Value: "paragraph",
									},
								},
							},
							{
								Children: []node.Inline{
									&node.Text{
										Value: "=",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "escape numbered heading",
			input: `\##`,
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Text{
										Value: "##",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "escape numbered heading after paragraph",
			input: `paragraph
\##`,
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Text{
										Value: "paragraph",
									},
								},
							},
							{
								Children: []node.Inline{
									&node.Text{
										Value: "##",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "escape code block",
			input: "\\``",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Text{
										Value: "``",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "escape code block after paragraph",
			input: "paragraph\n\\``",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Text{
										Value: "paragraph",
									},
								},
							},
							{
								Children: []node.Inline{
									&node.Text{
										Value: "``",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "escape unordered list",
			input: `\-`,
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Text{
										Value: `-`,
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "escape unordered list after paragraph",
			input: `paragraph
\-`,
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Text{
										Value: "paragraph",
									},
								},
							},
							{
								Children: []node.Inline{
									&node.Text{
										Value: `-`,
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "block quote",
			input: "> Get me some leafs",
			want: &node.Document{
				Children: []node.Node{
					&node.BlockQuote{
						Children: []node.Block{
							&node.Paragraph{
								Lines: node.Lines{
									{
										Children: []node.Inline{
											&node.Text{
												Value: "Get me some leafs",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "block quote multiline",
			input: `
> Get me some leafs
> and some juice
`,
			want: &node.Document{
				Children: []node.Node{
					&node.BlockQuote{
						Children: []node.Block{
							&node.Paragraph{
								Lines: node.Lines{
									{
										Children: []node.Inline{
											&node.Text{
												Value: "Get me some leafs",
											},
										},
									},
									{
										Children: []node.Inline{
											&node.Text{
												Value: "and some juice",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cmpPretty(t, cmpData{
				name:  tc.name,
				input: tc.input,
				want:  tc.want,
			})
		})
	}
}

func TestBOM(t *testing.T) {
	const BOM = "\uFEFF"

	cases := []struct {
		name      string
		isAllowed bool
		input     string
		want      *node.Document
	}{
		{
			name:      "BOM at the beginning",
			isAllowed: true,
			input:     BOM + "= Koala",
			want: &node.Document{
				Children: []node.Node{
					&node.Heading{
						Level: 1,
						Children: []node.Inline{
							&node.Text{
								Value: "Koala",
							},
						},
					},
				},
			},
		},
		{
			name:      "BOM in the middle",
			isAllowed: false,
			input:     "= Ko" + BOM + "ala",
			want: &node.Document{
				Children: []node.Node{
					&node.Heading{
						Level: 1,
						Children: []node.Inline{
							&node.Text{
								Value: "Ko" + BOM + "ala",
							},
						},
					},
				},
			},
		},
		{
			name:      "BOM at the end",
			isAllowed: false,
			input:     "= Koala" + BOM,
			want: &node.Document{
				Children: []node.Node{
					&node.Heading{
						Level: 1,
						Children: []node.Inline{
							&node.Text{
								Value: "Koala" + BOM,
							},
						},
					},
				},
			},
		},
	}

	var errHandler = func(err error, count uint) {
		if count != 1 || !errors.Is(err, parser.ErrIllegalBOM) {
			t.Error(err)
		}
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.isAllowed {
				cmpPretty(t, cmpData{
					name:  tc.name,
					input: tc.input,
					want:  tc.want,
				})
				return
			}

			cmpPretty(t, cmpData{
				name:       tc.name,
				input:      tc.input,
				want:       tc.want,
				errCount:   1,
				errHandler: errHandler,
			})

		})
	}
}

func TestNUL(t *testing.T) {
	const NUL = "\u0000"

	cases := []struct {
		name  string
		input string
		want  *node.Document
	}{
		{
			name:  "NUL at the beginning",
			input: NUL + "= Koala",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Text{
										Value: string(utf8.RuneError) + "= Koala",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "NUL in the middle",
			input: "= Ko" + NUL + "ala",
			want: &node.Document{
				Children: []node.Node{
					&node.Heading{
						Level: 1,
						Children: []node.Inline{
							&node.Text{
								Value: "Ko" + string(utf8.RuneError) + "ala",
							},
						},
					},
				},
			},
		},
		{
			name:  "NUL at the end",
			input: "= Koala" + NUL,
			want: &node.Document{
				Children: []node.Node{
					&node.Heading{
						Level: 1,
						Children: []node.Inline{
							&node.Text{
								Value: "Koala" + string(utf8.RuneError),
							},
						},
					},
				},
			},
		},
	}

	var errHandler = func(err error, count uint) {
		if count != 1 || !errors.Is(err, parser.ErrIllegalNUL) {
			t.Error(err)
		}
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cmpPretty(t, cmpData{
				name:       tc.name,
				input:      tc.input,
				want:       tc.want,
				errCount:   1,
				errHandler: errHandler,
			})
		})
	}
}

func TestIllegalUTF8Encoding(t *testing.T) {
	const fcb = "\x80" // first continuation byte

	cases := []struct {
		name  string
		input string
		want  *node.Document
	}{
		{
			name:  "fcb at the beginning",
			input: fcb + "= Koala",
			want: &node.Document{
				Children: []node.Node{
					&node.Heading{
						Level: 1,
						Children: []node.Inline{
							&node.Text{
								Value: "Koala", // fcb is skipped as it is the first...
							},
						},
					},
				},
			},
		},
		{
			name:  "fcb in the middle",
			input: "= Ko" + fcb + "ala",
			want: &node.Document{
				Children: []node.Node{
					&node.Heading{
						Level: 1,
						Children: []node.Inline{
							&node.Text{
								Value: "Ko" + fcb + "ala",
							},
						},
					},
				},
			},
		},
		{
			name:  "fcb at the end",
			input: "= Koala" + fcb,
			want: &node.Document{
				Children: []node.Node{
					&node.Heading{
						Level: 1,
						Children: []node.Inline{
							&node.Text{
								Value: "Koala" + fcb,
							},
						},
					},
				},
			},
		},
	}

	var errHandler = func(err error, count uint) {
		if count != 1 || !errors.Is(err, parser.ErrIllegalUTF8Encoding) {
			t.Error(err)
		}
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cmpPretty(t, cmpData{
				name:       tc.name,
				input:      tc.input,
				want:       tc.want,
				errCount:   1,
				errHandler: errHandler,
			})
		})
	}
}

func TestUnicode(t *testing.T) {
	const eur = "\u20AC" // euro sign

	cases := []struct {
		name  string
		input string
		want  *node.Document
	}{
		{
			name:  "code point at the beginning",
			input: eur + "= Koala",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Lines: node.Lines{
							{
								Children: []node.Inline{
									&node.Text{
										Value: eur + "= Koala",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "code point in the middle",
			input: "= Ko" + eur + "ala",
			want: &node.Document{
				Children: []node.Node{
					&node.Heading{
						Level: 1,
						Children: []node.Inline{
							&node.Text{
								Value: "Ko" + eur + "ala",
							},
						},
					},
				},
			},
		},
		{
			name:  "code point at the end",
			input: "= Koala" + eur,
			want: &node.Document{
				Children: []node.Node{
					&node.Heading{
						Level: 1,
						Children: []node.Inline{
							&node.Text{
								Value: "Koala" + eur,
							},
						},
					},
				},
			},
		},
	}

	var errHandler = func(err error, count uint) {
		if count > 0 {
			t.Error(err)
		}
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cmpPretty(t, cmpData{
				name:       tc.name,
				input:      tc.input,
				want:       tc.want,
				errCount:   0,
				errHandler: errHandler,
			})
		})
	}
}

type cmpData struct {
	name       string
	input      string
	want       *node.Document
	errCount   uint
	errHandler parser.ErrorHandler
}

func cmpPretty(t *testing.T, data cmpData) {
	t.Helper()

	p := parser.New(data.input, data.errHandler)

	doc, errCount := p.ParseDocument()
	if doc == nil {
		t.Fatalf("ParseDocument() returned nil")
	}
	if errCount != data.errCount {
		t.Errorf("got %d errors, want %d", errCount, data.errCount)
	}

	gotPretty := printer.Pretty(doc, 0)
	wantPretty := printer.Pretty(data.want, 0)

	if gotPretty != wantPretty {
		t.Errorf(
			"document \"%s\" is incorrect, from input `%s`\ngot:\n%s\nwant:\n%s",
			data.name,
			data.input,
			gotPretty,
			wantPretty,
		)
	}
}
