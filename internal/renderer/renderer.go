package renderer

import (
	"fmt"
	"html/template"
	"io"
	"strconv"
	"strings"
	"to/internal/aggregator"
	"to/internal/node"
)

var FuncMap = template.FuncMap{
	"onlyLineComment":  onlyLineComment,
	"head":             head,
	"body":             body,
	"primarySecondary": parsePrimarySecondary,
	"groupBySeqNum":    groupBySeqNum,
	"isSeqNumGroup":    isSeqNumGroup,
	"toSeqNumGroup":    toSeqNumGroup,
	"toSeqNumItem":     toSeqNumItem,
}

func New(tmpl *template.Template, data map[string]interface{}) *Renderer {
	return &Renderer{tmpl, data}
}

type Renderer struct {
	tmpl *template.Template
	data map[string]interface{}
}

func (r *Renderer) Render(out io.Writer, nodes []node.Node) {
	for i, n := range nodes {
		if n.Node() == "BlankLine" {
			continue
		}

		if i > 0 {
			out.Write([]byte("\n"))
		}

		switch n.(type) {
		case node.Boxed, node.Ranked, node.BlockChildren, node.InlineChildren, node.Content, node.Lines:
		default:
			panic(fmt.Sprintf("Render: unexpected node %T", n))
		}

		data := make(map[string]interface{})

		if m, ok := n.(node.Boxed); ok {
			switch k := n.(type) {
			case *node.SeqNumBox:
				data["SeqNums"] = k.SeqNums
				data["SeqNum"] = k.SeqNum()
			default:
				panic(fmt.Sprintf("Render: unexpected Boxed node %T", n))
			}

			n = m.Unbox()
		}

		data["TextContent"] = node.ExtractText(n)

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
			data["Text"] = strings.Join(lines, "\n")
		}

		for k, v := range r.data {
			data[k] = v
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

type seqNumNode interface {
	seqNumNode()
}

type seqNumGroup []seqNumNode

func (g seqNumGroup) seqNumNode() {}

type seqNumItem aggregator.Item

func (i seqNumItem) seqNumNode() {}

func isSeqNumGroup(v interface{}) bool {
	_, ok := v.(seqNumGroup)
	return ok
}

func toSeqNumGroup(v interface{}) seqNumGroup {
	return v.(seqNumGroup)
}

func toSeqNumItem(v interface{}) seqNumItem {
	return v.(seqNumItem)
}

const trace = false

func groupBySeqNum(items []aggregator.Item) seqNumGroup {
	s := seqNumGrouper{}
	return s.groupBySeqNum(items)
}

type seqNumGrouper struct {
	lowest int
	indent int
}

func (s *seqNumGrouper) groupBySeqNum(items []aggregator.Item) seqNumGroup {
	if trace {
		defer s.trace("groupBySeqNum")()
	}

	var group seqNumGroup

	var depth int
	if len(items) > 0 {
		depth = len(items[0].SeqNums)
	}

	if trace {
		s.printf("depth=%d", depth)
	}

	if s.lowest == 0 || s.lowest > 0 && depth < s.lowest {
		if trace {
			s.printf("set lowest=%d", depth)
		}
		s.lowest = depth
	}

	for i := 0; i < len(items); i++ {
		item := items[i]
		cur := len(item.SeqNums) // current depth

		if trace {
			s.printf("item %s", item.SeqNum)
		}

		if cur > depth {
			if trace {
				s.printf("%d > depth:", cur)
			}

			g := s.groupBySeqNum(items[i:])
			i += len(g) - 1
			group = append(group, g)
		} else if cur == depth {
			if trace {
				s.printf("%d == depth, add to group", cur)
			}

			group = append(group, seqNumItem(item))
		} else if cur < depth {
			if trace {
				s.printf("%d < depth:", cur)
			}

			if cur < s.lowest {
				g := s.groupBySeqNum(items[i:])
				i += len(g) - 1
				group = append(seqNumGroup{group}, g...)
			}
			break
		}
	}

	return group
}

func (s *seqNumGrouper) trace(msg string) func() {
	s.printf("%s (", msg)
	s.indent++

	return func() {
		s.indent--
		s.print(")")
	}
}

func (s *seqNumGrouper) printf(format string, v ...interface{}) {
	s.print(fmt.Sprintf(format, v...))
}

func (s *seqNumGrouper) print(msg string) {
	fmt.Println(strings.Repeat("\t", s.indent) + msg)
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
