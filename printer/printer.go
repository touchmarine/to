package printer

import (
	"bytes"
	"fmt"
	"github.com/touchmarine/to/config"
	"github.com/touchmarine/to/node"
	"github.com/touchmarine/to/stringifier"
	"io"
	"os"
	"strings"
	"unicode/utf8"
)

const trace = false

func Fprint(out io.Writer, conf *config.Config, nodes []node.Node) {
	p := &printer{
		conf: conf,
		w:    out,
	}
	p.print(nodes)
}

type printer struct {
	conf *config.Config
	w    io.Writer

	indent int
	atBOL  bool
}

func (p *printer) print(nodes []node.Node) {
	for i, n := range nodes {
		switch m := n.(type) {
		case node.Boxed:
			if trace {
				stringifier.StringifyTo(os.Stdout, n)
			}

			switch k := m.(type) {
			case *node.Hat:
				p.writeLines([]byte("%"), k.Lines())

				n = m.Unbox()
				if n == nil {
					if i < len(nodes)-1 {
						p.w.Write([]byte("\n\n"))
					}
					continue
				}

				p.w.Write([]byte("\n"))
			case *node.SeqNumBox:
				n = m.Unbox()
				if n == nil {
					continue
				}
			default:
				panic(fmt.Sprintf("printer: unexpected Boxed node %T", n))
			}
		}

		if trace {
			stringifier.StringifyTo(os.Stdout, n)
		}

		name := n.Node()

		if _, ok := n.(*node.Line); ok && name == "Line" {
			p.atBOL = true
		}

		if i > 0 && p.indent > 0 {
			p.w.Write(bytes.Repeat([]byte(" "), p.indent))
		}

		pre, post := p.delimiters(n)

		p.w.Write([]byte(pre))

		hanging := false
		switch n.(type) {
		case *node.Hanging, *node.HangingVerbatim:
			hanging = true
		}

		if hanging {
			p.indent += len(pre)
		}

		if ln, ok := n.(*node.Line); ok && p.needBlockEscape(ln) {
			p.w.Write([]byte(`\`))
		}

		switch m := n.(type) {
		case node.Text:
			p.printText(m)
		case node.ContentInlineChildren:
			if ic := m.InlineChildren(); len(ic) > 0 {
				p.print(node.InlinesToNodes(ic))
				p.w.Write([]byte(post + pre))
			}
			p.w.Write(m.Content())
		case node.BlockChildren:
			p.print(node.BlocksToNodes(m.BlockChildren()))
		case node.InlineChildren:
			p.print(node.InlinesToNodes(m.InlineChildren()))
		case node.Lines:
			p.writeLines(nil, m.Lines())
		case node.Content:
			p.w.Write(m.Content())
		}

		if hanging {
			p.indent -= len(pre)
		}

		p.w.Write([]byte(post))

		if _, ok := n.(node.Block); ok && i < len(nodes)-1 {
			if _, ok := n.(*node.Line); ok && name == "Line" {
				// do nothing
			} else if p.indent == 0 {
				p.w.Write([]byte("\n"))
			}
			p.w.Write([]byte("\n"))
		}

		if p.atBOL {
			p.atBOL = false
		}
	}
}

func (p *printer) writeLines(prefix []byte, lines [][]byte) {
	for i, ln := range lines {
		if i > 0 {
			p.w.Write([]byte("\n"))
			if p.indent > 0 {
				p.w.Write(bytes.Repeat([]byte(" "), p.indent))
			}
		}

		if len(ln) > 0 {
			if len(prefix) > 0 {
				p.w.Write(prefix)
				p.w.Write([]byte(" "))
			}

			p.w.Write(bytes.Trim(ln, " \t"))
		}
	}
}

func (p *printer) needBlockEscape(ln *node.Line) bool {
	txt := node.ExtractText(ln)
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

func (p *printer) printText(t node.Text) {
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

	b.WriteTo(p.w)

}

func (p *printer) delimiters(n node.Node) (string, string) {
	var pre, post string

	name := n.Node()
	switch name {
	case "Text", "Line", "Paragraph":
	case "LineComment":
		pre = "//"

		txt := node.ExtractText(n)
		if txt != "" {
			pre += " "
		}
	default:
		el, ok := p.conf.Element(name)
		if !ok {
			_, grpOk := p.conf.Group(name)
			if grpOk {
				return "", ""
			} else {
				panic("printer: unexpected element " + name)
			}
		}

		delim := el.Delimiter
		typ := el.Type

		switch n.(type) {
		case node.Block:
			switch m := n.(type) {
			case *node.Fenced:
				pre = delim + delim
				post = "\n" + delim + delim
			case node.Ranked:
				rank := m.Rank()
				if rank > 0 {
					for i := 0; i < int(rank); i++ {
						pre += delim
					}
				} else {
					pre = delim
				}

				txt := node.ExtractText(n)
				if txt != "" {
					pre += " "
				}
			default:
				pre = delim

				txt := node.ExtractText(n)
				if txt != "" {
					pre += " "
				}
			}

		case node.Inline:
			r, _ := utf8.DecodeRuneInString(delim)
			// perfom the same check as parser
			if r == utf8.RuneError {
				panic("printer: invalid UTF-8 encoding in delimiter")
			}

			if p.atBOL {
				// add backslash escape if needed
				for _, e := range p.conf.Elements {
					if node.TypeCategory(e.Type) == node.CategoryBlock {
						if e.Delimiter == delim {
							p.w.Write([]byte(`\`))
							break
						}
					}
				}
			}

			switch typ {
			case node.TypeUniform:
				pre = delim + delim
				post = pre
			case node.TypeEscaped:
				var content []byte
				if m, ok := n.(node.Content); ok {
					content = m.Content()
				} else {
					panic("printer: escaped element " + name + " does not implement node.Content")
				}

				if bytes.Contains(content, []byte(delim+delim)) {
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
				n,
				name,
				delim,
			))
		}
	}

	return pre, post
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
