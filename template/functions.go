package template

import (
	"errors"
	"html/template"
	"log"
	"strings"

	"github.com/touchmarine/to/node"
)

func Functions(tmpl *template.Template) template.FuncMap {
	return template.FuncMap{
		"Logf":     Logf,
		"Template": MakeTemplateFunction(tmpl),
		"Children": Children,
	}
}

func Logf(format string, v ...interface{}) string {
	log.Printf(format, v...)
	return ""
}

func MakeTemplateFunction(tmpl *template.Template) func(name string, v ...interface{}) (template.HTML, error) {
	return func(name string, v ...interface{}) (template.HTML, error) {
		var arg interface{}
		switch len(v) {
		case 0:
		case 1:
			arg = v[0]
		default:
			return template.HTML(""), errors.New("multiple arguments are not supported")
		}

		var b strings.Builder
		if err := tmpl.ExecuteTemplate(&b, name, arg); err != nil {
			return template.HTML(""), err
		}
		return template.HTML(b.String()), nil
	}
}

func Children(n *node.Node) []*node.Node {
	var nodes []*node.Node
	for c := firstElement(n.FirstChild); c != nil; c = firstElement(c.NextSibling) {
		nodes = append(nodes, c)
	}
	return nodes
}

// firstElement returns the first node that represents an element and not a
// plain container.
func firstElement(n *node.Node) *node.Node {
	if n == nil {
		return nil
	}
	if n.Element != "" {
		return n
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if x := firstElement(c); x != nil {
			return x
		}
	}
	return nil
}
