package renderer

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/touchmarine/to/aggregator"
	"github.com/touchmarine/to/node"
	"html/template"
	"io"
	"strconv"
	"strings"
)

var FuncMap = template.FuncMap{
	"dict":          dict,
	"head":          head,
	"body":          body,
	"groupBySeqNum": groupBySeqNum,
	"isSeqNumGroup": isSeqNumGroup,
	"parseAttrs":    parseAttrs,
	"htmlAttrs":     htmlAttrs,
	"trimSpacing":   trimSpacing,
}

func New(tmpl *template.Template, data map[string]interface{}) *Renderer {
	return &Renderer{tmpl, data}
}

type Renderer struct {
	tmpl *template.Template
	data map[string]interface{}
}

func (r *Renderer) RenderWithCustomRoot(out io.Writer, nodes []node.Node) {
	if err := r.tmpl.ExecuteTemplate(out, "root", nodes); err != nil {
		panic(err)
	}
}

func (r *Renderer) Render(out io.Writer, nodes []node.Node, tmplData map[string]interface{}) {
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
		case *node.Hat:
			lines := btosSlice(k.Lines())
			joined := strings.Join(lines, "\n")

			data["Hat"] = joined
			data["HatAttrs"] = parseAttrs(lines)
		case *node.SeqNumBox:
			data["SeqNums"] = k.SeqNums
			data["SeqNum"] = k.SeqNum()
		default:
			panic(fmt.Sprintf("render: unexpected Boxed node %T", n))
		}
	}
}

func (r *Renderer) FuncMap() template.FuncMap {
	return template.FuncMap{
		"render":         r.renderFunc,
		"renderWithData": r.renderWithDataFunc,
	}
}

func (r *Renderer) renderFunc(v interface{}) template.HTML {
	return r.renderWithDataFunc(v, nil)
}

func (r *Renderer) renderWithDataFunc(v interface{}, data map[string]interface{}) template.HTML {
	var b strings.Builder

	switch n := v.(type) {
	case []node.Node:
		r.Render(&b, n, data)
	case node.Node:
		r.Render(&b, []node.Node{n}, data)
	default:
		panic(fmt.Sprintf("render: unexpected node %T", v))
	}

	return template.HTML(b.String())
}

// btosSlice returns a slice of strings from a slice of bytes.
func btosSlice(p [][]byte) []string {
	var lines []string
	for _, line := range p {
		lines = append(lines, string(line))
	}
	return lines
}

func dict(v ...interface{}) (map[string]interface{}, error) {
	if len(v)%2 > 0 {
		return nil, fmt.Errorf("dict got odd number of parameters")
	}

	dict := make(map[string]interface{}, len(v)/2)

	for i := 0; i < len(v); i += 2 {
		key, ok := v[i].(string)
		if !ok {
			return nil, fmt.Errorf("dict key not a string")
		}

		if i+1 < len(v) {
			dict[key] = v[i+1]
		}
	}

	return dict, nil
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

const trace = false

func groupBySeqNum(items []aggregator.Item) seqNumGroup {
	s := seqNumGrouper{}
	return s.groupBySeqNum(items)
}

// seqNumGrouper groups aggregator items by their sequence number.
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

func parseAttrs(lines []string) map[string]string {
	var b strings.Builder
	for _, l := range lines {
		if strings.HasPrefix(l, "!") {
			b.WriteString(l[1:])
		}
	}

	reader := strings.NewReader(b.String())

	var p attrParser
	p.init(reader)
	p.parse()

	return p.attrs
}

// attrParser holds the parser state.
//
// attrParser parses HTML-like attributes. An attribute has the folowing form:
//
// <name> = <value>
//
// where <name> is a string of Unicode characters except spacing or newline and
// <value> is a string of Unicode characaters except spacing or a double quote
// '"' delimited string that can contain spacing and escape sequences or a
// single quote "'" delimited string that can contain raw content.
//
// Only the name is required. Spacing is a space or a tab.
//
// Attributes must be separated by spacing except after an attribute with quoted
// value—they can be placed one after another as such:
//
// a="b"c=d
type attrParser struct {
	reader *bufio.Reader

	ch byte

	attrs map[string]string
}

func (p *attrParser) init(r io.Reader) {
	p.reader = bufio.NewReader(r)
	p.attrs = make(map[string]string)
}

func (p *attrParser) next() bool {
	b, err := p.reader.ReadByte()
	if err != nil {
		return false
	}

	p.ch = b
	return true
}

func (p *attrParser) peek() byte {
	b, err := p.reader.Peek(1)
	if err != nil && !errors.Is(err, io.EOF) {
		panic(err)
	}

	if l := len(b); l == 0 {
		return 0
	} else if l == 1 {
		return b[0]
	} else {
		panic("renderer: unexpected byte length")
	}
}

func (p *attrParser) isSpacing() bool {
	return p.ch == '\t' || p.ch == ' '
}

func (p *attrParser) parse() {
	for p.next() {
		if p.isSpacing() || p.ch == '\n' {
			continue
		}

		name, value := p.parseAttr()
		if name != "" {
			p.attrs[name] = value
		}
	}
}

func (p *attrParser) parseAttr() (string, string) {
	var name, value strings.Builder

	var hitEquals bool // whether encountered name-value delimiter "="
	var quote byte     // which quote if any

	for {
		if hitEquals {
			// in value part of attribute

			if quote != 0 {
				// inside quotes

				if p.ch == quote {
					// closing quote
					break
				}

				if quote == '"' {
					// inside double quotes '"'

					if peek := p.peek(); p.ch == '\\' && (peek == '\\' || peek == '"') {
						// escape
						value.WriteByte(peek)

						if !p.next() { // skip escape backslash
							panic("renderer: no escape backslash")
						}
						if !p.next() { // skip escaped char
							panic("renderer: no escaped char")
						}

						continue
					}
				} else if quote == '\'' {
					// inside single quotes "'" (raw content)
				} else {
					panic("renderer: quote is neither '\"' or \"'\"")
				}

				value.WriteByte(p.ch)
			} else {
				if p.isSpacing() || p.ch == '\n' {
					break
				}

				value.WriteByte(p.ch)
			}
		} else if p.ch == '=' {
			// name-value delimiter

			hitEquals = true

			if peek := p.peek(); peek == '"' || peek == '\'' {
				// opening quote
				quote = peek

				if !p.next() {
					panic("renderer: no equals char")
				}
				if !p.next() {
					panic("renderer: no opening quote")
				}

				continue
			}
		} else {
			// in name part of attribute

			if p.isSpacing() || p.ch == '\n' {
				break
			}

			name.WriteByte(p.ch)
		}

		if !p.next() {
			break
		}
	}

	return name.String(), value.String()
}

// htmlAttrs returns a HTML-formatted string of attributes from the given attrs.
func htmlAttrs(attrs map[string]string) template.HTMLAttr {
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

func trimSpacing(s string) string {
	return strings.Trim(s, " \t")
}
