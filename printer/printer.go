package printer

import (
	"bytes"
	"fmt"
	"github.com/touchmarine/to/node"
	"io"
	"strings"
	"unicode/utf8"
)

const trace = false

// ElementMap maps Elements to Names.
type ElementMap map[string]Element

type Element struct {
	Name        string
	Type        node.Type
	Delimiter   string
	Matcher     string
	DoNotRemove bool
}

func Fprint(w io.Writer, elementMap ElementMap, nodes []node.Node) {
	p := &printer{
		w:          w,
		elementMap: elementMap,
		nodes:      nodes,

		pos: -1,
	}

	p.init()
	p.printNodes()
}

type printer struct {
	w          io.Writer
	elementMap ElementMap
	nodes      []node.Node
	parent     *printer

	n   node.Node
	e   Element
	pos int

	prefixes         []string
	closingDelimiter string            // inline closing delimiter
	replacementMap   map[string]string // replacement map for marker elements
	leaf             string            // leaf element name

	// tracing
	indent int
}

func (p *printer) init() {
	// add marker elements to replacment map
	// marker elements are prefixed inline element with no matcher aka.
	// elements that have no content
	for name, e := range p.elementMap {
		if e.Type == node.TypeLeaf || e.Type == node.TypeText {
			delete(p.elementMap, name)
		} else if e.Type == node.TypePrefixed && e.Matcher == "" {
			// marker, e.g., "\" line break, without content

			if p.replacementMap == nil {
				p.replacementMap = make(map[string]string)
			}

			p.replacementMap[e.Name] = e.Delimiter
		}
	}
}

func (p *printer) closingDelimiters() []string {
	var d []string

	current := p
	for current != nil {
		if current.closingDelimiter != "" {
			d = append(d, current.closingDelimiter)
		}

		current = current.parent
	}

	return d
}

