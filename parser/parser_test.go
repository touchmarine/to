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
		want  *node.Document
	}{
		{
			name:  "underscore",
			input: "Tibsey is a _koala_.",
			want: &node.Document{
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
			want: &node.Document{
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
			want: &node.Document{
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
			want: &node.Document{
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
			want: &node.Document{
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
			want: &node.Document{
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
			want: &node.Document{
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
			want: &node.Document{
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
			want: &node.Document{
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
			want: &node.Document{
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
			want: &node.Document{
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
			want: &node.Document{
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
			want: &node.Document{
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
			want: &node.Document{
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
						Children: []node.Inline{
							&node.Text{
								Value: "# Koala",
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
						Children: []node.Inline{
							&node.Text{
								Value: "The koala is an iconic Australian animal. Often called...",
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
						Children: []node.Inline{
							&node.Text{
								Value: "The koala is an iconic Australian animal. Often called...",
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
						Children: []node.Inline{
							&node.Text{
								Value: "Koala lives in the...",
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
						Children: []node.Inline{
							&node.Text{
								Value: "Koala lives in the...",
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
						Children: []node.Inline{
							&node.Text{
								Value: "The koala is an iconic Australian animal. Often called...",
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
						Children: []node.Inline{
							&node.Text{
								Value: "The koala is an iconic Australian animal. Often called...",
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
						Children: []node.Inline{
							&node.Text{
								Value: "The koala is an iconic Australian animal. Often called...",
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
						Children: []node.Inline{
							&node.Text{
								Value: "The koala is an iconic Australian animal. Often called...",
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
			name:  "link",
			input: "<https://koala.test>",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
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
		{
			name:  "relative link",
			input: "</koalas>",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
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
		{
			name:  "reference link",
			input: "<#habitat>",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
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
		{
			name:  "email link",
			input: "<mailto:https://koala.test>",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
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
		{
			name:  "link after unterminated strong",
			input: "**<https://koala.test> koalas are awesome",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
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
		{
			name:  "link with strong",
			input: "Look at <https://**koala**.test> website",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
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
		{
			name:  "link with unterminated strong",
			input: "Look at <https://**koala.test> website",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
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
		{
			name:  "unterminated link",
			input: "Look at <https://koala.test website",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
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
		{
			name:  "unterminated link another link",
			input: "Look at <https://koala.test website <#habitat>",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
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
		{
			name:  "two-part link",
			input: "<Koala bears><https://koala.test>",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
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
		{
			name:  "two-part link empahsized",
			input: "__<Koala bears><https://koala.test>__",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
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
		{
			name:  "two-part link strong text",
			input: "<**Koala bears**><https://koala.test>",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
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
		{
			name:  "two-part link unterminated strong text",
			input: "<**Koala bears><https://koala.test>",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
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
		{
			name:  "link with <",
			input: "<Koala bears<https://koala.test>",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
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
		{
			name:  "link with < unterminated strong",
			input: "<**Koala bears<https://koala.test>",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
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
		{
			name:  "link with < and strong",
			input: "<**Koala bears**<https://koala.test>",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
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
		{
			name:  "two-part link strong and emphasis",
			input: "<**__Koala__ bears**><https://koala.test>",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
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
		{
			name:  "two-part link plain destination",
			input: "<Koala bears><https://**koala**.test>",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
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
		{
			name:  "consecutive links",
			input: "<https://koala.test><https://eucalyptus.test>",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
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
		{
			name:  "two links",
			input: "<https://koala.test> <https://eucalyptus.test>",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
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
		{
			name:  "consecutive two-part links",
			input: "<Koala bears><https://koala.test><Eucalyptus><https://eucalyptus.test>",
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
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
		{
			name: "links in paragraphs",
			input: `<https://koala.test>
<https://eucalyptus.test>

<#habitat>`,
			want: &node.Document{
				Children: []node.Node{
					&node.Paragraph{
						Children: []node.Inline{
							&node.Link{
								Destination: "https://koala.test",
								Children: []node.Inline{
									&node.Text{
										Value: "https://koala.test",
									},
								},
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
					&node.Paragraph{
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
						Children: []node.Inline{
							&node.Text{
								Value: "ALREADY A PARAGRAPH",
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
			name:  "unordered list",
			input: "- milk",
			want: &node.Document{
				Children: []node.Node{
					&node.List{
						Type: node.UnorderedList,
						ListItems: [][]node.Node{
							{
								&node.Text{
									Value: " milk",
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
						ListItems: [][]node.Node{
							{
								&node.Text{
									Value: " milk",
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
						ListItems: [][]node.Node{
							{
								&node.Text{
									Value: "milk",
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
						ListItems: [][]node.Node{
							{
								&node.Text{
									Value: " milk",
								},
							},
							{
								&node.Text{
									Value: " sugar",
								},
							},
							{
								&node.Text{
									Value: " bananas",
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
						ListItems: [][]node.Node{
							{
								&node.Text{
									Value: " milk",
								},
							},
							{
								&node.Text{
									Value: " sugar",
								},
							},
							{
								&node.Text{
									Value: " bananas",
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
						ListItems: [][]node.Node{
							{
								&node.Text{
									Value: " Tuesday",
								},
								&node.List{
									Type: node.UnorderedList,
									ListItems: [][]node.Node{
										{
											&node.Text{
												Value: " milk",
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
						ListItems: [][]node.Node{
							{
								&node.Text{
									Value: " Tuesday",
								},
								&node.List{
									Type: node.UnorderedList,
									ListItems: [][]node.Node{
										{
											&node.Text{
												Value: " milk",
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
						ListItems: [][]node.Node{
							{
								&node.Text{
									Value: " Tuesday",
								},
								&node.Text{
									Value: "Shopping list:",
								},
								&node.List{
									Type: node.UnorderedList,
									ListItems: [][]node.Node{
										{
											&node.Text{
												Value: " milk",
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
						ListItems: [][]node.Node{
							{
								&node.List{
									Type: node.UnorderedList,
									ListItems: [][]node.Node{
										{
											&node.Text{
												Value: " inner",
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
						ListItems: [][]node.Node{
							{
								&node.List{
									Type: node.UnorderedList,
									ListItems: [][]node.Node{
										{
											&node.Text{
												Value: " inner",
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
						ListItems: [][]node.Node{
							{
								&node.Text{
									Value: " Tuesday",
								},
								&node.List{
									Type: node.UnorderedList,
									ListItems: [][]node.Node{
										{
											&node.Text{
												Value: " milk",
											},
											&node.List{
												Type: node.UnorderedList,
												ListItems: [][]node.Node{
													{
														&node.Text{
															Value: " low-fat",
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
						ListItems: [][]node.Node{
							{
								&node.Text{
									Value: " Tuesday",
								},
								&node.List{
									Type: node.UnorderedList,
									ListItems: [][]node.Node{
										{
											&node.Text{
												Value: " milk",
											},
											&node.List{
												Type: node.UnorderedList,
												ListItems: [][]node.Node{
													{
														&node.Text{
															Value: " low-fat",
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
						ListItems: [][]node.Node{
							{
								&node.Text{
									Value: " Tuesday",
								},
								&node.List{
									Type: node.UnorderedList,
									ListItems: [][]node.Node{
										{
											&node.Text{
												Value: " oatmeal",
											},
										},
										{
											&node.Text{
												Value: " milk",
											},
											&node.Text{
												Value: "milk that is:",
											},
											&node.List{
												Type: node.UnorderedList,
												ListItems: [][]node.Node{
													{
														&node.Text{
															Value: " low-fat",
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
						ListItems: [][]node.Node{
							{
								&node.Text{
									Value: " Tuesday",
								},
								&node.List{
									Type: node.UnorderedList,
									ListItems: [][]node.Node{
										{
											&node.Text{
												Value: " oatmeal",
											},
										},
										{
											&node.Text{
												Value: " milk",
											},
										},
									},
								},
							},
						},
					},
					&node.Paragraph{
						Children: []node.Inline{
							&node.Text{
								Value: "milk that is:",
							},
						},
					},
					&node.List{
						Type: node.UnorderedList,
						ListItems: [][]node.Node{
							{
								&node.Text{
									Value: " low-fat",
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
						ListItems: [][]node.Node{
							{
								&node.Text{
									Value: " Tuesday",
								},
								&node.List{
									Type: node.UnorderedList,
									ListItems: [][]node.Node{
										{
											&node.Text{
												Value: " Shopping list:",
											},
											&node.Text{
												Value: "Go buy some milk and sugar",
											},
										},
										{
											&node.Text{
												Value: " Work:",
											},
											&node.Text{
												Value: "Climb the bamboo and eat eucalyptus",
											},
										},
									},
								},
							},
							{
								&node.Text{
									Value: " Wednesday",
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
						Children: []node.Inline{
							&node.Text{
								Value: "List",
							},
						},
					},
					&node.List{
						Type: node.UnorderedList,
						ListItems: [][]node.Node{
							{
								&node.Text{
									Value: " 1",
								},
								&node.List{
									Type: node.UnorderedList,
									ListItems: [][]node.Node{
										{
											&node.Text{
												Value: " 2",
											},
											&node.List{
												Type: node.UnorderedList,
												ListItems: [][]node.Node{
													{
														&node.Text{
															Value: " 3",
														},
														&node.List{
															Type: node.UnorderedList,
															ListItems: [][]node.Node{
																{
																	&node.Text{
																		Value: " 4",
																	},
																},
															},
														},
													},
													{
														&node.Text{
															Value: " 5",
														},
													},
												},
											},
										},
										{
											&node.Text{
												Value: " 6",
											},
										},
									},
								},
							},
							{
								&node.Text{
									Value: " 7",
								},
							},
						},
					},
					&node.Paragraph{
						Children: []node.Inline{
							&node.Text{
								Value: "End",
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
						Children: []node.Inline{
							&node.Text{
								Value: "List",
							},
						},
					},
					&node.List{
						Type: node.UnorderedList,
						ListItems: [][]node.Node{
							{
								&node.Text{
									Value: " 1",
								},
								&node.List{
									Type: node.UnorderedList,
									ListItems: [][]node.Node{
										{
											&node.Text{
												Value: " 2",
											},
											&node.List{
												Type: node.UnorderedList,
												ListItems: [][]node.Node{
													{
														&node.Text{
															Value: " 3",
														},
														&node.List{
															Type: node.UnorderedList,
															ListItems: [][]node.Node{
																{
																	&node.Text{
																		Value: " 4",
																	},
																},
															},
														},
													},
													{
														&node.Text{
															Value: " 5",
														},
													},
												},
											},
										},
										{
											&node.Text{
												Value: " 6",
											},
										},
									},
								},
							},
							{
								&node.Text{
									Value: " 7",
								},
							},
						},
					},
					&node.Paragraph{
						Children: []node.Inline{
							&node.Text{
								Value: "End",
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
						Children: []node.Inline{
							&node.Text{
								Value: "List",
							},
						},
					},
					&node.List{
						Type: node.UnorderedList,
						ListItems: [][]node.Node{
							{
								&node.Text{
									Value: " 1",
								},
								&node.Text{
									Value: "one",
								},
								&node.List{
									Type: node.UnorderedList,
									ListItems: [][]node.Node{
										{
											&node.Text{
												Value: " 2",
											},
											&node.Text{
												Value: "two",
											},
											&node.List{
												Type: node.UnorderedList,
												ListItems: [][]node.Node{
													{
														&node.Text{
															Value: " 3",
														},
														&node.Text{
															Value: "three",
														},
														&node.List{
															Type: node.UnorderedList,
															ListItems: [][]node.Node{
																{
																	&node.Text{
																		Value: " 4",
																	},
																	&node.Text{
																		Value: "four",
																	},
																},
															},
														},
													},
													{
														&node.Text{
															Value: " 5",
														},
														&node.Text{
															Value: "five",
														},
													},
												},
											},
										},
										{
											&node.Text{
												Value: " 6",
											},
											&node.Text{
												Value: "six",
											},
										},
									},
								},
							},
							{
								&node.Text{
									Value: " 7",
								},
								&node.Text{
									Value: "seven",
								},
							},
						},
					},
					&node.Paragraph{
						Children: []node.Inline{
							&node.Text{
								Value: "End",
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

			want := p.ParseDocument()
			if want == nil {
				t.Fatalf("ParseDocument() returned nil")
			}

			if want.Pretty(1) != tc.want.Pretty(1) {
				t.Errorf(
					"document \"%s\" is incorrect, from input `%s`\ngot:\n%s\nwant:\n%s",
					tc.name,
					tc.input,
					want.Pretty(1),
					tc.want.Pretty(1),
				)
			}
		})
	}
}
