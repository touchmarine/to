package template

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"strings"

	"github.com/touchmarine/to/node"
)

func Functions(tmpl *template.Template) template.FuncMap {
	return template.FuncMap{
		"log":              Log,
		"logf":             Logf,
		"error":            Error,
		"errorf":           Errorf,
		"dynamicTemplate":  MakeTemplateFunction(tmpl),
		"elementChildren":  ElementChildren,
		"trimSpacing":      TrimSpacing,
		"parseAttributes":  ParseAttributes,
		"nodeSetData":      NodeSetData,
		"attributesToHTML": AttributesToHTML,
	}
}

func Log(v ...interface{}) string {
	log.Print(v...)
	return ""
}

func Logf(format string, v ...interface{}) string {
	log.Printf(format, v...)
	return ""
}

func Error(text string) (string, error) {
	return "", errors.New(text)
}

func Errorf(format string, v ...interface{}) (string, error) {
	return "", fmt.Errorf(format, v...)
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

func ElementChildren(n *node.Node) []*node.Node {
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

func TrimSpacing(s string) string {
	return strings.Trim(s, " \t")
}

func NodeSetData(n *node.Node, key string, v interface{}) *node.Node {
	if n != nil {
		if n.Data == nil {
			n.Data = node.Data{}
		}
		n.Data[key] = v
	}
	return n
}

// AttributesToHTML returns a HTML-formatted string of attributes from the given
// attributes.
func AttributesToHTML(attrs map[string]string) template.HTMLAttr {
	var b strings.Builder

	var i int
	for name, value := range attrs {
		if i > 0 {
			b.WriteString(" ")
		}

		b.WriteString(name)
		if value != "" {
			b.WriteString(`="` + value + `"`)
		}

		i++
	}

	return template.HTMLAttr(b.String())
}
