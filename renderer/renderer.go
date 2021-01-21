package renderer

import (
	"fmt"
	"strconv"
	"strings"
	"to/node"
)

func HTML(nod interface{}, indent int) string {
	var b strings.Builder

	switch n := nod.(type) {
	case []node.Node:
		for _, c := range n {
			b.WriteString(HTML(c, indent))
		}

	case []node.Inline:
		for _, c := range n {
			b.WriteString(HTML(c, indent))
		}

	case *node.Document:
		b.WriteString(HTML(n.Children, indent))

	case *node.Paragraph:
		b.WriteString(indented("<p>", indent))
		b.WriteString(HTML(n.Lines, indent))
		b.WriteString(indented("</p>\n", indent))

	case node.Lines:
		for i, line := range n {
			b.WriteString(HTML(line, indent))

			// line break if not last
			if i != len(n)-1 {
				b.WriteString("<br>\n")
			}
		}

	case *node.Line:
		for _, c := range n.Children {
			b.WriteString(HTML(c, indent))
		}

	case *node.Text:
		b.WriteString(indented(n.Value, indent))

	case *node.Emphasis:
		b.WriteString("<em>")
		b.WriteString(HTML(n.Children, 0))
		b.WriteString("</em>")

	case *node.Strong:
		b.WriteString("<strong>")
		b.WriteString(HTML(n.Children, 0))
		b.WriteString("</strong>")

	case *node.Heading:
		strLevel := strconv.FormatUint(uint64(n.Level), 10)

		if n.Level < 7 {
			b.WriteString(indented("<h"+strLevel+">", indent))
		} else {
			b.WriteString(indented(`<div role="heading" aria-level="`+strLevel+`">`, indent))
		}

		if n.IsNumbered {
			for i, seqNum := range n.SeqNums {
				if i > 0 {
					b.WriteString(".")
				}

				b.WriteString(strconv.FormatUint(uint64(seqNum), 10))
			}
			b.WriteString(" ")
		}

		b.WriteString(HTML(n.Children, 0))

		if n.Level < 7 {
			b.WriteString(indented("</h"+strLevel+">\n", indent))
		} else {
			b.WriteString(indented("</div>\n", indent))
		}

	case *node.Link:
		b.WriteString(indented(`<a href="`+n.Destination+`">`, indent))
		b.WriteString(HTML(n.Children, 0))
		b.WriteString(indented("</a>", indent))

	case *node.CodeBlock:
		innerIndent := indent
		if n.Filename != "" {
			innerIndent++

			b.WriteString(indented("<div>\n", indent))
			b.WriteString(indented(n.Filename+"\n", innerIndent))
		}

		b.WriteString(indented("<pre><code>", innerIndent))
		b.WriteString(n.Body)
		b.WriteString(indented("</code></pre>\n", innerIndent))

		if n.Filename != "" {
			b.WriteString(indented("</div>\n", indent))
		}

	case *node.List:
		b.WriteString(indented("<ul>\n", indent))
		b.WriteString(HTML(n.ListItems, indent+1))
		b.WriteString(indented("</ul>\n", indent))

	case []*node.ListItem:
		for _, listItem := range n {
			b.WriteString(HTML(listItem, indent))
		}

	case *node.ListItem:
		b.WriteString(indented("<li>", indent))

		for _, c := range n.Children {
			switch m := c.(type) {
			case node.Block:
				if lines, ok := m.(node.Lines); ok && len(lines) == 1 {
					b.WriteString(HTML(c, 0))
					break
				}

				b.WriteString("\n")
				b.WriteString(HTML(c, indent+1))
				b.WriteString("\n")
				b.WriteString(indented("", indent))
			case node.Inline:
				b.WriteString(HTML(c, 0))
			default:
				panic(fmt.Sprintf("renderer.HTML: unsupported node type %T", m))
			}
		}

		b.WriteString("</li>\n")

	case *node.BlockQuote:
		b.WriteString("<blockquote>")
		for _, c := range n.Children {
			b.WriteString(HTML(c, 0))
		}
		b.WriteString("</blockquote>")

	default:
		panic(fmt.Sprintf("renderer.HTML: unsupported node type %T", n))
	}

	return b.String()
}

func indented(s string, indent int) string {
	return strings.Repeat("\t", indent) + s
}
