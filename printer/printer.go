package printer

import (
	"bytes"
	"fmt"
	"github.com/touchmarine/to/config"
	"github.com/touchmarine/to/node"
	"io"
	"strings"
	"unicode/utf8"
)

const trace = false

func Fprint(w io.Writer, conf *config.Config, nodes []node.Node) {
	p := &printer{
		conf:  conf,
		w:     w,
		nodes: nodes,

		pos: -1,
	}

	p.printNodes()
}

type printer struct {
	conf   *config.Config
	w      io.Writer
	nodes  []node.Node
	parent node.Node

	n   node.Node
	pos int

	prefixes []string

	// tracing
	indent int
}

func (p *printer) printChildren(w io.Writer, nodes []node.Node) {
	if p.n == nil {
		panic("printer: nil parent")
	}

	if trace {
		defer p.trace("printChildren")()
	}

	n := &printer{
		conf:   p.conf,
		w:      w,
		nodes:  nodes,
		parent: p.n,

		pos: -1,

		prefixes: p.prefixes,

		indent: p.indent,
	}

	n.printNodes()
}

func (p *printer) printNodes() {
	if trace {
		defer p.tracef("printNodes (%d)", len(p.nodes))()
	}

	for p.next() {
		if p.pos > 0 {
			if _, isInGroup := p.parent.(*node.Group); isLine(p.n) || isInGroup {
				p.newline(p.w)
			} else {
				p.newline(p.w)
				p.newline(p.w)
			}
		}

		if !p.unbox() {
			continue
		}

		p.printNode()
	}
}

func (p *printer) next() bool {
	p.pos++

	if p.pos >= len(p.nodes) {
		return false
	}

	p.n = p.nodes[p.pos]

	return true
}

func (p *printer) unbox() bool {
	var n node.Node

	switch m := p.n.(type) {
	case node.Boxed:
		if trace {
			defer p.trace("unbox")()
		}

		switch k := m.(type) {
		case *node.Hat:
			p.printLines(p.w, []byte("%"), k.Lines())

			n = m.Unbox()
			if n == nil {
				if p.pos < len(p.nodes) {
					p.newline(p.w)
					p.newline(p.w)
				}

				if trace {
					p.print("empty")
				}

				return false
			}

			p.w.Write([]byte("\n"))
		case *node.SeqNumBox:
			n = m.Unbox()
			if n == nil {
				if trace {
					p.print("empty")
				}

				return false
			}
		default:
			panic(fmt.Sprintf("printer: unexpected Boxed node %T", n))
		}

		p.n = n
	}

	return true
}

func (p *printer) newline(w io.Writer) {
	if trace {
		p.print("newline")
	}

	w.Write([]byte("\n"))

	if len(p.prefixes) > 0 {
		w.Write([]byte(strings.Join(p.prefixes, " ")))
	}
}

