package template

import (
	"fmt"
	"github.com/touchmarine/to/node"
	"html/template"
	"io"
	"strconv"
	"strings"
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

func (r *renderer) renderFunc(v interface{}) template.HTML {
	return r.renderWithDataFunc(v, nil)
}

func (r *renderer) renderWithDataFunc(v interface{}, data map[string]interface{}) template.HTML {
	var b strings.Builder

	switch n := v.(type) {
	case []node.Node:
		r.render(&b, n, data)
	case node.Node:
		r.render(&b, []node.Node{n}, data)
	default:
		panic(fmt.Sprintf("render: unexpected node %T", v))
	}

	return template.HTML(b.String())
}

func (r *renderer) render(out io.Writer, nodes []node.Node, tmplData map[string]interface{}) {
	for i, n := range nodes {
		if i > 0 {
			out.Write([]byte("\n"))
		}

		switch n.(type) {
		case node.BlockChildren, node.InlineChildren, node.Content,
			node.Lines, node.Composited, node.Ranked, node.Boxed:
		default:
			panic(fmt.Sprintf("render: unexpected node %T", n))
		}

		data := make(map[string]interface{})

		if boxed, isBoxed := n.(node.Boxed); isBoxed {
			unboxed := boxed.Unbox()
			if unboxed == nil {
				continue
			}

			fillData(data, n)

			n = unboxed
		}

		fillData(data, n)

		for k, v := range r.data {
			data[k] = v
		}

		// data from renderWithData template function
		for k, v := range tmplData {
			data[k] = v
		}

		name := n.Node()
		if err := r.tmpl.ExecuteTemplate(out, name, data); err != nil {
			panic(err)
		}
	}
}

func fillData(data map[string]interface{}, n node.Node) {
	data["Self"] = n
	data["TextContent"] = node.ExtractText(n)

	if m, ok := n.(node.BlockChildren); ok {
		data["BlockChildren"] = m.BlockChildren()
	}

	if m, ok := n.(node.InlineChildren); ok {
		data["InlineChildren"] = m.InlineChildren()
	}

	if m, ok := n.(node.Content); ok {
		data["Content"] = string(m.Content())
	}

	if m, ok := n.(node.Lines); ok {
		lines := btosSlice(m.Lines())

		data["Lines"] = lines
		data["Text"] = strings.Join(lines, "\n")
	}

	if m, ok := n.(node.Composited); ok {
		primary, secondary := make(map[string]interface{}), make(map[string]interface{})
		fillData(primary, m.Primary())
		fillData(secondary, m.Secondary())

		data["PrimaryElement"] = primary
		data["SecondaryElement"] = secondary
	}

	if m, ok := n.(*node.Sticky); ok {
		sticky, target := make(map[string]interface{}), make(map[string]interface{})
		fillData(sticky, m.Sticky())
		fillData(target, m.Target())

		data["StickyElement"] = sticky
		data["TargetElement"] = target
	}

	if m, ok := n.(node.Ranked); ok {
		data["Rank"] = strconv.FormatUint(uint64(m.Rank()), 10)
	}

	if _, ok := n.(node.Boxed); ok {
		switch k := n.(type) {
		case *node.SequentialNumberBox:
			data["SequentialNumbers"] = k.SequentialNumbers
			data["SequentialNumber"] = k.SequentialNumber()
		default:
			panic(fmt.Sprintf("render: unexpected Boxed node %T", n))
		}
	}
}

// btosSlice converts [][]byte to []string.
func btosSlice(p [][]byte) []string {
	var lines []string
	for _, line := range p {
		lines = append(lines, string(line))
	}
	return lines
}
