package renderer

import (
	"fmt"
	"html/template"
	"io"
	"strconv"
	"strings"
	"to/internal/node"
)

var FuncMap = template.FuncMap{
	"onlyLineComment": onlyLineComment,
}

func New(tmpl *template.Template) *Renderer {
	return &Renderer{
		tmpl: tmpl,
	}
}

type Renderer struct {
	tmpl *template.Template
}

func (r *Renderer) Render(out io.Writer, nodes []node.Node) {
	for i, n := range nodes {
		if i > 0 {
			out.Write([]byte("\n"))
		}

		switch n.(type) {
		case node.Ranked, node.BlockChildren, node.InlineChildren, node.Content, node.HeadBody:
		default:
			panic(fmt.Sprintf("HTML: unexpected node %T", n))
		}

		data := make(map[string]interface{})

		if m, ok := n.(node.Ranked); ok && m.Rank() > 0 {
			data["Rank"] = strconv.FormatUint(uint64(m.Rank()), 10)
		}

		if m, ok := n.(node.BlockChildren); ok {
			data["BlockChildren"] = m.BlockChildren()
		}

		if m, ok := n.(node.InlineChildren); ok {
			data["InlineChildren"] = m.InlineChildren()
		}

		if m, ok := n.(node.Content); ok {
			data["Content"] = string(m.Content())
		}

		if m, ok := n.(node.HeadBody); ok {
			data["Head"] = string(m.Head())
			data["Body"] = string(m.Body())
		}

		name := n.Node()
		if err := r.tmpl.ExecuteTemplate(out, name, data); err != nil {
			panic(err)
		}
	}
}

func (r *Renderer) FuncMap() template.FuncMap {
	return template.FuncMap{
		"render": func(v interface{}) template.HTML {
			var b strings.Builder

			switch n := v.(type) {
			case node.Node:
				r.Render(&b, []node.Node{n})
			default:
				panic(fmt.Sprintf("render: unexpected node %T", v))
			}

			return template.HTML(b.String())
		},
	}
}

func onlyLineComment(inlines []node.Inline) bool {
	if len(inlines) == 1 {
		_, ok := inlines[0].(node.LineComment)
		return ok
	}
	return false
}
