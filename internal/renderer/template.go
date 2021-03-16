package renderer

import (
	"fmt"
	"html/template"
	"io"
	"strconv"
	"strings"
	"to/internal/node"
)

var ElementTemplates = map[string]string{
	"Text": "{{.Content}}",
	"Line": "{{if not (onlyLineComment .InlineChildren)}}<div>{{range .InlineChildren}}{{render .}}{{end}}</div>{{end}}",

	"Blockquote": `<blockquote>
{{- range .BlockChildren}}
	{{render .}}
{{end -}}
</blockquote>`,
	"DescriptionList": `<dl>
{{- range $index, $element := .BlockChildren}}
	{{- if eq $index 0}}
	<dt>{{render $element}}</dt>
	{{- else}}
	<dd>{{render $element}}</dd>
	{{- end}}
{{end -}}
</dl>`,
	"CodeBlock": "<pre><code>{{.Body}}</code></pre>",
	"Heading": `<h{{.Rank}}>
{{- range .BlockChildren}}
	{{render .}}
{{end -}}
</h{{.Rank}}>`,

	"Emphasis": "<em>{{range .InlineChildren}}{{render .}}{{end}}</em>",
	"Strong":   "<strong>{{range .InlineChildren}}{{render .}}{{end}}</strong>",
	"Code":     "<code>{{.Content}}</code>",
	"Link":     `<a href="{{.Content}}">{{range .InlineChildren}}{{render .}}{{else}}{{.Content}}{{end}}</a>`,
}

var tmpls *template.Template

func init() {
	tmpls = template.New("tmpls")
	tmpls.Funcs(template.FuncMap{
		"render":          render,
		"onlyLineComment": onlyLineComment,
	})
	for name, tmpl := range ElementTemplates {
		template.Must(tmpls.New(name).Parse(tmpl))
	}
}

func render(a interface{}) template.HTML {
	var b strings.Builder

	switch n := a.(type) {
	case node.Node:
		HTML(&b, []node.Node{n})
	default:
		panic(fmt.Sprintf("render: unexpected node %T", a))
	}

	return template.HTML(b.String())
}

func HTML(w io.Writer, nodes []node.Node) {
	for i, n := range nodes {
		if i > 0 {
			w.Write([]byte("\n"))
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
		if err := tmpls.ExecuteTemplate(w, name, data); err != nil {
			panic(err)
		}
	}
}

func onlyLineComment(inlines []node.Inline) bool {
	if len(inlines) == 1 {
		_, ok := inlines[0].(node.LineComment)
		return ok
	}
	return false
}
