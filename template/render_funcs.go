package template

import (
	"html/template"
	"io"
	"strings"

	"github.com/touchmarine/to/node"
)

// RenderFunctions returns template functions for rendering nodes.
func RenderFunctions(tmpl *template.Template, data map[string]interface{}) template.FuncMap {
	r := &renderer{tmpl, data}
	return r.functions()
}

type renderer struct {
	tmpl *template.Template
	data map[string]interface{}
}

func (r *renderer) functions() template.FuncMap {
	return template.FuncMap{
		"render":         r.renderFunc,
		"renderWithData": r.renderWithDataFunc,
	}
}

func (r *renderer) renderFunc(n *node.Node) template.HTML {
	return r.renderWithDataFunc(n, nil)
}

func (r *renderer) renderWithDataFunc(n *node.Node, data map[string]interface{}) template.HTML {
	var b strings.Builder
	r.render(&b, n, data)
	return template.HTML(b.String())
}

func (r *renderer) render(w io.Writer, n *node.Node, tmplData map[string]interface{}) {
	if n == nil {
		return
	}

	data := map[string]interface{}{}

	m := walkUntilNonContainer(n, func(container *node.Node) {
		fillData(data, container)
	})
	if m == nil {
		panic("template: nil non-container node")
	}

	fillData(data, m)
	for k, v := range r.data {
		data[k] = v
	}
	// data from renderWithData template function
	for k, v := range tmplData {
		data[k] = v
	}

	if err := r.tmpl.ExecuteTemplate(w, m.Element, data); err != nil {
		panic(err)
	}
}

// walkUntilNonContainer walks the given node and its descendants until it
// encounters a non-container node and calls the given function for every
// container it encounters in the process. It returns the first non-container
// node, never nil.
func walkUntilNonContainer(n *node.Node, fn func(*node.Node)) *node.Node {
	if n.Type == node.TypeContainer {
		fn(n)
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			return walkUntilNonContainer(c, fn)
		}
	}
	return n
}

func fillData(data map[string]interface{}, n *node.Node) {
	data["Self"] = n
}
