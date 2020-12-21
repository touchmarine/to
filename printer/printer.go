package printer

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"to/node"
)

const tab = ".   "

// do not add quotes
type unquoted string

// Pretty returns a prettified string representation of nod; used for testing
// and debugging.
func Pretty(nod interface{}, indent int) string {
	switch n := nod.(type) {
	case []node.Node:
		return prettyNodes(n, indent)

	case *node.Document:
		return element("Document", map[string]interface{}{
			"Children": n.Children,
		}, indent)

	case *node.Paragraph:
		return element("Paragraph", map[string]interface{}{
			"Lines": node.LinesToNodes(n.Lines),
		}, indent)

	case node.Lines:
		return indented("Lines", indent) + prettyNodes(node.LinesToNodes(n), indent)

	case *node.Line:
		return element("Line", map[string]interface{}{
			"Children": node.InlinesToNodes(n.Children),
		}, indent)

	case *node.Text:
		return indented(`"`+n.Value+`"`, indent)

	case *node.Emphasis:
		return element("Emphasis", map[string]interface{}{
			"Children": node.InlinesToNodes(n.Children),
		}, indent)

	case *node.Strong:
		return element("Strong", map[string]interface{}{
			"Children": node.InlinesToNodes(n.Children),
		}, indent)

	case *node.Heading:
		return element("Heading", map[string]interface{}{
			"Level":      unquoted(strconv.Itoa(n.Level)),
			"IsNumbered": unquoted(strconv.FormatBool(n.IsNumbered)),
			"Children":   node.InlinesToNodes(n.Children),
		}, indent)

	case *node.Link:
		return element("Link", map[string]interface{}{
			"Destination": n.Destination,
			"Children":    node.InlinesToNodes(n.Children),
		}, indent)

	case *node.CodeBlock:
		return element("CodeBlock", map[string]interface{}{
			"Language":    n.Language,
			"Filename":    n.Filename,
			"MetadataRaw": n.MetadataRaw,
			"Body":        n.Body,
		}, indent)

	case *node.List:
		return element("List", map[string]interface{}{
			"Type":      n.Type,
			"ListItems": node.ListItemsToNodes(n.ListItems),
		}, indent)

	case *node.ListItem:
		return element("ListItem", map[string]interface{}{
			"Children": n.Children,
		}, indent)

	default:
		panic(fmt.Sprintf("printer.Pretty: unexpected node type %T", n))
	}
}

// prettyNodes is like Pretty() but accepts a slice of nodes.
func prettyNodes(nodes []node.Node, indent int) string {
	if len(nodes) == 0 {
		return "[]"
	}

	var b strings.Builder
	b.WriteString("[\n")

	for _, node := range nodes {
		b.WriteString(Pretty(node, indent+1) + ",\n")
	}

	b.WriteString(indented("]", indent))

	return b.String()
}

// element returns a prettified string representation of an element and fields.
//
// It calls Pretty() if a slice of nodes is a field value.
func element(name string, fields map[string]interface{}, indent int) string {
	var b strings.Builder
	b.WriteString(indented(name+"{", indent))

	// sort map keys
	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// print fields (key-value pairs) in separate lines, indented
	for _, k := range keys {
		i := fields[k]
		b.WriteString(fmt.Sprintf("\n%s: ", indented(k, indent+1))) // print key

		switch v := i.(type) {
		case []node.Node:
			b.WriteString(Pretty(v, indent+1))
		case string:
			b.WriteString(`"` + v + `"`)
		case unquoted:
			b.WriteString(string(v))
		case fmt.Stringer:
			b.WriteString(v.String())
		default:
			panic(fmt.Sprintf("unsupported value type"))
		}

		b.WriteString(",")
	}

	// move into next line and indent closing '}' if fields are present
	if len(fields) > 0 {
		b.WriteString("\n" + indented("}", indent))
		return b.String()
	}

	b.WriteString("}")
	return b.String()
}

func indented(s string, indent int) string {
	return strings.Repeat(tab, indent) + s
}
