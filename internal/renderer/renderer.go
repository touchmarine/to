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
	"onlyLineComment":  onlyLineComment,
	"head":             head,
	"body":             body,
	"primarySecondary": parsePrimarySecondary,
}

func New(tmpl *template.Template) *Renderer {
	return &Renderer{tmpl}
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
		case *node.SeqNumBox, node.Ranked, node.BlockChildren, node.InlineChildren, node.Content, node.Lines:
		default:
			panic(fmt.Sprintf("Render: unexpected node %T", n))
		}

		data := make(map[string]interface{})

		if m, ok := n.(node.Boxed); ok {
			switch k := n.(type) {
			case *node.SeqNumBox:
				data["SeqNum"] = k.SeqNum
			default:
				panic(fmt.Sprintf("Render: unexpected Boxed node %T", n))
			}

			n = m.Unbox()
		}

		if m, ok := n.(node.Ranked); ok {
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

		if m, ok := n.(node.Lines); ok {
			var lines []string
			for _, line := range m.Lines() {
				lines = append(lines, string(line))
			}
			data["Lines"] = lines
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

func head(lines []string) string {
	if len(lines) > 0 {
		return lines[0]
	}
	return ""
}

func body(lines []string) string {
	if len(lines) > 1 {
		return strings.Join(lines[1:], "\n")
	}
	return ""
}

type primarySecondary struct {
	Primary   template.HTMLAttr
	Secondary template.HTMLAttr
}

func parsePrimarySecondary(lines []string) primarySecondary {
	trimmed := make([]string, len(lines))
	for i := 0; i < len(lines); i++ {
		trimmed[i] = strings.Trim(lines[i], " \t")
	}

	s := strings.Join(trimmed, " ")
	i := strings.IndexAny(s, " \t")

	var prim, sec string
	if i > -1 {
		prim = s[:i]
		if i+1 < len(s) {
			sec = s[i+1:]
		}
	} else {
		prim = s
	}

	return primarySecondary{template.HTMLAttr(prim), template.HTMLAttr(sec)}
}

type namedUnnamed struct {
	Unnamed []string
	Named   map[string]string
}

func parseNamedUnnamed(lines []string) namedUnnamed {
	s := strings.Join(lines, ";")
	fields := strings.FieldsFunc(s, func(ch rune) bool {
		return ch == ';'
	})

	for i := 0; i < len(fields); i++ {
		// trim spacing
		fields[i] = strings.Trim(fields[i], " \t")
	}

	var x int
	for i := 0; i < len(fields); i++ {
		// filter out empty fields
		field := fields[i]
		if field != "" {
			fields[i] = field
			x++
		}
	}
	fields = fields[:x]

	var u []string
	n := map[string]string{}

	for _, field := range fields {
		i := strings.Index(field, ":")
		if i > -1 {
			// found
			name := field[:i]
			name = strings.Trim(name, " \t")

			var val string
			if i+1 < len(field) {
				val = field[i+1:]
			}
			val = strings.Trim(val, " \t")

			n[name] = val
		} else {
			ufields := strings.FieldsFunc(field, func(ch rune) bool {
				return ch == ' ' || ch == '\t'
			})

			u = append(u, ufields...)
		}
	}

	return namedUnnamed{u, n}
}
