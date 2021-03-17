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
	"onlyLineComment": onlyLineComment,
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
		case node.Ranked, node.BlockChildren, node.InlineChildren, node.Content, node.HeadBody:
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
