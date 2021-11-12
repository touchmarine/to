package printer

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/touchmarine/to/node"
	"github.com/touchmarine/to/parser"
)

const tabWidth = 8

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
	Elements   Elements
	LineLength int // line length to wrap text at
}

func (p Printer) Fprint(w io.Writer, n *node.Node) error {
	pp := printer{
		elements:   p.Elements,
		lineLength: p.LineLength,
	}
	if x, ok := w.(writer); ok {
		pp.w = x
		return pp.print(n)
	}

	buf := bufio.NewWriter(w)
	pp.w = buf
	if err := pp.print(n); err != nil {
		return err
	}
	return buf.Flush()
}

type printer struct {
	w          writer
	elements   Elements
	lineLength int

	prefixes       []string // opened block prefixes
	lastPrefixLine int      // last line on which a prefix was written
	line           int      // current line
	textColumn     int      // byte number (zero-based)
	screenColumn   int      // utf8 number (zero-based, tab=tabWidth)
}

// print prints the node in its canonical form.
//
// Blocks are separated by single lines, except in groups, such as lists or
// stickies, where they are placed immediately one after another.
func (p printer) print(n *node.Node) error {
	if n.Type == node.TypeError {
		return fmt.Errorf("error node (%s)", n)
	}

	if n.IsBlock() || (n.Type == node.TypeContainer && !isInlineContainer(n)) {
		if n.PreviousSibling != nil {
			if n.Parent != nil && n.Parent.Type == node.TypeContainer && n.Parent.Element != "" {
				// is in a group like list or sticky
				p.newline()
			} else {
				p.newline()
				p.writePrefix(withoutTrailingSpacing)
				p.newline()
			}
		}
	}

	if n.IsBlock() {
		p.writePrefix(withTrailingSpacing)
	}

	e, _ := p.elements[n.Element]
	switch n.Type {
	case node.TypeContainer:
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if err := p.print(c); err != nil {
				return err
			}
		}
	case node.TypeVerbatimLine:
		p.writeString(e.Delimiter)
		if t := n.TextContent(); t != "" {
			p.writeString(strings.TrimRight(t, " \t"))
		}
	case node.TypeWalled:
		defer p.addPrefix(e.Delimiter)()
		p.writeString(e.Delimiter)

		if x := searchFirstNonContainer(n.FirstChild); x != nil {
			p.writeByte(' ')
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if err := p.print(c); err != nil {
					return err
				}
			}
		}
	case node.TypeVerbatimWalled:
		defer p.addPrefix(e.Delimiter)()
		p.writeString(e.Delimiter)
		lines := strings.Split(n.TextContent(), "\n")
		for i, line := range lines {
			if i > 0 {
				p.newline()
				p.writePrefix(withoutTrailingSpacing)
			}
			p.writeString(strings.TrimRight(line, " \t"))
		}
	case node.TypeHanging:
		defer p.addPrefix(" ")()
		p.writeString(e.Delimiter)

		if x := searchFirstNonContainer(n.FirstChild); x != nil {
			p.writeByte(' ')
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if err := p.print(c); err != nil {
					return err
				}
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

		p.writeString(delimiter)

		if x := searchFirstNonContainer(n.FirstChild); x != nil {
			p.writeByte(' ')
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if err := p.print(c); err != nil {
					return err
				}
			}
		}
	case node.TypeFenced:
		// opening delimiter
		p.writeString(e.Delimiter)
		text := n.TextContent()
		needsEscape := fencedNeedsEscape(text, e.Delimiter)
		if needsEscape {
			p.writeByte('\\')
		}

		if v, ok := n.Data[parser.KeyOpeningText]; ok {
			// opening text on the opening delimiter line
			openingText, isString := v.(string)
			if !isString {
				return fmt.Errorf("openingText is not string (%T %s)", n.Data[parser.KeyOpeningText], n)
			}
			p.writeString(openingText)
		}
		p.newline()
		p.writePrefix(withTrailingSpacing)

		if text != "" {
			lines := strings.Split(text, "\n")
			for _, line := range lines {
				p.writeString(line)
				p.newline()
				p.writePrefix(withTrailingSpacing)
			}
		}

		// closing delimiter
		if needsEscape {
			p.writeByte('\\')
		}
		p.writeString(e.Delimiter)

	case node.TypeLeaf:
		if p.needBlockEscape(n) {
			p.writeByte('\\')
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if err := p.print(c); err != nil {
				return err
			}
		}
	case node.TypeUniform:
		p.writeString(e.Delimiter + e.Delimiter)

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if err := p.print(c); err != nil {
				return err
			}
		}

		if p.hasEscapeClashingElementAtEnd(n) {
			// otherwise `**\` would return `**\**`
			p.writeByte(' ')
		}
		counter := counterpartInString(e.Delimiter)
		p.writeString(counter + counter)

	case node.TypeEscaped:
		text := n.TextContent()
		needsEscape := strings.Contains(text, e.Delimiter)
		p.writeString(e.Delimiter + e.Delimiter)
		if needsEscape {
			p.writeByte('\\')
		}

		p.writeString(text)

		if p.hasEscapeClashingElementAtEnd(n) {
			// otherwise `**\` would return `**\**`
			p.writeByte(' ')
		}
		if needsEscape {
			p.writeByte('\\')
		}
		counter := counterpartInString(e.Delimiter)
		p.writeString(counter + counter)

	case node.TypePrefixed:
		p.writeString(e.Delimiter)
		p.writeString(n.TextContent())
	case node.TypeText:
		p.writeText(n)

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

