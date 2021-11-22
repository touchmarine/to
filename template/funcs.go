package template

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"strings"

	"github.com/touchmarine/to/node"
)

// Funcs returns the set of Touch template functions.
func Funcs(tmpl *template.Template, global map[string]interface{}) template.FuncMap {
	return template.FuncMap{
		"log":              Log,
		"logf":             Logf,
		"error":            Error,
		"errorf":           Errorf,
		"dynamicTemplate":  MakeTemplateFunction(tmpl),
		"elementChildren":  ElementChildren,
		"trimSpacing":      TrimSpacing,
		"parseAttributes":  ParseAttributes,
		"setData":          NodeSetData,
		"attributesToHTML": AttributesToHTML,
		"global":           MakeGlobalMapFunction(global),
		"get":              Dot,
		"set":              Set,
		"setDefault":       SetDefault,
	}
}

// Log wraps log.Print.
func Log(v ...interface{}) string {
	log.Print(v...)
	return ""
}

// Logf wraps log.Printf.
func Logf(format string, v ...interface{}) string {
	log.Printf(format, v...)
	return ""
}

// Error returns a new error.
func Error(text string) (string, error) {
	return "", errors.New(text)
}

// Errorf returns a new formatted error.
func Errorf(format string, v ...interface{}) (string, error) {
	return "", fmt.Errorf(format, v...)
}

// MakeTemplateFunction returns a function that can be used like the default Go
// {{template}} function but supports variable template names.
//
// e.g. {{dynamicTemplate $c.Element $c}}
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

// ElementChildren returns a list of element children-children that represent an
// element and not a plain node (e.g. container).
func ElementChildren(n *node.Node) []*node.Node {
	var nodes []*node.Node
	for c := firstElement(n.FirstChild); c != nil; c = firstElement(c.NextSibling) {
		nodes = append(nodes, c)
	}
	return nodes
}

// firstElement returns the first node that represents an element and not a
// plain node (e.g. container).
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

// TrimSpacing trims spaces and tabs.
func TrimSpacing(s string) string {
	return strings.Trim(s, " \t")
}

// NodeSetData sets an entry in node.Data.
func NodeSetData(n *node.Node, key string, v interface{}) *node.Node {
	if n != nil {
		if n.Data == nil {
			n.Data = node.Data{}
		}
		n.Data[key] = v
	}
	return n
}
