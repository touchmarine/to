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
		if !p.unbox() {
			continue
		}

		p.printNode()

		if peek := p.peek(); peek != nil && !p.isInline() {
			_, isInSticky := p.parent.(*node.Sticky)
			_, isInGroup := p.parent.(*node.Group)

			if isInSticky || isInGroup {
				p.newline(p.w)
			} else {
				p.newline(p.w)
				p.prefix(p.w, false)
				p.newline(p.w)
			}
		}
	}
}

func (p *printer) next() bool {
	if p.pos+1 < len(p.nodes) {
		p.n = p.nodes[p.pos+1]
		p.pos++

		return true
	}

	return false
}

func (p *printer) peek() node.Node {
	if p.pos+1 < len(p.nodes) {
		return p.nodes[p.pos+1]
	}
	return nil
}

func (p *printer) unbox() bool {
	var n node.Node

	switch m := p.n.(type) {
	case node.Boxed:
		if trace {
			defer p.trace("unbox")()
		}

		var b bytes.Buffer

		defer func() {
			if trace {
				p.printf("return %q", b.String())
			}

			b.WriteTo(p.w)
		}()

		switch m.(type) {
		case *node.SeqNumBox:
			n = m.Unbox()
			if n == nil {
				if trace {
					p.print("empty seqNumBox")
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
}

func (p *printer) printNode() {
	if trace {
		defer p.tracef("printNode (%d)", p.pos)()
	}

	if isEmpty(p.n) {
		if trace {
			p.print("return, empty")
		}

		return
	}

	var b bytes.Buffer

	defer func() {
		if trace {
			p.printf("return %q", b.String())
		}

		b.WriteTo(p.w)
	}()

	if !p.isInline() && p.pos > 0 {
		p.prefix(&b, true)
	}

	pre, post := p.delimiters()

	switch p.n.(type) {
	case node.Block:
		if pre != "" {
			b.WriteString(pre)

			switch p.n.(type) {
			case *node.Fenced, *node.VerbatimWalled:
			default:
				b.WriteString(" ")
			}

			defer func(size int) {
				p.prefixes = p.prefixes[:size]
			}(len(p.prefixes))

			switch p.n.(type) {
			case *node.Hanging, *node.RankedHanging:
				s := strings.Repeat(" ", len(pre))
				p.prefixes = append(p.prefixes, s)
			case *node.Walled, *node.VerbatimWalled:
				p.prefixes = append(p.prefixes, pre)
			}
		}

		if post != "" {
			defer func() {
				switch p.n.(type) {
				case *node.Fenced:
					p.newline(&b)
					p.prefix(&b, true)
				}

				b.WriteString(post)
			}()
		}
	case node.Inline:
		b.WriteString(pre)

		defer func() {
			b.WriteString(post)
		}()
	default:
		panic(fmt.Sprintf("printer: unexpected node type %T", p.n))
	}

	if isTextBlock(p.n) && p.needBlockEscape() {
		b.WriteString(`\`)
	}

	switch m := p.n.(type) {
	case node.Text:
		p.printText(&b, m)
	case node.BlockChildren:
		p.printChildren(&b, node.BlocksToNodes(m.BlockChildren()))
	case node.InlineChildren:
		p.printChildren(&b, node.InlinesToNodes(m.InlineChildren()))
	case node.Lines:
		_, isVerbatim := p.n.(*node.VerbatimWalled)
		p.printLines(&b, m.Lines(), !isVerbatim)
	case node.Content:
		c := m.Content()

		if _, ok := p.n.(node.Block); ok {
			c = bytes.Trim(c, " \t")
		}

		b.Write(c)
	case node.Composited:
		p.printChildren(&b, []node.Node{m.Primary(), m.Secondary()})
	}

	return
}

func (p *printer) printLines(w io.Writer, lines [][]byte, spacing bool) {
	if trace {
		defer p.trace("printLines")()
	}

	for i, ln := range lines {
		if i > 0 {
			p.newline(w)
			p.prefix(w, spacing)
		}

		if len(ln) > 0 {
			w.Write(ln)
		}
	}
}

// prefix writes the current prefix to w. It adds a trailing space if spacing
// is true and removes any trailing spacing if spacing is false.
func (p *printer) prefix(w io.Writer, spacing bool) {
	if len(p.prefixes) == 0 {
		return
	}

	prefix := strings.Join(p.prefixes, " ")

	if spacing {
		prefix += " "
	} else {
		prefix = strings.Trim(prefix, " \t")
	}

	if trace {
		p.printf("prefix=%q", prefix)
	}

	w.Write([]byte(prefix))
}

func (p *printer) needBlockEscape() bool {
	textBlock, isBasicBlock := p.n.(*node.BasicBlock)
	if !isBasicBlock || p.n.Node() != "TextBlock" {
		panic("printer: expected TextBlock")
	}

	children := textBlock.InlineChildren()
	if len(children) == 0 {
		return false
	}

	child := children[0]

	if text, ok := child.(node.Text); ok {
		// we only need to check the first text node, any other node
		// will have it's own non-block delimiter

		content := text.Content()

		if len(content) > 1 && content[0] == '\\' {
			return !p.hasInlineDelimiterPrefix(content[1:]) &&
				p.hasBlockDelimiterPrefix(content[1:])
		} else if len(content) > 0 {
			return !p.hasInlineDelimiterPrefix(content) &&
				p.hasBlockDelimiterPrefix(content)
		} else {
			return false
		}
	}

	return false
}

func (p *printer) hasBlockDelimiterPrefix(content []byte) bool {
	for _, e := range p.conf.Elements {
		if node.TypeCategory(e.Type) == node.CategoryBlock &&
			bytes.HasPrefix(content, []byte(e.Delimiter)) {
			return true
		}
	}

	return false
}

func (p *printer) printText(w io.Writer, t node.Text) {
	if trace {
		defer p.trace("printText")()
	}

	content := t.Content()
	if len(content) == 0 {
		return
	}

	if trace {
		p.printf("content=%q", content)
	}

	var b bytes.Buffer

	for i := 0; i < len(content); i++ {
		ch := content[i]

		// backslash escape checks
		if ch == '\\' && i+1 < len(content) && content[i+1] == '\\' {
			// consecutive backslashes

			b.WriteString(`\`)
		} else if ch == '\\' && i == len(content)-1 && p.peek() != nil && !isEmpty(p.peek()) {
			// backslash at the end of text content with a non-empty
			// inline element behind it

			_, isInline := p.peek().(node.Inline)
			if !isInline {
				panic("peek is not inline")
			}

			b.WriteString(`\`)
		} else if ch == '\\' && i+1 < len(content) && p.hasInlineDelimiterPrefix(content[i+1:]) {
			b.WriteString(`\`)
		} else if i == len(content)-1 && p.peek() != nil && !isEmpty(p.peek()) {
			peek := p.peek()

			_, isInline := peek.(node.Inline)
			if !isInline {
				panic("peek is not inline")
			}

			// get first node if composited
			composited, isComposite := peek.(node.Composited)
			if isComposite {
				peek = composited.Primary()
			}

			e, isElement := p.conf.Element(peek.Node())
			if !isElement {
				panic("printer: node " + peek.Node() + " not found")
			}

			if len(e.Delimiter) < 1 {
				panic("printer: invalid delimiter")
			}

			delimiter := e.Delimiter[0]

			if ch == delimiter {
				// current character and first character of the
				// next element's delimiter form an inline
				// element delimiter

				b.WriteString(`\`)
			}
		} else if p.hasInlineDelimiterPrefix(content[i:]) {
			// matches inline delimiter

			b.WriteString(`\`)
		}

		b.WriteByte(ch)
	}

	if trace {
		p.printf("return %q", b.String())
	}

	b.WriteTo(w)
}

func (p *printer) hasInlineDelimiterPrefix(content []byte) bool {
	for _, e := range p.conf.Elements {
		if node.TypeCategory(e.Type) == node.CategoryInline {
			delimiter := []byte(e.Delimiter + e.Delimiter)

			if bytes.HasPrefix(content, delimiter) {
				return true
			}
		}
	}

	return false
}

func (p *printer) delimiters() (string, string) {
	var pre, post string

	switch name := p.n.Node(); name {
	case "Text", "TextBlock", "Paragraph":
	default:
		el, ok := p.conf.Element(name)
		if !ok {
			_, isComposite := p.conf.Composite(name)
			_, isSticky := p.conf.Sticky(name)
			_, isGroup := p.conf.Group(name)
			if isComposite || isSticky || isGroup {
				return "", ""
			} else {
				panic("printer: unexpected element " + name)
			}
		}

		delim := el.Delimiter

		switch p.n.(type) {
		case node.Block:
			switch m := p.n.(type) {
			case *node.Fenced:
				pre = delim
				post = delim

				lines := m.Lines()

				var content []byte
				if len(lines) > 1 {
					content = bytes.Join(lines[1:], []byte("\n"))
				}

				if bytes.Contains(content, []byte(delim)) {
					// needs escape
					pre += "\\"
					post = "\\" + post
				}
			case node.Ranked:
				rank := m.Rank()

				for i := 0; i < rank; i++ {
					pre += delim
				}
			case *node.BasicBlock, *node.VerbatimLine, *node.Walled, *node.VerbatimWalled,
				*node.Hanging, *node.RankedHanging, *node.Group, *node.Sticky:
				pre = delim
			default:
				panic(fmt.Sprintf("printer: unexpected node type %T (%s)", p.n, name))
			}

		case node.Inline:
			switch p.n.(type) {
			case *node.Uniform, *node.Escaped:
				r, _ := utf8.DecodeRuneInString(delim)
				counterDelim := counterpart(r)

				pre = delim + delim
				post = string(counterDelim) + string(counterDelim)
			case *node.Prefixed:
				pre = delim
			default:
				panic(fmt.Sprintf("printer: unexpected node type %T (%s)", p.n, name))
			}

			if m, isEscaped := p.n.(*node.Escaped); isEscaped {
				content := m.Content()

				if bytes.Contains(content, []byte(delim+delim)) {
					// needs escape
					pre += "\\"
					post = "\\" + post
				}
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

	return pre, post
}

func (p *printer) atBOL() bool {
	return isTextBlock(p.parent) && p.pos == 0
}

func (p *printer) isInline() bool {
	_, ok := p.n.(node.Inline)
	return ok
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

func isEmpty(n node.Node) bool {
	return node.ExtractText(n) == ""
}

func isTextBlock(n node.Node) bool {
	_, ok := n.(*node.BasicBlock)
	return ok && n.Node() == "TextBlock"
}

func trimLines(lines [][]byte) [][]byte {
	var l [][]byte

	for _, line := range lines {
		t := bytes.Trim(line, " \t")
		if len(t) == 0 {
			continue
		}

		l = append(l, t)
	}

	return l
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