func (p *printer) printNode() {
	if trace {
		defer p.tracef("printNode (%d)", p.pos)()
	}

	var b bytes.Buffer

	for _, prefix := range p.prefixes {
		// TODO: Differentiate real and space prefixes for hanging
		// blocks and add spacing between prefixes
		p.printf("prefix=%q", prefix)
		b.WriteString(prefix)
	}

	pre, post, needInlineEscape := p.delimiters()

	switch p.n.(type) {
	case node.Block:
		defer func(b []string) {
			p.prefixes = b
		}(p.prefixes)

		switch p.n.(type) {
		case *node.Hanging, *node.HangingVerbatim:
			s := strings.Repeat(" ", len(pre))
			p.prefixes = append(p.prefixes, s)
		default:
			p.prefixes = append(p.prefixes, pre)
		}

		if post != "" {
			defer func() {
				switch p.n.(type) {
				case *node.Fenced:
					p.newline(&b)
				}

				b.Write([]byte(post))
			}()
		}
	case node.Inline:
		if p.atBOL() && needInlineEscape {
			b.Write([]byte(`\`))
		}

		b.Write([]byte(pre))

		defer func() {
			b.Write([]byte(post))
		}()
	default:
		panic(fmt.Sprintf("printer: unexpected node type %T", p.n))
	}

	if isLine(p.n) && p.needBlockEscape() {
		b.Write([]byte(`\`))
	}

	switch m := p.n.(type) {
	case node.Text:
		p.printText(&b, m)
	case node.ContentInlineChildren:
		if ic := m.InlineChildren(); len(ic) > 0 {
			p.printChildren(&b, node.InlinesToNodes(ic))
			b.Write([]byte(post + pre))
		}

		b.Write(m.Content())
	case node.BlockChildren:
		p.printChildren(&b, node.BlocksToNodes(m.BlockChildren()))
	case node.InlineChildren:
		p.printChildren(&b, node.InlinesToNodes(m.InlineChildren()))
	case node.Lines:
		p.printLines(&b, nil, m.Lines())
	case node.Content:
		b.Write(m.Content())
	}

	if trace {
		p.printf("return %q", b.String())
	}

	b.WriteTo(p.w)

	return
}

func (p *printer) atBOL() bool {
	return isLine(p.parent) && p.pos == 0
}

func (p *printer) printLines(w io.Writer, prefix []byte, lines [][]byte) {
	if trace {
		defer p.trace("printLines")()
	}

	for i, ln := range lines {
		if i > 0 {
			p.newline(w)
		}

		if len(ln) > 0 {
			if len(prefix) > 0 {
				w.Write(prefix)
				w.Write([]byte(" "))
			}

			w.Write(bytes.Trim(ln, " \t"))
		}
	}
}

func (p *printer) needBlockEscape() bool {
	txt := node.ExtractText(p.n)
	if txt == "" {
		return false
	}

	for _, e := range p.conf.Elements {
		if node.TypeCategory(e.Type) == node.CategoryBlock {
			if strings.HasPrefix(txt, e.Delimiter) {
				return true
			}
		}
	}

	return false
}

func (p *printer) printText(w io.Writer, t node.Text) {
	if trace {
		defer p.trace("printText")()
	}

	con := string(t.Content())
	if con == "" {
		return
	}

	var b bytes.Buffer

	for i, ch := range con {
		if ch == '\\' {
			b.Write([]byte(`\`))
		}

	L:
		for _, e := range p.conf.Elements {
			if node.TypeCategory(e.Type) == node.CategoryInline {
				// perfom the same check as parser
				d, _ := utf8.DecodeRuneInString(e.Delimiter)
				if d == utf8.RuneError {
					panic("printer: invalid UTF-8 encoding in delimiter")
				}

				var cases []string

				switch e.Type {
				case node.TypeUniform:
					cases = append(cases, string(d)+string(d))
				case node.TypeEscaped:
					cases = append(cases, string(d)+string(d))

					for _, c := range counterpartChars {
						cases = append(cases, string(d)+string(c))
					}
				case node.TypeForward:
				default:
					panic(fmt.Sprintf("printer: unexpected node type %s", e.Type))
				}

				for _, c := range cases {
					if strings.HasPrefix(con[i:], c) {
						b.Write([]byte(`\`))
						break L
					}
				}
			}
		}

		b.WriteRune(ch)
	}

	if trace {
		p.printf("return %q", b.String())
	}

	b.WriteTo(w)
}

func (p *printer) delimiters() (string, string, bool) {
	var pre, post string
	var needInlineEscape bool // whether inline delimiter should be escaped at BOL

	switch name := p.n.Node(); name {
	case "Text", "Line", "Paragraph":
	case "LineComment":
		pre = "//"
	default:
		el, ok := p.conf.Element(name)
		if !ok {
			_, grpOk := p.conf.Group(name)
			if grpOk {
				return "", "", false
			} else {
				panic("printer: unexpected element " + name)
			}
		}

		delim := el.Delimiter
		typ := el.Type

		switch p.n.(type) {
		case node.Block:
			switch m := p.n.(type) {
			case *node.Fenced:
				pre = delim + delim
				post = delim + delim
			case node.Ranked:
				rank := m.Rank()
				if rank > 0 {
					for i := 0; i < int(rank); i++ {
						pre += delim
					}
				} else {
					pre = delim
				}
			default:
				pre = delim
			}

		case node.Inline:
			r, _ := utf8.DecodeRuneInString(delim)
			// perfom the same check as parser
			if r == utf8.RuneError {
				panic("printer: invalid UTF-8 encoding in delimiter")
			}

			if p.needInlineEscape(delim) {
				needInlineEscape = true
			}

			switch typ {
			case node.TypeUniform:
				pre = delim + delim
				post = pre
			case node.TypeEscaped:
				var content []byte
				if m, ok := p.n.(node.Content); ok {
					content = m.Content()
				} else {
					panic("printer: escaped element " + name + " does not implement node.Content")
				}

				if bytes.Contains(content, []byte(delim+delim)) {
					// find another escape combination for delim

					for _, ch := range counterpartChars {
						cp := counterpart(ch)
						pre0 := delim + string(ch)
						post0 := string(cp) + delim

						if bytes.Contains(content, []byte(pre0)) ||
							bytes.Contains(content, []byte(post0)) {
							continue
						}

						pre = pre0
						post = post0
						break
					}

					if pre == "" || post == "" {
						panic("printer: no escape character available")
					}
				} else {
					pre = delim + delim
					post = pre
				}
			case node.TypeForward:
				pre = delim
				post = string(counterpart(r))
			default:
				panic(fmt.Sprintf("printer: unexpected node type %s (%s)", typ, name))
			}

		default:
			panic(fmt.Sprintf(
				"parser: unexpected node type %T (element=%q, delimiter=%q)",
				p.n,
				name,
				delim,
			))
		}
	}

	return pre, post, needInlineEscape
}

func (p *printer) needInlineEscape(delim string) bool {
	for _, e := range p.conf.Elements {
		if node.TypeCategory(e.Type) == node.CategoryBlock {
			if e.Delimiter == delim {
				return true
			}
		}
	}

	return false
}

func (p *printer) tracef(format string, v ...interface{}) func() {
	return p.trace(fmt.Sprintf(format, v...))
}

func (p *printer) trace(msg string) func() {
	var name string
	if p.n != nil {
		name = p.n.Node()
	}

	p.printf("%q %T -> %s (", name, p.n, msg)
	p.indent++

	return func() {
		p.indent--
		p.print(")")
	}
}

func (p *printer) printf(format string, v ...interface{}) {
	p.print(fmt.Sprintf(format, v...))
}

func (p *printer) print(msg string) {
	fmt.Println(strings.Repeat("\t", p.indent) + msg)
}

func isLine(n node.Node) bool {
	_, ok := n.(*node.Line)
	return ok && n.Node() == "Line"
}

func counterpart(ch rune) rune {
	c, ok := leftRightChars[ch]
	if ok {
		return c
	}
	return ch
}

// ordered counterpart characters, must contain same chars as leftRightChars
var counterpartChars = []rune{
	'{',
	'[',
	'(',
	'<',
	'}',
	']',
	')',
	'>',
}

// same as in parser/parser.go
var leftRightChars = map[rune]rune{
	'{': '}',
	'}': '{',
	'[': ']',
	']': '[',
	'(': ')',
	')': '(',
	'<': '>',
	'>': '<',
}
