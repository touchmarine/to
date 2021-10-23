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

const trace = false

// Elements maps Elements to Names.
type Elements map[string]Element

type Element struct {
	Name        string
	Type        node.Type
	Delimiter   string
	Matcher     string
	DoNotRemove bool
}

type writer interface {
	io.Writer
	io.ByteWriter
	io.StringWriter
}

func Fprint(w io.Writer, elements Elements, n *node.Node) error {
	p := printer{
		elements: elements,
	}

	if x, ok := w.(writer); ok {
		return p.print(x, n)
	}

	buf := bufio.NewWriter(w)
	if err := p.print(buf, n); err != nil {
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

// hasPreviousPrintableSibling reports whether any previous sibling returns
// p.hasPrintableContent() true.
func (p printer) hasPreviousPrintableSibling(n *node.Node) bool {
	for s := n.PreviousSibling; s != nil; s = s.PreviousSibling {
		if p.hasPrintableContent(s) {
			return true
		}
	}
	return false
}

func (p printer) searchFirstPreviousPrintableSibling(n *node.Node) *node.Node {
	for s := n.PreviousSibling; s != nil; s = s.PreviousSibling {
		if p.hasPrintableContent(s) {
			return s
		}
	}
	return nil
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
//
// Table of number of newlines between blocks:
//
// previous siblings                                                       | relationship          | newlines
// ------------------------------------------------------------------------|-----------------------|---------
// no previous sibling/no printable previous sibling                       | any                   | 0
// printable immediate previous sibling                                    | in group              | 1
// printable immediate previous sibling                                    | not in group          | 2
// non printable immediate previous sibling and printable previous sibling | same previous sibling | 1
// non printable immediate previous sibling and printable previous sibling | in group              | 1
// non printable immediate previous sibling and printable previous sibling | not in group          | 2
//
// print normally places two lines between blocks. It places only one line:
//
//
// Blocks are separated by two lines, except
// - in a group like list or sticky
//   	-a
//      -b
// - or in a group interrupted by an empty block:
//   	-a
//   	> 	<-- treat it as if it is not here
//   	-b
// they are placed on one line.
//
//
// current block relationship                    | newlines
// ----------------------------------------------|---------
// non-printable immediate previous sibling and same non-immediate printable previous sibling | 1
// in group like list or sticky                  | 1
// not in group                                  | 2
func (p printer) print(w writer, n *node.Node) error {
	if n.Type == node.TypeError {
		return fmt.Errorf("error node (%s)", n)
	}

	if !p.hasPrintableContent(n) {
		return nil
	}

	if n.IsBlock() || n.Type == node.TypeContainer {
		if x := p.searchFirstPreviousPrintableSibling(n); x != nil {
			// has non-empty previous sibling
			if x != n.PreviousSibling && x.Type == n.Type && x.Element == n.Element ||
				n.Parent != nil && n.Parent.Type == node.TypeContainer && n.Parent.Element != "" {
				// has the same non-immediate previous sibling
				// (e.g., list split by an empty blockquote) or
				// is in a group like list or sticky
				p.newline(w)
			} else {
				p.newline(w)
				p.writePrefix(w, withoutTrailingSpacing)
				p.newline(w)
			}
		}
	}

	//if (n.IsBlock() || n.Type == node.TypeContainer) && p.hasPreviousPrintableSibling(n) {
	//	if n.Type == node.TypeContainer && n.Element != "" ||
	//		n.Parent != nil && n.Parent.Type == node.TypeContainer && n.Parent.Element != "" {
	//		// is or is in a transformer node like sticky or group
	//		p.newline(w)
	//	} else {
	//		p.newline(w)
	//		p.writePrefix(w, withoutTrailingSpacing)
	//		p.newline(w)
	//	}
	//}

	//if n.IsBlock() || n.Type == node.TypeContainer {
	//	if n.PreviousSibling != nil && p.hasPrintableContent(n.PreviousSibling) {
	//		// has immediate printable previous sibling
	//		if n.Parent != nil && n.Parent.Type == node.TypeContainer && n.Parent.Element != "" {
	//			// in a group like list or sticky
	//			p.newline(w)
	//		} else {
	//			p.newline(w)
	//			p.writePrefix(w, withoutTrailingSpacing)
	//			p.newline(w)
	//		}
	//	} else if x := p.searchFirstPreviousPrintableSibling(n); x != nil {
	//		// has printable previous sibling but not an immediate
	//		// printable previous sibling
	//		//
	//		// example where "-" = list item and ">" = blockquote:
	//		// 	-a
	//		// 	>
	//		// 	-b	<-- we (n) are here
	//		if x.Type == n.Type && x.Element == n.Element {
	//			// same previous sibling, treat it as if
	//			// immediate previous sibling was not there
	//			p.newline(w)
	//		} else {
	//			// different previous sibling
	//			p.newline(w)
	//			p.writePrefix(w, withoutTrailingSpacing)
	//			p.newline(w)
	//		}
	//	}
	//}

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
		if n.Value != "" {
			w.WriteString(" ")
			w.WriteString(strings.Trim(n.Value, " \t"))
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
		lines := strings.Split(n.Value, "\n")
		lines = removeBlankLines(lines)
		for _, line := range lines {
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
		needsEscape := fencedNeedsEscape(n.Value, e.Delimiter)
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

		if n.Value != "" {
			lines := strings.Split(n.Value, "\n")
			for i, line := range lines {
				if i > 0 {
					p.newline(w)
					p.writePrefix(w, withTrailingSpacing)
				}
				w.WriteString(line)
			}
		}

		// closing delimiter
		p.newline(w)
		p.writePrefix(w, withTrailingSpacing)
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

		// closing delimiter
		if p.hasEscapeClashingElementAtEnd(n) {
			// otherwise `**\` would return `**\**`

			p.newline(w)
			p.writePrefix(w, withTrailingSpacing)
		}

		counter := counterpartInString(e.Delimiter)
		w.WriteString(counter + counter)

	case node.TypeEscaped:
		needsEscape := strings.Contains(n.Value, e.Delimiter)

		w.WriteString(e.Delimiter + e.Delimiter)
		if needsEscape {
			w.WriteString(`\`)
		}

		w.WriteString(n.Value)

		// closing delimiter
		if p.hasEscapeClashingElementAtEnd(n) {
			// otherwise `**\` would return `**\**`

			p.newline(w)
			p.writePrefix(w, withTrailingSpacing)
		}

		if needsEscape {
			w.WriteString(`\`)
		}
		counter := counterpartInString(e.Delimiter)
		w.WriteString(counter + counter)

	case node.TypePrefixed:
		w.WriteString(e.Delimiter)
		w.WriteString(n.Value)
	case node.TypeText:
		w.WriteString(p.text(n))

	default:
		return fmt.Errorf("unexpected node type %v (%s)", n.Type, n)
	}

	return nil
}

func (p printer) hasPrintableContent(n *node.Node) bool {
	return p.doNotRemove(n) || n.TextContent() != ""
}

// doNotRemove returns true if n or any of its children has set DoNotRemove.
func (p printer) doNotRemove(n *node.Node) bool {
	if e, found := p.elements[n.Element]; found && e.DoNotRemove {
		return true
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if p.doNotRemove(c) {
			return true
		}
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
				return p.hasBlockDelimiterPrefix(content) &&
					!p.hasInlineDelimiterPrefix(content)

			}
		}
	}
	return false
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

func (p printer) text(n *node.Node) string {
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

		// backslash escape checks
		if ch == '\\' && i+1 < len(content) && content[i+1] == '\\' {
			// A: consecutive backslashes

			b.WriteByte('\\')
		} else if ch == '\\' && i+1 < len(content) && isPunct(content[i+1]) {
			// B: escape backslash so it doesn't escape the
			// following punctuation

			b.WriteByte('\\')
		} else if ch == '\\' && i == len(content)-1 && n.NextSibling != nil && p.hasPrintableContent(n.NextSibling) {
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
		} else if i == len(content)-1 && n.NextSibling != nil && p.hasPrintableContent(n.NextSibling) {
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
