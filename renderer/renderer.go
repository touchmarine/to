package renderer

import (
	"fmt"
	"strconv"
	"strings"
	"to/node"
)

func HTML(nod node.Node, indent int) string {
	var b strings.Builder

	switch n := nod.(type) {
	case *node.Document:
		for _, c := range n.Children {
			b.WriteString(HTML(c, indent))
		}

	case *node.Paragraph:
		b.WriteString(indented("<p>", indent))

		for i, line := range n.Lines {
			if i > 0 {
				b.WriteString("<br>\n")
			}

			b.WriteString(HTML(line, 0))
		}

		b.WriteString(indented("</p>\n", indent))

	case *node.Line:
		for _, c := range n.Children {
			b.WriteString(HTML(c, indent))
		}

	case *node.Text:
		b.WriteString(indented(n.Value, indent))

	case *node.Emphasis:
		b.WriteString("<em>")
		for _, c := range n.Children {
			b.WriteString(HTML(c, 0))
		}
		b.WriteString("</em>")

	case *node.Strong:
		b.WriteString("<strong>")
		for _, c := range n.Children {
			b.WriteString(HTML(c, 0))
		}
		b.WriteString("</strong>")

	case *node.Heading:
		strLevel := strconv.Itoa(n.Level)

		if n.Level < 7 {
			b.WriteString(indented("<h"+strLevel+">", indent))
		} else {
			b.WriteString(indented(`<div role="heading" aria-level="`+strLevel+`">`, indent))
		}

		for _, c := range n.Children {
			b.WriteString(HTML(c, 0))
		}

		if n.Level < 7 {
			b.WriteString(indented("</h"+strLevel+">\n", indent))
		} else {
			b.WriteString(indented("</div>\n", indent))
		}

	case *node.Link:
		b.WriteString(indented(`<a href="https://www.nationalgeographic.com/animals/mammals/k/koala/">`, indent))

		for _, c := range n.Children {
			b.WriteString(HTML(c, 0))
		}

		b.WriteString(indented("</a>", indent))

	case *node.CodeBlock:
		innerIndent := indent
		if n.Filename != "" {
			innerIndent++

			b.WriteString(indented("<div>\n", indent))
			b.WriteString(indented(n.Filename+"\n", innerIndent))
		}

		b.WriteString(indented("<pre><code>\n", innerIndent))
		b.WriteString(n.Body)
		b.WriteString(indented("</code></pre>\n", innerIndent))

		if n.Filename != "" {
			b.WriteString(indented("</div>\n", indent))
		}

	case *node.List:
		b.WriteString(indented("<ul>\n", indent))

		for _, listItem := range n.ListItems {
			b.WriteString(indented("<li>\n", indent+1))

			for j, c := range listItem {
				if j > 0 {
					b.WriteString("\n")
				}

				b.WriteString(HTML(c, indent+2))

				// if not last
				if j != len(n.ListItems)-1 {
					b.WriteString("\n")
				}
			}

			b.WriteString(indented("</li>\n", indent+1))
		}

		b.WriteString(indented("</ul>\n", indent))

	default:
		panic(fmt.Sprintf("renderer.HTML: unexpected node type %T", n))
	}

	return b.String()
}

func indented(s string, indent int) string {
	return strings.Repeat("\t", indent) + s
}