func (p *printer) printChildren(w io.Writer, nodes []node.Node) {
	if p.n == nil {
		panic("printer: nil parent")
	}

	if trace {
		defer p.trace("printChildren")()
	}

	n := &printer{
		w:          w,
		elementMap: p.elementMap,
		nodes:      nodes,
		parent:     p,

		pos: -1,

		prefixes:         p.prefixes,
		closingDelimiter: p.closingDelimiter,

		indent: p.indent,
	}

	n.init()
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
			var parentNode node.Node
			if p.parent != nil {
				parentNode = p.parent.n
			}

			_, isInSticky := parentNode.(*node.Sticky)
			_, isInGroup := parentNode.(*node.Group)

			if isInSticky || isInGroup {
				p.newline(p.w)
			} else {
				p.newline(p.w)
				p.prefix(p.w, noTrailingSpacing)
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
		case *node.SequentialNumberBox:
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

	if !p.shouldKeep(p.n) {
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
		p.prefix(&b, trailingSpacing)
	}

	pre, post := p.delimiters()

	switch p.n.(type) {
	case node.Block:
		if pre != "" {
			b.WriteString(pre)

			switch p.n.(type) {
			case *node.Fenced, *node.VerbatimWalled:
			default:
				if node.ExtractTextWithReplacements(p.n, p.replacementMap) != "" {
					// not empty as in we will print
					// something after this. p.shouldKeep()
					// would return true for `.tos`
					b.WriteString(" ")
				}
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
					p.prefix(&b, trailingSpacing)
				}

				b.WriteString(post)
			}()
		}
	case node.Inline:
		b.WriteString(pre)

		if post != "" {
			p.closingDelimiter = post

			defer func() {
				if p.hasEscapeClashingElementAtEnd() {
					// otherwise `**\` would return `**\**`

					p.newline(&b)
					p.prefix(&b, trailingSpacing)
				}

				b.WriteString(post)

				p.closingDelimiter = ""
			}()
		}
	default:
		panic(fmt.Sprintf("printer: unexpected node type %T", p.n))
	}

	if _, isLeaf := p.n.(*node.Leaf); isLeaf && p.needBlockEscape() {
		b.WriteString(`\`)
	}

	switch m := p.n.(type) {
	case *node.Text:
		if p.closingDelimiter != "" {
			// escape last characters in text according to delimiter
			p.printText(newDelimiterEscapeWriter(&b, p.closingDelimiter), m)
		} else {
			p.printText(&b, m)
		}
	case node.BlockChildren:
		p.printChildren(&b, node.BlocksToNodes(m.BlockChildren()))
	case node.InlineChildren:
		p.printChildren(&b, node.InlinesToNodes(m.InlineChildren()))
	case node.Lines:
		_, isVerbatim := p.n.(*node.VerbatimWalled)
		p.printLines(&b, m.Lines(), isVerbatim)
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

type delimiterEscapeWriter struct {
	delimiterEscaper
}

func newDelimiterEscapeWriter(w io.Writer, delimiter string) delimiterEscapeWriter {
	_, size := utf8.DecodeRuneInString(delimiter)
	if size != 1 {
		panic(fmt.Sprintf("printer: unexpected delimiter %q", delimiter))
	}

	dch := delimiter[0]

	return delimiterEscapeWriter{
		delimiterEscaper{
			w:         w,
			delimiter: delimiter,
			dch:       dch,
		},
	}
}

func (e delimiterEscapeWriter) Write(p []byte) (int, error) {
	e.init(p)
	return e.write()
}

type delimiterEscaper struct {
	w         io.Writer
	delimiter string
	dch       byte // delimiter character to escape

	src    []byte
	offset int
	ch     byte
}

func (e *delimiterEscaper) init(src []byte) {
	e.src = src
	e.offset = -1
	e.ch = 0
	e.next()
}

func (e *delimiterEscaper) next() {
	if e.offset+1 < len(e.src) {
		e.offset++
		e.ch = e.src[e.offset]
		return
	}

	e.offset = len(e.src)
	e.ch = 0
}

func (e *delimiterEscaper) peek() byte {
	if e.offset+1 < len(e.src) {
		return e.src[e.offset+1]
	}

	return 0
}

// TODO: Don't need this complex writer-escaper. Just do what this function does
// at the end of p.printText().
func (e *delimiterEscaper) write() (int, error) {
	nn := 0 // number of bytes written *from p* to satisfy io.Writer interface

	for ; e.ch > 0; e.next() {
		// escape characters at end of text that would otherwise escape
		// the closing delimiter
		if len(e.src) > 1 && e.offset == len(e.src)-2 && e.ch != '\\' && e.peek() == '\\' {
			// unescaped "\" at end of text
			n, err := e.w.Write([]byte{e.ch})
			nn += n
			if err != nil {
				return nn, err
			}

			e.next()

			_, err = e.w.Write([]byte(`\`))
			if err != nil {
				return nn, err
			}
		} else if len(e.src) == 1 && e.ch == '\\' ||
			len(e.src) > 0 && e.offset == len(e.src)-1 && (e.ch == e.dch) {
			// unescaped "\" at end of text (similar to above) or
			// closing delimiter character at end of text
			_, err := e.w.Write([]byte(`\`))
			if err != nil {
				return nn, err
			}
		}

		n, err := e.w.Write([]byte{e.ch})
		nn += n
		if err != nil {
			return nn, err
		}
	}

	if nn < len(e.src) {
		// satisfy io.Writer interface
		return nn, fmt.Errorf("written less than got")
	}

	return nn, nil
}

func (p *printer) printLines(w io.Writer, lines [][]byte, isVerbatim bool) {
	if trace {
		defer p.trace("printLines")()
	}

	for i, ln := range lines {
		if i > 0 {
			p.newline(w)
			if isVerbatim {
				p.prefix(w, noSpacing)
			} else {
				p.prefix(w, trailingSpacing)
			}
		}

		if len(ln) > 0 {
			w.Write(ln)
		}
	}
}

type prefixSpacing int

const (
	noSpacing prefixSpacing = iota
	trailingSpacing
	noTrailingSpacing
)

// prefix writes the current prefix to w. It adds a trailing space if spacing
// is trailingSpacing and removes any trailing spacing if spacing is
// noTrailingSpacing.
func (p *printer) prefix(w io.Writer, spacing prefixSpacing) {
	if len(p.prefixes) == 0 {
		return
	}

	prefix := strings.Join(p.prefixes, " ")

	switch spacing {
	case noSpacing:
	case trailingSpacing:
		prefix += " "
	case noTrailingSpacing:
		prefix = strings.Trim(prefix, " \t")

	}

	if trace {
		p.printf("prefix=%q", prefix)
	}

	w.Write([]byte(prefix))
}

func (p *printer) needBlockEscape() bool {
	leaf, isLeaf := p.n.(*node.Leaf)
	if !isLeaf {
		panic("printer: expected leaf node")
	}

	children := leaf.InlineChildren()
	if len(children) == 0 {
		return false
	}

	child := children[0]

	if text, ok := child.(*node.Text); ok {
		// we only need to check the first text node, any other node
		// will have it's own non-block delimiter

		content := text.Content()

		if len(content) > 0 {
			return !p.hasInlineDelimiterPrefix(content) &&
				p.hasBlockDelimiterPrefix(content)
		} else {
			return false
		}
	}

	return false
}

func (p *printer) hasBlockDelimiterPrefix(content []byte) bool {
	for _, e := range p.elementMap {
		if node.TypeCategory(e.Type) == node.CategoryBlock {
			// same logic as in parser/parser.go
			delimiter := ""
			if e.Type == node.TypeRankedHanging {
				delimiter = e.Delimiter + e.Delimiter
			} else {
				delimiter = e.Delimiter
			}

			if bytes.HasPrefix(content, []byte(delimiter)) {
				return true
			}
		}
	}

	return false
}

// hasEscapeClashingElementAtEnd returns true if the last inline child has the
// escape character as a delimiter.
func (p *printer) hasEscapeClashingElementAtEnd() bool {
	m, ok := p.n.(node.InlineChildren)
	if !ok {
		return false
	}

	last := m.InlineChildren()[len(m.InlineChildren())-1]

	e, found := p.elementMap[last.Node()]
	return found && e.Type == node.TypePrefixed && e.Delimiter == "\\"
}

func (p *printer) printText(w io.Writer, t *node.Text) {
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

			if trace {
				p.print("escape-A")
			}

			b.WriteString(`\`)
		} else if ch == '\\' && i+1 < len(content) && isPunct(content[i+1]) {
			// escape backslash so it doesn't escape the following
			// punctuation

			if trace {
				p.print("escape-B")
			}

			b.WriteString(`\`)
		} else if ch == '\\' && i == len(content)-1 && p.peek() != nil && p.shouldKeep(p.peek()) {
			// escape backslash so it doesn't escape the following
			// non-emtpy inline element

			if trace {
				p.print("escape-C")
			}

			_, isInline := p.peek().(node.Inline)
			if !isInline {
				panic("peek is not inline")
			}

			b.WriteString(`\`)
		} else if ch == '\\' && i+1 < len(content) && p.hasInlineDelimiterPrefix(content[i+1:]) {
			// escape backslash so it doesn't escape the following
			// inline delimiter
			//
			// needed for non-punctuation delimiters that aren't
			// caught by escape-B; as such don't need to check
			// the closing delimiters as non-punctuation can only be
			// prefixed elements

			if trace {
				p.print("escape-D")
			}

			b.WriteString(`\`)
		} else if p.hasInlineDelimiterPrefix(content[i:]) || p.hasClosingDelimiterPrefix(content[i:]) {
			// escape inline delimiter

			if trace {
				p.print("escape-E")
			}

			b.WriteString(`\`)
		} else if i == len(content)-1 && p.peek() != nil && p.shouldKeep(p.peek()) {
			// last character and the following non-empty element
			// delimiter's first character may form an inline
			// delimiter

			if trace {
				p.print("escape-F")
			}

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

			e, isElement := p.elementMap[peek.Node()]
			if !isElement {
				panic("printer: node " + peek.Node() + " not found")
			}

			if len(e.Delimiter) < 1 {
				panic("printer: invalid delimiter")
			}

			delimiter := e.Delimiter[0]

			if ch == delimiter {
				// escape inline delimiter

				b.WriteString(`\`)
			}
		}

		b.WriteByte(ch)
	}

	if trace {
		p.printf("return %q", b.String())
	}

	b.WriteTo(w)
}

// isPunct determines whether ch is an ASCII punctuation character.
func isPunct(ch byte) bool {
	// same as in parser/parser.go
	return ch >= 0x21 && ch <= 0x2F ||
		ch >= 0x3A && ch <= 0x40 ||
		ch >= 0x5B && ch <= 0x60 ||
		ch >= 0x7B && ch <= 0x7E
}

func (p *printer) hasInlineDelimiterPrefix(content []byte) bool {
	for _, e := range p.elementMap {
		if node.TypeCategory(e.Type) == node.CategoryInline {
			// same logic as in parser/parser.go
			delimiter := ""
			runes := utf8.RuneCountInString(e.Delimiter)

			if runes > 0 && e.Type == node.TypePrefixed {
				delimiter = e.Delimiter
			} else if runes == 1 {
				delimiter = e.Delimiter + e.Delimiter
			} else {
				panic("parser: invalid inline delimiter " + e.Delimiter)
			}

			if bytes.HasPrefix(content, []byte(delimiter)) {
				return true
			}
		}
	}

	return false
}

func (p *printer) hasClosingDelimiterPrefix(content []byte) bool {
	for _, delimiter := range p.closingDelimiters() {
		if bytes.HasPrefix(content, []byte(delimiter)) {
			return true
		}
	}

	return false
}

func (p *printer) delimiters() (string, string) {
	if p.n.Node() == "Paragraph" {
		return "", ""
	}

	e, found := p.elementMap[p.n.Node()]
	if !found {
		return "", ""
	}

	var pre, post string

	switch p.n.(type) {
	case node.Block:
		switch m := p.n.(type) {
		case *node.Fenced:
			pre = e.Delimiter
			post = e.Delimiter

			lines := m.Lines()

			var content []byte
			if len(lines) > 1 {
				content = bytes.Join(lines[1:], []byte("\n"))
			}

			if bytes.Contains(content, []byte(e.Delimiter)) {
				// needs escape
				pre += "\\"
				post = "\\" + post
			}
		case node.Ranked:
			rank := m.Rank()

			for i := 0; i < rank; i++ {
				pre += e.Delimiter
			}
		case *node.VerbatimLine, *node.Walled,
			*node.VerbatimWalled, *node.Hanging,
			*node.RankedHanging, *node.Group, *node.Sticky:
			pre = e.Delimiter
		default:
			panic(fmt.Sprintf("printer: unexpected node type %T", p.n))
		}

	case node.Inline:
		switch p.n.(type) {
		case *node.Uniform, *node.Escaped:
			r, _ := utf8.DecodeRuneInString(e.Delimiter)
			counterDelim := counterpart(r)

			pre = e.Delimiter + e.Delimiter
			post = string(counterDelim) + string(counterDelim)
		case *node.Prefixed:
			pre = e.Delimiter
		default:
			panic(fmt.Sprintf("printer: unexpected node type %T", p.n))
		}

		if m, isEscaped := p.n.(*node.Escaped); isEscaped {
			content := m.Content()

			if bytes.Contains(content, []byte(e.Delimiter+e.Delimiter)) {
				// needs escape
				pre += "\\"
				post = "\\" + post
			}
		}

	default:
		panic(fmt.Sprintf("parser: unexpected node type %T", p.n))
	}

	return pre, post
}

func (p *printer) atBOL() bool {
	if p.parent != nil {
		_, isLeaf := p.parent.n.(*node.Leaf)
		return isLeaf && p.pos == 0
	}
	return false
}

func (p *printer) isInline() bool {
	_, ok := p.n.(node.Inline)
	return ok
}

func (p *printer) shouldKeep(n node.Node) bool {
	return doNotRemove(p.elementMap, n) || !isEmpty(n)
}

// doNotRemove returns true if n or any of its children has set DoNotRemove.
func doNotRemove(elementMap ElementMap, n node.Node) bool {
	if e, found := elementMap[n.Node()]; found && e.DoNotRemove {
		return true

	}

	switch n.(type) {
	case node.BlockChildren, node.InlineChildren, node.Content, node.Lines,
		node.Composited, node.Boxed:
	default:
		panic(fmt.Sprintf("printer: unexpected node type %T", n))
	}

	if m, ok := n.(node.Boxed); ok {
		unboxed := m.Unbox()
		if unboxed == nil {
			return false
		}

		return doNotRemove(elementMap, unboxed)
	}

	if m, ok := n.(node.Composited); ok {
		return doNotRemove(elementMap, m.Primary())
	}

	if m, ok := n.(node.BlockChildren); ok {
		for _, c := range m.BlockChildren() {
			if doNotRemove(elementMap, c) {
				return true
			}
		}
	}

	if m, ok := n.(node.InlineChildren); ok {
		for _, c := range m.InlineChildren() {
			if doNotRemove(elementMap, c) {
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

func isEmpty(n node.Node) bool {
	return node.ExtractText(n) == ""
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
