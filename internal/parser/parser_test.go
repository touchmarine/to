package parser_test

import (
	"strconv"
	"testing"
	"to/internal/node"
	"to/internal/parser"
	"to/internal/printer"
	"to/internal/token"
)

func TestParse_Parse(t *testing.T) {
	cases := []struct {
		tokens []tl
		blocks []node.Block
	}{
		{
			[]tl{
				{token.GT, ">"},
				{token.VLINE, "|"},
				{token.GT, ">"},
				{token.TEXT, "a"},
				{token.LINEFEED, "\n"},
				{token.TEXT, "b"},
				{token.LINEFEED, "\n"},
				{token.VLINE, "|"},
				{token.TEXT, "c"},
			},
			[]node.Block{
				&node.Blockquote{
					[]node.Block{
						&node.Paragraph{
							[]node.Block{
								&node.Blockquote{
									[]node.Block{
										&node.Line{
											[]node.Inline{
												node.Text("a"),
											},
										},
									},
								},
							},
						},
					},
				},
				&node.Line{
					[]node.Inline{
						node.Text("b"),
					},
				},
				&node.Paragraph{
					[]node.Block{
						&node.Line{
							[]node.Inline{
								node.Text("c"),
							},
						},
					},
				},
			},
		},
		{
			[]tl{
				{token.VLINE, "|"},
				{token.TEXT, "a"},
				{token.LINEFEED, "\n"},
				{token.VLINE, "|"},
				{token.GT, ">"},
				{token.TEXT, "b"},
				{token.LINEFEED, "\n"},
				{token.VLINE, "|"},
				{token.GT, ">"},
				{token.VLINE, "|"},
				{token.TEXT, "c"},
			},
			[]node.Block{
				&node.Paragraph{
					[]node.Block{
						&node.Line{
							[]node.Inline{
								node.Text("a"),
							},
						},
						&node.Blockquote{
							[]node.Block{
								&node.Line{
									[]node.Inline{
										node.Text("b"),
									},
								},
								&node.Paragraph{
									[]node.Block{
										&node.Line{
											[]node.Inline{
												node.Text("c"),
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
			[]tl{
				{token.VLINE, "|"},
				{token.TEXT, "a"},
				{token.LINEFEED, "\n"},
				{token.VLINE, "|"},
				{token.GT, ">"},
				{token.LINEFEED, "\n"},
				{token.VLINE, "|"},
				{token.GT, ">"},
				{token.VLINE, "|"},
				{token.TEXT, "c"},
				{token.LINEFEED, "\n"},
				{token.TEXT, "d"},
			},
			[]node.Block{
				&node.Paragraph{
					[]node.Block{
						&node.Line{
							[]node.Inline{
								node.Text("a"),
							},
						},
						&node.Blockquote{
							[]node.Block{
								&node.Paragraph{
									[]node.Block{
										&node.Line{
											[]node.Inline{
												node.Text("c"),
											},
										},
									},
								},
							},
						},
					},
				},
				&node.Line{
					[]node.Inline{
						node.Text("d"),
					},
				},
			},
		},
		{
			[]tl{
				{token.VLINE, "|"},
				{token.GT, ">"},
				{token.LINEFEED, "\n"},
				{token.VLINE, "|"},
				{token.VLINE, "|"},
			},
			[]node.Block{
				&node.Paragraph{
					[]node.Block{
						&node.Blockquote{},
						&node.Paragraph{},
					},
				},
			},
		},
		{
			[]tl{
				{token.GT, ">"},
				{token.VLINE, "|"},
				{token.GT, ">"},
				{token.VLINE, "|"},
				{token.LINEFEED, "\n"},
				{token.GT, ">"},
				{token.VLINE, "|"},
				{token.GT, ">"},
				{token.VLINE, "|"},
				{token.TEXT, "a"},
				{token.LINEFEED, "\n"},
				{token.GT, ">"},
				{token.VLINE, "|"},
				{token.GT, ">"},
				{token.VLINE, "|"},
				{token.GT, ">"},
				{token.LINEFEED, "\n"},
				{token.GT, ">"},
				{token.TEXT, "b"},
			},
			[]node.Block{
				&node.Blockquote{
					[]node.Block{
						&node.Paragraph{
							[]node.Block{
								&node.Blockquote{
									[]node.Block{
										&node.Paragraph{
											[]node.Block{
												&node.Line{
													[]node.Inline{
														node.Text("a"),
													},
												},
												&node.Blockquote{},
											},
										},
									},
								},
							},
						},
						&node.Line{
							[]node.Inline{
								node.Text("b"),
							},
						},
					},
				},
			},
		},
		{
			// pyramid
			// a
			// |b
			// |>c
			// |>|d
			// |>e
			// |f
			// g
			[]tl{
				{token.TEXT, "a"},
				{token.LINEFEED, "\n"},
				{token.VLINE, "|"},
				{token.TEXT, "b"},
				{token.LINEFEED, "\n"},
				{token.VLINE, "|"},
				{token.GT, ">"},
				{token.TEXT, "c"},
				{token.LINEFEED, "\n"},
				{token.VLINE, "|"},
				{token.GT, ">"},
				{token.VLINE, "|"},
				{token.TEXT, "d"},
				{token.LINEFEED, "\n"},
				{token.VLINE, "|"},
				{token.GT, ">"},
				{token.TEXT, "e"},
				{token.LINEFEED, "\n"},
				{token.VLINE, "|"},
				{token.TEXT, "f"},
				{token.LINEFEED, "\n"},
				{token.TEXT, "g"},
			},
			[]node.Block{
				&node.Line{[]node.Inline{node.Text("a")}},
				&node.Paragraph{
					[]node.Block{
						&node.Line{[]node.Inline{node.Text("b")}},
						&node.Blockquote{
							[]node.Block{
								&node.Line{[]node.Inline{node.Text("c")}},
								&node.Paragraph{
									[]node.Block{
										&node.Line{[]node.Inline{node.Text("d")}},
									},
								},
								&node.Line{[]node.Inline{node.Text("e")}},
							},
						},
						&node.Line{[]node.Inline{node.Text("f")}},
					},
				},
				&node.Line{[]node.Inline{node.Text("g")}},
			},
		},
	}

	for _, c := range cases {
		var name string
		for _, pair := range c.tokens {
			name += pair.lit
		}

		t.Run(literal(name), func(t *testing.T) {
			test(t, c.tokens, c.blocks)
		})
	}
}

func TestParagraph(t *testing.T) {
	cases := []struct {
		tokens []tl
		blocks []node.Block
	}{
		{
			[]tl{{token.VLINE, "|"}},
			[]node.Block{
				&node.Paragraph{},
			},
		},
		{
			[]tl{{token.VLINE, "|"}, {token.VLINE, "|"}},
			[]node.Block{
				&node.Paragraph{
					[]node.Block{
						&node.Paragraph{},
					},
				},
			},
		},
		{
			[]tl{
				{token.VLINE, "|"},
				{token.VLINE, "|"},
				{token.TEXT, "a"},
			},
			[]node.Block{
				&node.Paragraph{
					[]node.Block{
						&node.Paragraph{
							[]node.Block{
								&node.Line{
									[]node.Inline{
										node.Text("a"),
									},
								},
							},
						},
					},
				},
			},
		},
		{
			[]tl{
				{token.VLINE, "|"},
				{token.LINEFEED, "\n"},
				{token.VLINE, "|"},
				{token.TEXT, "a"},
			},
			[]node.Block{
				&node.Paragraph{
					[]node.Block{
						&node.Line{
							[]node.Inline{
								node.Text("a"),
							},
						},
					},
				},
			},
		},
		{
			[]tl{
				{token.VLINE, "|"},
				{token.TEXT, "a"},
				{token.LINEFEED, "\n"},
				{token.VLINE, "|"},
				{token.TEXT, "b"},
			},
			[]node.Block{
				&node.Paragraph{
					[]node.Block{
						&node.Line{
							[]node.Inline{
								node.Text("a"),
							},
						},
						&node.Line{
							[]node.Inline{
								node.Text("b"),
							},
						},
					},
				},
			},
		},
		{
			[]tl{
				{token.VLINE, "|"},
				{token.TEXT, "a"},
				{token.LINEFEED, "\n"},
				{token.VLINE, "|"},
				{token.VLINE, "|"},
				{token.TEXT, "b"},
			},
			[]node.Block{
				&node.Paragraph{
					[]node.Block{
						&node.Line{
							[]node.Inline{
								node.Text("a"),
							},
						},
						&node.Paragraph{
							[]node.Block{
								&node.Line{
									[]node.Inline{
										node.Text("b"),
									},
								},
							},
						},
					},
				},
			},
		},
		{
			[]tl{
				{token.VLINE, "|"},
				{token.TEXT, "a"},
				{token.LINEFEED, "\n"},
				{token.VLINE, "|"},
				{token.VLINE, "|"},
				{token.TEXT, "b"},
				{token.LINEFEED, "\n"},
				{token.VLINE, "|"},
				{token.VLINE, "|"},
				{token.VLINE, "|"},
				{token.TEXT, "c"},
			},
			[]node.Block{
				&node.Paragraph{
					[]node.Block{
						&node.Line{
							[]node.Inline{
								node.Text("a"),
							},
						},
						&node.Paragraph{
							[]node.Block{
								&node.Line{
									[]node.Inline{
										node.Text("b"),
									},
								},
								&node.Paragraph{
									[]node.Block{
										&node.Line{
											[]node.Inline{
												node.Text("c"),
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
			[]tl{
				{token.VLINE, "|"},
				{token.TEXT, "a"},
				{token.LINEFEED, "\n"},
				{token.TEXT, "b"},
			},
			[]node.Block{
				&node.Paragraph{
					[]node.Block{
						&node.Line{
							[]node.Inline{
								node.Text("a"),
							},
						},
					},
				},
				&node.Line{
					[]node.Inline{
						node.Text("b"),
					},
				},
			},
		},
		{
			[]tl{
				{token.VLINE, "|"},
				{token.VLINE, "|"},
				{token.TEXT, "a"},
				{token.LINEFEED, "\n"},
				{token.TEXT, "b"},
			},
			[]node.Block{
				&node.Paragraph{
					[]node.Block{
						&node.Paragraph{
							[]node.Block{
								&node.Line{
									[]node.Inline{
										node.Text("a"),
									},
								},
							},
						},
					},
				},
				&node.Line{
					[]node.Inline{
						node.Text("b"),
					},
				},
			},
		},
		{
			[]tl{
				{token.VLINE, "|"},
				{token.VLINE, "|"},
				{token.TEXT, "a"},
				{token.LINEFEED, "\n"},
				{token.VLINE, "|"},
				{token.TEXT, "b"},
			},
			[]node.Block{
				&node.Paragraph{
					[]node.Block{
						&node.Paragraph{
							[]node.Block{
								&node.Line{
									[]node.Inline{
										node.Text("a"),
									},
								},
							},
						},
						&node.Line{
							[]node.Inline{
								node.Text("b"),
							},
						},
					},
				},
			},
		},
		{
			[]tl{
				{token.VLINE, "|"},
				{token.VLINE, "|"},
				{token.VLINE, "|"},
				{token.TEXT, "a"},
				{token.LINEFEED, "\n"},
				{token.VLINE, "|"},
				{token.VLINE, "|"},
				{token.TEXT, "b"},
				{token.LINEFEED, "\n"},
				{token.VLINE, "|"},
				{token.TEXT, "c"},
			},
			[]node.Block{
				&node.Paragraph{
					[]node.Block{
						&node.Paragraph{
							[]node.Block{
								&node.Paragraph{
									[]node.Block{
										&node.Line{
											[]node.Inline{
												node.Text("a"),
											},
										},
									},
								},
								&node.Line{
									[]node.Inline{
										node.Text("b"),
									},
								},
							},
						},
						&node.Line{
							[]node.Inline{
								node.Text("c"),
							},
						},
					},
				},
			},
		},
	}

	for _, c := range cases {
		var name string
		for _, pair := range c.tokens {
			name += pair.lit
		}

		t.Run(literal(name), func(t *testing.T) {
			test(t, c.tokens, c.blocks)
		})
	}
}

func TestBlockquote(t *testing.T) {
	cases := []struct {
		tokens []tl
		blocks []node.Block
	}{
		{
			[]tl{{token.GT, ">"}},
			[]node.Block{
				&node.Blockquote{},
			},
		},
		{
			[]tl{{token.GT, ">"}, {token.GT, ">"}},
			[]node.Block{
				&node.Blockquote{
					[]node.Block{
						&node.Blockquote{},
					},
				},
			},
		},
		{
			[]tl{
				{token.GT, ">"},
				{token.GT, ">"},
				{token.TEXT, "a"},
			},
			[]node.Block{
				&node.Blockquote{
					[]node.Block{
						&node.Blockquote{
							[]node.Block{
								&node.Line{
									[]node.Inline{
										node.Text("a"),
									},
								},
							},
						},
					},
				},
			},
		},
		{
			[]tl{{token.GT, ">"}, {token.VLINE, "|"}},
			[]node.Block{
				&node.Blockquote{
					[]node.Block{
						&node.Paragraph{},
					},
				},
			},
		},
	}

	for _, c := range cases {
		var name string
		for _, pair := range c.tokens {
			name += pair.lit
		}

		t.Run(literal(name), func(t *testing.T) {
			test(t, c.tokens, c.blocks)
		})
	}
}

func TestListItem(t *testing.T) {
	cases := []struct {
		tokens []tl
		blocks []node.Block
	}{
		{
			[]tl{{token.HYPEN, "-"}},
			[]node.Block{
				&node.ListItem{},
			},
		},
		{
			[]tl{
				{token.HYPEN, "-"},
				{token.TEXT, "a"},
			},
			[]node.Block{
				&node.ListItem{
					[]node.Block{
						&node.Line{
							[]node.Inline{
								node.Text("a"),
							},
						},
					},
				},
			},
		},
		{
			[]tl{
				{token.HYPEN, "-"},
				{token.HYPEN, "-"},
				{token.TEXT, "a"},
			},
			[]node.Block{
				&node.ListItem{
					[]node.Block{
						&node.ListItem{
							[]node.Block{
								&node.Line{
									[]node.Inline{
										node.Text("a"),
									},
								},
							},
						},
					},
				},
			},
		},
		{
			[]tl{
				{token.HYPEN, "-"},
				{token.TEXT, "a"},
				{token.LINEFEED, "\n"},
				{token.HYPEN, "-"},
				{token.TEXT, "b"},
			},
			[]node.Block{
				&node.ListItem{
					[]node.Block{
						&node.Line{
							[]node.Inline{
								node.Text("a"),
							},
						},
					},
				},
				&node.ListItem{
					[]node.Block{
						&node.Line{
							[]node.Inline{
								node.Text("b"),
							},
						},
					},
				},
			},
		},
		{
			[]tl{
				{token.HYPEN, "-"},
				{token.TEXT, "a"},
				{token.LINEFEED, "\n"},
				{token.TEXT, "b"},
			},
			[]node.Block{
				&node.ListItem{
					[]node.Block{
						&node.Line{
							[]node.Inline{
								node.Text("a"),
							},
						},
					},
				},
				&node.Line{[]node.Inline{node.Text("b")}},
			},
		},
		{
			[]tl{
				{token.INDENT, " "},
				{token.HYPEN, "-"},
				{token.TEXT, "a"},
				{token.LINEFEED, "\n"},
				{token.INDENT, " "},
				{token.TEXT, "b"},
			},
			[]node.Block{
				&node.ListItem{
					[]node.Block{
						&node.Line{
							[]node.Inline{
								node.Text("a"),
							},
						},
					},
				},
				&node.Line{[]node.Inline{node.Text("b")}},
			},
		},
		{
			[]tl{
				{token.HYPEN, "-"},
				{token.TEXT, "a"},
				{token.LINEFEED, "\n"},
				{token.INDENT, " "},
				{token.TEXT, "b"},
			},
			[]node.Block{
				&node.ListItem{
					[]node.Block{
						&node.Line{
							[]node.Inline{
								node.Text("a"),
							},
						},
						&node.Line{
							[]node.Inline{
								node.Text("b"),
							},
						},
					},
				},
			},
		},
		{
			[]tl{
				{token.HYPEN, "-"},
				{token.TEXT, "a"},
				{token.LINEFEED, "\n"},
				{token.INDENT, "  "},
				{token.TEXT, "b"},
			},
			[]node.Block{
				&node.ListItem{
					[]node.Block{
						&node.Line{
							[]node.Inline{
								node.Text("a"),
							},
						},
						&node.Line{
							[]node.Inline{
								node.Text("b"),
							},
						},
					},
				},
			},
		},
		{
			[]tl{
				{token.INDENT, " "},
				{token.HYPEN, "-"},
				{token.TEXT, "a"},
				{token.LINEFEED, "\n"},
				{token.INDENT, "  "},
				{token.TEXT, "b"},
			},
			[]node.Block{
				&node.ListItem{
					[]node.Block{
						&node.Line{
							[]node.Inline{
								node.Text("a"),
							},
						},
						&node.Line{
							[]node.Inline{
								node.Text("b"),
							},
						},
					},
				},
			},
		},
		{
			[]tl{
				{token.HYPEN, "-"},
				{token.TEXT, "a"},
				{token.LINEFEED, "\n"},
				{token.INDENT, " "},
				{token.HYPEN, "-"},
				{token.TEXT, "b"},
			},
			[]node.Block{
				&node.ListItem{
					[]node.Block{
						&node.Line{
							[]node.Inline{
								node.Text("a"),
							},
						},
						&node.ListItem{
							[]node.Block{
								&node.Line{[]node.Inline{node.Text("b")}},
							},
						},
					},
				},
			},
		},
		{
			[]tl{
				{token.VLINE, "|"},
				{token.INDENT, " "},
				{token.HYPEN, "-"},
				{token.INDENT, " "},
				{token.TEXT, "a"},
				{token.LINEFEED, "\n"},
				{token.VLINE, "|"},
				{token.INDENT, "  "},
				{token.TEXT, "b"},
			},
			[]node.Block{
				&node.Paragraph{
					[]node.Block{
						&node.ListItem{
							[]node.Block{
								&node.Line{
									[]node.Inline{
										node.Text("a"),
									},
								},
								&node.Line{
									[]node.Inline{
										node.Text("b"),
									},
								},
							},
						},
					},
				},
			},
		},
		{
			[]tl{
				{token.HYPEN, "-"},
				{token.INDENT, " "},
				{token.TEXT, "a"},
				{token.LINEFEED, "\n"},
				{token.INDENT, " "},
				{token.VLINE, "|"},
				{token.INDENT, " "},
				{token.TEXT, "b"},
			},
			[]node.Block{
				&node.ListItem{
					[]node.Block{
						&node.Line{[]node.Inline{node.Text("a")}},
						&node.Paragraph{
							[]node.Block{
								&node.Line{[]node.Inline{node.Text("b")}},
							},
						},
					},
				},
			},
		},
	}

	for _, c := range cases {
		var name string
		for _, pair := range c.tokens {
			name += pair.lit
		}

		t.Run(literal(name), func(t *testing.T) {
			test(t, c.tokens, c.blocks)
		})
	}
}

func TestCodeBlock(t *testing.T) {
	cases := []struct {
		tokens []tl
		blocks []node.Block
	}{
		{
			[]tl{
				{token.GRAVEACCENTS, "``"},
				{token.LINEFEED, "\n"},
				{token.GRAVEACCENTS, "``"},
			},
			[]node.Block{
				&node.CodeBlock{
					Head: "",
					Body: "",
				},
			},
		},
		{
			// scanner would return this as one token "````"
			[]tl{
				{token.GRAVEACCENTS, "``"},
				{token.GRAVEACCENTS, "``"},
			},
			[]node.Block{
				&node.CodeBlock{
					Head: "",
					Body: "",
				},
			},
		},
		{
			[]tl{
				{token.GRAVEACCENTS, "``"},
			},
			[]node.Block{
				&node.CodeBlock{
					Head: "",
					Body: "",
				},
			},
		},
		{
			[]tl{
				{token.GRAVEACCENTS, "``"},
				{token.TEXT, "a"},
			},
			[]node.Block{
				&node.CodeBlock{
					Head: "a",
					Body: "",
				},
			},
		},
		{
			[]tl{
				{token.GRAVEACCENTS, "``"},
				{token.TEXT, "a"},
				{token.LINEFEED, "\n"},
				{token.TEXT, "b"},
				{token.GRAVEACCENTS, "``"},
			},
			[]node.Block{
				&node.CodeBlock{
					Head: "a",
					Body: "b",
				},
			},
		},
		{
			[]tl{
				{token.GRAVEACCENTS, "``"},
				{token.TEXT, "a"},
				{token.LINEFEED, "\n"},
				{token.TEXT, "b"},
				{token.LINEFEED, "\n"},
				{token.TEXT, "c"},
				{token.LINEFEED, "\n"},
				{token.TEXT, "d"},
				{token.LINEFEED, "\n"},
				{token.GRAVEACCENTS, "``"},
			},
			[]node.Block{
				&node.CodeBlock{
					Head: "a",
					Body: "b\nc\nd\n",
				},
			},
		},
		{
			[]tl{
				{token.VLINE, "|"},
				{token.GRAVEACCENTS, "``"},
				{token.LINEFEED, "\n"},
				{token.VLINE, "|"},
				{token.GRAVEACCENTS, "``"},
			},
			[]node.Block{
				&node.Paragraph{
					[]node.Block{
						&node.CodeBlock{
							Head: "",
							Body: "",
						},
					},
				},
			},
		},
		{
			[]tl{
				{token.VLINE, "|"},
				{token.GRAVEACCENTS, "``"},
				{token.LINEFEED, "\n"},
				{token.GRAVEACCENTS, "``"},
			},
			[]node.Block{
				&node.Paragraph{
					[]node.Block{
						&node.CodeBlock{
							Head: "",
							Body: "",
						},
					},
				},
				&node.CodeBlock{
					Head: "",
					Body: "",
				},
			},
		},
		{
			[]tl{
				{token.HYPEN, "-"},
				{token.GRAVEACCENTS, "``"},
				{token.LINEFEED, "\n"},
				{token.GRAVEACCENTS, "``"},
			},
			[]node.Block{
				&node.ListItem{
					[]node.Block{
						&node.CodeBlock{
							Head: "",
							Body: "",
						},
					},
				},
				&node.CodeBlock{
					Head: "",
					Body: "",
				},
			},
		},
		{
			[]tl{
				{token.HYPEN, "-"},
				{token.GRAVEACCENTS, "``"},
				{token.LINEFEED, "\n"},
				{token.INDENT, " "},
				{token.TEXT, "a"},
				{token.GRAVEACCENTS, "``"},
			},
			[]node.Block{
				&node.ListItem{
					[]node.Block{
						&node.CodeBlock{
							Head: "",
							Body: "a",
						},
					},
				},
			},
		},
		{
			[]tl{
				{token.GRAVEACCENTS, "```"},
				{token.TEXT, "a"},
				{token.LINEFEED, "\n"},
				{token.GRAVEACCENTS, "``"},
				{token.TEXT, "b"},
				{token.LINEFEED, "\n"},
				{token.TEXT, "c"},
				{token.LINEFEED, "\n"},
				{token.GRAVEACCENTS, "``"},
				{token.LINEFEED, "\n"},
				{token.GRAVEACCENTS, "```"},
			},
			[]node.Block{
				&node.CodeBlock{
					Head: "a",
					Body: "``b\nc\n``\n",
				},
			},
		},
		{
			[]tl{
				{token.VLINE, "|"},
				{token.TEXT, "a"},
				{token.LINEFEED, "\n"},
				{token.VLINE, "|"},
				{token.GRAVEACCENTS, "``"},
				{token.TEXT, "metadata"},
				{token.LINEFEED, "\n"},
				{token.VLINE, "|"},
				{token.TEXT, "b"},
				{token.LINEFEED, "\n"},
				{token.VLINE, "|"},
				{token.GRAVEACCENTS, "``"},
				{token.TEXT, "c"},
			},
			[]node.Block{
				&node.Paragraph{
					[]node.Block{
						&node.Line{
							[]node.Inline{
								node.Text("a"),
							},
						},
						&node.CodeBlock{
							Head: "metadata",
							Body: "b\n",
						},
						&node.Line{
							[]node.Inline{
								node.Text("c"),
							},
						},
					},
				},
			},
		},
	}

	for _, c := range cases {
		var name string
		for _, pair := range c.tokens {
			name += pair.lit
		}

		t.Run(literal(name), func(t *testing.T) {
			test(t, c.tokens, c.blocks)
		})
	}
}

func TestLine(t *testing.T) {
	cases := []struct {
		tokens []tl
		blocks []node.Block
	}{
		{
			[]tl{{token.TEXT, "a"}},
			[]node.Block{
				&node.Line{
					[]node.Inline{node.Text("a")},
				},
			},
		},
		{
			[]tl{
				{token.TEXT, "a"},
				{token.LINEFEED, "\n"},
				{token.TEXT, "b"},
			},
			[]node.Block{
				&node.Line{
					[]node.Inline{node.Text("a")},
				},
				&node.Line{
					[]node.Inline{node.Text("b")},
				},
			},
		},
	}

	for _, c := range cases {
		var name string
		for _, pair := range c.tokens {
			name += pair.lit
		}

		t.Run(literal(name), func(t *testing.T) {
			test(t, c.tokens, c.blocks)
		})
	}
}

func TestEmphasis(t *testing.T) {
	cases := []struct {
		tokens []tl
		blocks []node.Block
	}{
		{
			[]tl{{token.LOWLINES, "__"}},
			[]node.Block{
				&node.Line{
					[]node.Inline{&node.Emphasis{}},
				},
			},
		},
		{
			[]tl{
				{token.LOWLINES, "__"},
				{token.TEXT, "a"},
			},
			[]node.Block{
				&node.Line{
					[]node.Inline{
						&node.Emphasis{
							[]node.Inline{
								node.Text("a"),
							},
						},
					},
				},
			},
		},
		{
			[]tl{
				{token.LOWLINES, "__"},
				{token.TEXT, "a"},
				{token.LOWLINES, "__"},
			},
			[]node.Block{
				&node.Line{
					[]node.Inline{
						&node.Emphasis{
							[]node.Inline{
								node.Text("a"),
							},
						},
					},
				},
			},
		},
		{
			[]tl{
				{token.LOWLINES, "__"},
				{token.TEXT, "a"},
				{token.LINEFEED, "\n"},
			},
			[]node.Block{
				&node.Line{
					[]node.Inline{
						&node.Emphasis{
							[]node.Inline{
								node.Text("a"),
							},
						},
					},
				},
			},
		},
		{
			[]tl{
				{token.TEXT, "a"},
				{token.LOWLINES, "__"},
			},
			[]node.Block{
				&node.Line{
					[]node.Inline{
						node.Text("a"),
						&node.Emphasis{},
					},
				},
			},
		},
		{
			[]tl{
				{token.VLINE, "|"},
				{token.LOWLINES, "__"},
			},
			[]node.Block{
				&node.Paragraph{
					[]node.Block{
						&node.Line{
							[]node.Inline{&node.Emphasis{}},
						},
					},
				},
			},
		},
	}

	for _, c := range cases {
		var name string
		for _, pair := range c.tokens {
			name += pair.lit
		}

		t.Run(literal(name), func(t *testing.T) {
			test(t, c.tokens, c.blocks)
		})
	}
}

func TestCode(t *testing.T) {
	cases := []struct {
		tokens []tl
		blocks []node.Block
	}{
		{
			[]tl{{token.GAP, "`*"}},
			[]node.Block{
				&node.Line{
					[]node.Inline{&node.Code{}},
				},
			},
		},
		{
			[]tl{
				{token.GAP, "`*"},
				{token.TEXT, "a"},
			},
			[]node.Block{
				&node.Line{
					[]node.Inline{
						&node.Code{"a"},
					},
				},
			},
		},
		{
			[]tl{
				{token.GAP, "`*"},
				{token.TEXT, "a"},
				{token.PAG, "*`"},
			},
			[]node.Block{
				&node.Line{
					[]node.Inline{
						&node.Code{"a"},
					},
				},
			},
		},
		{
			[]tl{
				{token.GAP, "`("},
				{token.TEXT, "a"},
				{token.PAG, ")`"},
			},
			[]node.Block{
				&node.Line{
					[]node.Inline{
						&node.Code{"a"},
					},
				},
			},
		},
		{
			[]tl{
				{token.GAP, "`*"},
				{token.TEXT, "a"},
				{token.LINEFEED, "\n"},
			},
			[]node.Block{
				&node.Line{
					[]node.Inline{
						&node.Code{"a"},
					},
				},
			},
		},
		{
			[]tl{
				{token.TEXT, "a"},
				{token.GAP, "`*"},
			},
			[]node.Block{
				&node.Line{
					[]node.Inline{
						node.Text("a"),
						&node.Code{},
					},
				},
			},
		},
		{
			[]tl{
				{token.TEXT, "a"},
				{token.GRAVEACCENTS, "``"},
			},
			[]node.Block{
				&node.Line{
					[]node.Inline{
						node.Text("a"),
						&node.Code{},
					},
				},
			},
		},
		{
			[]tl{
				{token.VLINE, "|"},
				{token.GAP, "`*"},
			},
			[]node.Block{
				&node.Paragraph{
					[]node.Block{
						&node.Line{
							[]node.Inline{&node.Code{}},
						},
					},
				},
			},
		},
	}

	for _, c := range cases {
		var name string
		for _, pair := range c.tokens {
			name += pair.lit
		}

		t.Run(literal(name), func(t *testing.T) {
			test(t, c.tokens, c.blocks)
		})
	}
}

func TestNoCode(t *testing.T) {
	cases := []struct {
		tokens []tl
		blocks []node.Block
	}{
		{
			[]tl{{token.GRAVEACCENTS, "``"}},
			[]node.Block{
				&node.CodeBlock{},
			},
		},
	}

	for _, c := range cases {
		var name string
		for _, pair := range c.tokens {
			name += pair.lit
		}

		t.Run(literal(name), func(t *testing.T) {
			test(t, c.tokens, c.blocks)
		})
	}
}

// token-literal pair struct
type tl struct {
	tok token.Token
	lit string
}

func test(t *testing.T, tokens []tl, wBlocks []node.Block) {
	t.Helper()

	s := &scanner{tokenLiterals: tokens}
	p := parser.New(s)
	blocks := p.Parse()

	got := printer.Print(blocks)
	want := printer.Print(wBlocks)

	if got != want {
		t.Errorf("\ngot\n%s\nwant\n%s", got, want)
	}
}

type scanner struct {
	i             int  // index
	tokenLiterals []tl // token-literal pairs
}

func (s *scanner) Scan() (token.Token, string) {
	if s.i >= len(s.tokenLiterals) {
		return token.EOF, ""
	}

	p := s.tokenLiterals[s.i]
	s.i++
	return p.tok, p.lit
}

func literal(s string) string {
	q := strconv.Quote(s)
	return q[1 : len(q)-1]
}
