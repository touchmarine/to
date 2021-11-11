package printer

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/touchmarine/to/node"
	"github.com/touchmarine/to/parser"
)

// Elements maps Elements to Names.
type Elements map[string]Element

type Element struct {
	Name      string
	Type      node.Type
	Delimiter string
	Matcher   string
}

type writer interface {
	io.Writer
	io.ByteWriter
	io.StringWriter
}

type Printer struct {
	Elements Elements
}

func (p Printer) Fprint(w io.Writer, n *node.Node) error {
	pp := printer{
		elements: p.Elements,
	}
	if x, ok := w.(writer); ok {
		return pp.print(x, n)
	}

	buf := bufio.NewWriter(w)
	if err := pp.print(buf, n); err != nil {
		return err
	}
	return buf.Flush()
}

type printer struct {
	elements Elements

	prefixes       []string
	lastPrefixLine int
	line           int
}

// print prints the node in its canonical form.
//
// Blocks are separated by single lines, except in groups, such as lists or
// stickies, where they are placed immediately one after another. Empty blocks
// are treated as if they were not present. For example, a group interrupted by
// an empty block is printed as a cohesive group:
//   	-a
//   	> 	<- as if it was not here
//   	-b
func (p printer) print(w writer, n *node.Node) error {
	if n.Type == node.TypeError {
		return fmt.Errorf("error node (%s)", n)
	}

	if n.IsBlock() || (n.Type == node.TypeContainer && !isInlineContainer(n)) {
		if n.PreviousSibling != nil {
			if n.Parent != nil && n.Parent.Type == node.TypeContainer && n.Parent.Element != "" {
				// is in a group like list or sticky
				p.newline(w)
			} else {
				p.newline(w)
				p.writePrefix(w, withoutTrailingSpacing)
				p.newline(w)
			}
		}
	}

	if n.IsBlock() {
		p.writePrefix(w, withTrailingSpacing)
	}

	e, _ := p.elements[n.Element]
	switch n.Type {
	case node.TypeContainer:
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if err := p.print(w, c); err != nil {
				return err
			}
		}
	case node.TypeVerbatimLine:
		w.WriteString(e.Delimiter)
		if text := n.TextContent(); text != "" {
			w.WriteString(" ")
			w.WriteString(strings.Trim(text, " \t"))
		}
	case node.TypeWalled:
		defer p.addPrefix(e.Delimiter)()
		w.WriteString(e.Delimiter)

		buf := &bytes.Buffer{}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if err := p.print(buf, c); err != nil {
				return err
			}
		}
		if buf.Len() > 0 {
			w.WriteString(" ")
			if _, err := buf.WriteTo(w); err != nil {
				return err
			}
		}

	case node.TypeVerbatimWalled:
		defer p.addPrefix(e.Delimiter)()
		w.WriteString(e.Delimiter)
		lines := strings.Split(n.TextContent(), "\n")
		lines = removeBlankLines(lines)
		for i, line := range lines {
			if i > 0 {
				p.newline(w)
				p.writePrefix(w, withoutTrailingSpacing)
			}
			w.WriteString(" ")
			w.WriteString(strings.Trim(line, " \t"))
		}
	case node.TypeHanging:
		defer p.addPrefix(" ")()
		w.WriteString(e.Delimiter)

		buf := &bytes.Buffer{}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if err := p.print(buf, c); err != nil {
				return err
			}
		}
		if buf.Len() > 0 {
			w.WriteString(" ")
			if _, err := buf.WriteTo(w); err != nil {
				return err
			}
		}

	case node.TypeRankedHanging:
		var delimiter string
		if v, ok := n.Data[parser.KeyRank]; ok {
			rank, isInt := v.(int)
			if !isInt {
				return fmt.Errorf("rank is not int (%T %s)", n.Data[parser.KeyRank], n)
			}

			delimiter = strings.Repeat(e.Delimiter, rank)
		} else {
			delimiter = e.Delimiter
		}

		prefix := strings.Repeat(" ", len(delimiter))
		defer p.addPrefix(prefix)()

		w.WriteString(delimiter)

		buf := &bytes.Buffer{}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if err := p.print(buf, c); err != nil {
				return err
			}
		}
		if buf.Len() > 0 {
			w.WriteString(" ")
			if _, err := buf.WriteTo(w); err != nil {
				return err
			}
		}
	case node.TypeFenced:
		// opening delimiter
		w.WriteString(e.Delimiter)
		text := n.TextContent()
		needsEscape := fencedNeedsEscape(text, e.Delimiter)
		if needsEscape {
			w.WriteString(`\`)
		}

		if v, ok := n.Data[parser.KeyOpeningText]; ok {
			// opening text on the opening delimiter line
			openingText, isString := v.(string)
			if !isString {
				return fmt.Errorf("openingText is not string (%T %s)", n.Data[parser.KeyOpeningText], n)
			}
			w.WriteString(openingText)
		}
		p.newline(w)
		p.writePrefix(w, withTrailingSpacing)

		if text != "" {
			lines := strings.Split(text, "\n")
			for _, line := range lines {
				w.WriteString(line)
				p.newline(w)
				p.writePrefix(w, withTrailingSpacing)
			}
		}

		// closing delimiter
		if needsEscape {
			w.WriteString(`\`)
		}
		w.WriteString(e.Delimiter)

	case node.TypeLeaf:
		if p.needBlockEscape(n) {
			w.WriteString(`\`)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if err := p.print(w, c); err != nil {
				return err
			}
		}
	case node.TypeUniform:
		w.WriteString(e.Delimiter + e.Delimiter)

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if err := p.print(w, c); err != nil {
				return err
			}
		}

		if p.hasEscapeClashingElementAtEnd(n) {
			// otherwise `**\` would return `**\**`
			w.WriteByte(' ')
		}
		counter := counterpartInString(e.Delimiter)
		w.WriteString(counter + counter)

	case node.TypeEscaped:
		text := n.TextContent()
		needsEscape := strings.Contains(text, e.Delimiter)
		w.WriteString(e.Delimiter + e.Delimiter)
		if needsEscape {
			w.WriteString(`\`)
		}

		w.WriteString(text)

		if p.hasEscapeClashingElementAtEnd(n) {
			// otherwise `**\` would return `**\**`
			w.WriteByte(' ')
		}
		if needsEscape {
			w.WriteString(`\`)
		}
		counter := counterpartInString(e.Delimiter)
		w.WriteString(counter + counter)

	case node.TypePrefixed:
		w.WriteString(e.Delimiter)
		w.WriteString(n.TextContent())
	case node.TypeText:
		w.WriteString(p.text(w, n))

	default:
		return fmt.Errorf("unexpected node type %v (%s)", n.Type, n)
	}

	return nil
}

// isInlineContainer reports whether a container is inline based on its
// children.
func isInlineContainer(n *node.Node) bool {
	if n == nil {
		panic("n is nil")
	}
	if n.Type != node.TypeContainer {
		panic("n is not type container")
	}
	if n.IsInline() {
		return true
	}
	if n.FirstChild != nil && n.FirstChild.IsInline() {
		return true
	}
	return false
}

func (p *printer) newline(w writer) {
	w.WriteString("\n")
	p.line++
}

type prefixSpacing int

const (
	withoutSpacing prefixSpacing = iota
	withTrailingSpacing
	withoutTrailingSpacing
)

func (p *printer) writePrefix(w writer, spacing prefixSpacing) {
	if p.lastPrefixLine == p.line {
		return
	}
	p.lastPrefixLine = p.line

	if len(p.prefixes) == 0 {
		return
	}

	prefix := strings.Join(p.prefixes, " ")
	switch spacing {
	case withoutSpacing:
	case withTrailingSpacing:
		prefix += " "
	case withoutTrailingSpacing:
		prefix = strings.Trim(prefix, " \t")
	default:
		panic(fmt.Sprintf("printer: unexpected spacing state (%d)", spacing))
	}
	w.WriteString(prefix)
}

func removeBlankLines(p []string) []string {
	var n []string
	for _, s := range p {
		if strings.Trim(s, " \t") != "" {
			n = append(n, s)
		}
	}
	return n
}

func fencedNeedsEscape(s, delimiter string) bool {
	lines := strings.Split(s, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, delimiter) {
			return true
		}
	}
	return false

}

func (p printer) needBlockEscape(n *node.Node) bool {
	if n.Type != node.TypeLeaf {
		panic(fmt.Sprintf("printer: expected leaf node (%s)", n))
	}
	if n.FirstChild != nil {
		if x := searchFirstNonContainer(n.FirstChild); x != nil && x.Type == node.TypeText {
			// we only need to check a text node, any other node
			// will have it's own non-block delimiter
			if content := x.Value; content != "" {
				return p.needsBlockEscape(content)

			}
		}
	}
	return false
}

func (p printer) needsBlockEscape(s string) bool {
	return p.hasBlockDelimiterPrefix(s) &&
		!p.hasInlineDelimiterPrefix(s)
}

func (p printer) hasBlockDelimiterPrefix(s string) bool {
	for _, e := range p.elements {
		if node.HasDelimiter(e.Type) && node.IsBlock(e.Type) {
			// same logic as in parser/parser.go
			delimiter := ""
			if e.Type == node.TypeRankedHanging {
				delimiter = e.Delimiter + e.Delimiter
			} else {
				delimiter = e.Delimiter
			}

			if strings.HasPrefix(s, delimiter) {
				return true
			}
		}
	}
	return false
}

func (p printer) hasInlineDelimiterPrefix(s string) bool {
	for _, e := range p.elements {
		if node.HasDelimiter(e.Type) && node.IsInline(e.Type) {
			// same logic as in parser/parser.go
			delimiter := ""
			runes := utf8.RuneCountInString(e.Delimiter)
			if runes > 0 && e.Type == node.TypePrefixed {
				delimiter = e.Delimiter
			} else if runes == 1 {
				delimiter = e.Delimiter + e.Delimiter
			} else {
				panic(fmt.Sprintf(
					"printer: invalid inline delimiter %q (%s %s)",
					e.Delimiter, e.Type, e.Name,
				))
			}

			if strings.HasPrefix(s, delimiter) {
				return true
			}
		}
	}
	return false
}

func (p printer) text(w writer, n *node.Node) string {
	if n.Type != node.TypeText {
		panic(fmt.Sprintf("printer: expected text node (%s)", n))
	}

	content := n.Value
	if content == "" {
		return ""
	}

	var b strings.Builder
	for i := 0; i < len(content); i++ {
		ch := content[i]

		if ch == '\n' {
			p.newline(&b)
			// i+1 must exist (so no need to bound check)
			//
			// must exist because parser doesn't leave newlines at
			// the end of text
			if p.needsBlockEscape(string(content[i+1:])) {
				b.WriteByte('\\')
			}
			p.writePrefix(&b, withTrailingSpacing)
			continue
		}

		// backslash escape checks
		if ch == '\\' && i+1 < len(content) && content[i+1] == '\\' {
			// A: consecutive backslashes

			b.WriteByte('\\')
		} else if ch == '\\' && i+1 < len(content) && isPunct(content[i+1]) {
			// B: escape backslash so it doesn't escape the
			// following punctuation

			b.WriteByte('\\')
		} else if ch == '\\' && i == len(content)-1 && n.NextSibling != nil {
			// C: escape backslash so it doesn't escape the
			// following non-emtpy inline element

			b.WriteByte('\\')
		} else if ch == '\\' && i+1 < len(content) && p.hasInlineDelimiterPrefix(content[i+1:]) {
			// D: escape backslash so it doesn't escape the
			// following inline delimiter
			//
			// needed for non-punctuation delimiters that aren't
			// caught by escape-B; as such don't need to check
			// the closing delimiters as non-punctuation can only be
			// prefixed elements

			b.WriteByte('\\')
		} else if p.hasInlineDelimiterPrefix(content[i:]) || p.hasClosingDelimiterPrefix(n, content[i:]) {
			// E: escape inline delimiter

			b.WriteByte('\\')
		} else if i == len(content)-1 && n.NextSibling != nil {
			// F: last character and the following non-empty
			// element's delimiter's first character may form an
			// inline delimiter

			if x := searchFirstNonContainer(n.NextSibling); x != nil {
				e, _ := p.elements[x.Element]
				if e.Delimiter != "" && ch == e.Delimiter[0] {
					// escape inline delimiter
					b.WriteByte('\\')
				}
			}

		}

		b.WriteByte(ch)
	}

	text := b.String()
	if parent := searchFirstNonContainerParent(n.Parent); parent != nil &&
		(parent.Type == node.TypeUniform || parent.Type == node.TypeEscaped) {
		// consider parent closing delimiter

		e, _ := p.elements[parent.Element]
		counter := counterpartInString(e.Delimiter)
		closingDelimiter := counter + counter
		if parent.Type == node.TypeEscaped && strings.Contains(parent.Value, e.Delimiter+e.Delimiter) {
			closingDelimiter = `\` + closingDelimiter
		}

		if len(text) == 1 && text[0] == '\\' ||
			len(text) > 1 && text[len(text)-2] != '\\' && text[len(text)-1] == '\\' {
			// unescaped "\" at the end of text

			text += `\`
		} else if len(text) > 0 && closingDelimiter != "" && text[len(text)-1] == closingDelimiter[0] {
			// closing delimiter character at end of text

			text = text[:len(text)-1] + `\` + text[len(text)-1:]
		}
	}

	return text
}

func containsInt(ii []int, x int) bool {
	for _, i := range ii {
		if x == i {
			return true
		}
	}
	return false
}

func searchFirstNonContainerParent(n *node.Node) *node.Node {
	if n.Type != node.TypeContainer {
		return n
	}

	for p := n.Parent; p != nil; p = p.Parent {
		if x := searchFirstNonContainerParent(p); x != nil {
			return x
		}
	}
	return nil
}

// isPunct determines whether ch is an ASCII punctuation character.
func isPunct(ch byte) bool {
	// same as in parser/parser.go
	return ch >= 0x21 && ch <= 0x2F ||
		ch >= 0x3A && ch <= 0x40 ||
		ch >= 0x5B && ch <= 0x60 ||
		ch >= 0x7B && ch <= 0x7E
}

func (p printer) hasClosingDelimiterPrefix(n *node.Node, s string) bool {
	var closingDelimiters []string
	for m := n; m != nil && !m.IsBlock(); m = m.Parent {
		if m.Type == node.TypeUniform || m.Type == node.TypeEscaped {
			e, _ := p.elements[m.Element]
			counter := counterpartInString(e.Delimiter)
			closingDelimiter := counter + counter
			if m.Type == node.TypeEscaped && strings.Contains(m.Value, e.Delimiter+e.Delimiter) {
				closingDelimiter = `\` + closingDelimiter
			}

			closingDelimiters = append(closingDelimiters, closingDelimiter)
		}
	}

	for _, delimiter := range closingDelimiters {
		if strings.HasPrefix(s, delimiter) {
			return true
		}
	}
	return false
}

// hasEscapeClashingElementAtEnd returns true if the last child (inline) has the
// escape character as a delimiter.
func (p printer) hasEscapeClashingElementAtEnd(n *node.Node) bool {
	if n.LastChild != nil {
		if x := searchFirstNonContainer(n.LastChild); x != nil {
			e, ok := p.elements[x.Element]
			return ok && e.Type == node.TypePrefixed && e.Delimiter == "\\"
		}
	}
	return false
}

func searchFirstNonContainer(n *node.Node) *node.Node {
	if n.Type != node.TypeContainer {
		return n
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if x := searchFirstNonContainer(c); x != nil {
			return x
		}
	}
	return nil
}

func (p *printer) addPrefix(s string) func() {
	size := len(p.prefixes)
	p.prefixes = append(p.prefixes, s)
	return func() {
		p.prefixes = p.prefixes[:size]
	}
}

func counterpartInString(s string) string {
	r, _ := utf8.DecodeRuneInString(s)
	return string(counterpart(r))
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
