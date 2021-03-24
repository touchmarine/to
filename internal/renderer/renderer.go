package renderer

import (
	"fmt"
	"html/template"
	"io"
	"strconv"
	"strings"
	"to/internal/config"
	"to/internal/node"
)

var FuncMap = template.FuncMap{
	"onlyLineComment":  onlyLineComment,
	"head":             head,
	"body":             body,
	"primarySecondary": parsePrimarySecondary,
}

func New(conf *config.Config, tmpl *template.Template) *Renderer {
	return &Renderer{
		conf:   conf,
		tmpl:   tmpl,
		seqMap: make(map[string]map[uint]uint),
	}
}

type Renderer struct {
	conf   *config.Config
	tmpl   *template.Template
	seqMap map[string]map[uint]uint // ranked sequence elements by node
}

func (r *Renderer) Render(out io.Writer, nodes []node.Node) {
	for i, n := range nodes {
		if i > 0 {
			out.Write([]byte("\n"))
		}

		switch n.(type) {
		case node.Ranked, node.BlockChildren, node.InlineChildren, node.Content, node.Lines:
		default:
			panic(fmt.Sprintf("HTML: unexpected node %T", n))
		}

		data := make(map[string]interface{})

		if m, ok := n.(node.Ranked); ok {
			r.incSeqNum(n)

			data["Rank"] = strconv.FormatUint(uint64(m.Rank()), 10)
			data["SeqNum"] = r.seqNum(n)
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

func (r *Renderer) seqNum(n node.Node) string {
	m, ok := r.seqMap[n.Node()]
	if !ok {
		panic(fmt.Sprintf("Renderer.seqNum: missing map for ranked node %s", n.Node()))
	}

	ranked, ok := n.(node.Ranked)
	if !ok {
		panic(fmt.Sprintf("Renderer.seqNum: node %s does not implement node.Ranked", n.Node()))
	}

	rank := ranked.Rank()
	el, ok := r.conf.Element(n.Node())
	if !ok {
		panic(fmt.Sprintf("Renderer.seqNum: missing element in config for node %s", n.Node()))
	}

	minRank := el.MinRank

	var seq []string
	for i := minRank; i <= rank; i++ {
		var seqNum uint64
		u, ok := m[i]
		if ok {
			seqNum = uint64(u)
		}
		seq = append(seq, strconv.FormatUint(seqNum, 10))
	}

	return strings.Join(seq, ".")
}

func (r *Renderer) incSeqNum(n node.Node) {
	if _, ok := r.seqMap[n.Node()]; !ok {
		r.seqMap[n.Node()] = make(map[uint]uint)
	}

	ranked, ok := n.(node.Ranked)
	if !ok {
		panic(fmt.Sprintf("Renderer.incSeqNum: node %s does not implement node.Ranked", n.Node()))
	}

	rank := ranked.Rank()

	r.seqMap[n.Node()][rank]++

	for rk, _ := range r.seqMap[n.Node()] {
		if rk > rank {
			r.seqMap[n.Node()][rk] = 0
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