func (p *printer) newline() {
	p.writeString("\n")
	p.line++
	p.textColumn = 0
	p.screenColumn = 0
}

type prefixSpacing int

const (
	withoutSpacing prefixSpacing = iota
	withTrailingSpacing
	withoutTrailingSpacing
)

func (p *printer) writePrefix(spacing prefixSpacing) {
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
	p.writeString(prefix)
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

func (p printer) writeText(n *node.Node) {
	if n.Type != node.TypeText {
		panic(fmt.Sprintf("printer: expected text node (%s)", n))
	}

	v := n.Value
	if v == "" {
		return
	}

	for i := 0; i < len(v); i++ {
		ch := v[i]

		if ch == '\n' {
			if p.lineLength <= 0 || p.lineLength > 0 && p.textColumn <= p.lineLength {
				// undefined line length or a defined line
				// length but the current newline is within the
				// allowed length
				p.newline()
				// i+1 must exist (so no need to bound check)
				//
				// must exist because parser doesn't leave newlines at
				// the end of text
				if p.needsBlockEscape(string(v[i+1:])) {
					p.writeByte('\\')
				}
				p.writePrefix(withTrailingSpacing)
			}
			continue
		}

		escape := false
		// backslash escape checks
		if ch == '\\' && i+1 < len(v) && v[i+1] == '\\' {
			// A: consecutive backslashes
			escape = true
		} else if ch == '\\' && i+1 < len(v) && isPunct(v[i+1]) {
			// B: escape backslash so it doesn't escape the
			// following punctuation
			escape = true
		} else if ch == '\\' && i == len(v)-1 && n.NextSibling != nil {
			// C: escape backslash so it doesn't escape the
			// following non-emtpy inline element
			escape = true
		} else if ch == '\\' && i+1 < len(v) && p.hasInlineDelimiterPrefix(v[i+1:]) {
			// D: escape backslash so it doesn't escape the
			// following inline delimiter
			//
			// needed for non-punctuation delimiters that aren't
			// caught by escape-B; as such don't need to check
			// the closing delimiters as non-punctuation can only be
			// prefixed elements
			escape = true
		} else if p.hasInlineDelimiterPrefix(v[i:]) || p.hasClosingDelimiterPrefix(n, v[i:]) {
			// E: escape inline delimiter
			escape = true
		} else if i == len(v)-1 && n.NextSibling != nil {
			// F: last character and the following non-empty
			// element's delimiter's first character may form an
			// inline delimiter

			if x := searchFirstNonContainer(n.NextSibling); x != nil {
				e, _ := p.elements[x.Element]
				if e.Delimiter != "" && ch == e.Delimiter[0] {
					// escape inline delimiter
					escape = true
				}
			}
		}

		escape2 := false // whether a second escape is needed
		if i == len(v)-1 {
			// last character
			if parent := searchFirstNonContainerParent(n.Parent); parent != nil &&
				(parent.Type == node.TypeUniform || parent.Type == node.TypeEscaped) {
				// consider parent closing delimiter

				e, _ := p.elements[parent.Element]
				counter := counterpartInString(e.Delimiter)
				closingDelimiter := counter + counter
				if parent.Type == node.TypeEscaped && strings.Contains(parent.Value, e.Delimiter+e.Delimiter) {
					closingDelimiter = `\` + closingDelimiter
				}

				if !escape && ch == '\\' || closingDelimiter != "" &&
					(escape && closingDelimiter[0] == '\\' || !escape && closingDelimiter[0] == ch) {
					// unescaped '\' or a closing delimiter character at end of text
					escape2 = true
				}
			}
		}

		if p.lineLength > 0 {
			// defined line length
			ll := p.textColumn + 1 // would be current line length after this iteration
			if escape {
				ll++
			}
			if escape2 {
				ll++
			}

			if ll > p.lineLength {
				p.newline()
				// i+1 must exist: look above
				if p.needsBlockEscape(string(v[i+1:])) {
					p.writeByte('\\')
				}
				p.writePrefix(withTrailingSpacing)
			}
		}

		if escape {
			p.writeByte('\\')
		}
		if escape2 {
			p.writeByte('\\')
		}
		p.writeByte(ch)
	}
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

func (p *printer) writeString(s string) {
	p.w.WriteString(s)
	p.textColumn += len(s)
	n := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\t' {
			// -1 because rune count already counts tabs as 1
			n += tabWidth - 1
		}
	}
	p.screenColumn = utf8.RuneCountInString(s) + n
}

func (p *printer) writeByte(b byte) {
	p.w.WriteByte(b)
	p.textColumn++
	if b == '\t' {
		p.screenColumn += tabWidth
	} else {
		p.screenColumn++
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
