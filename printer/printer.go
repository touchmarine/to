// Package printer allows for printing node trees as Touch formatted text in its
// canonical form.
package printer

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/touchmarine/to/node"
	"github.com/touchmarine/to/parser"
)

type writer interface {
	io.Writer
	io.ByteWriter
	io.StringWriter
}

// Printer prints Touch formatted text in its canonical form. Output depends on
// the values in this struct.
type Printer struct {
	Elements   parser.Elements // element set used to parse the node tree
	LineLength int             // line length to wrap text at
}

// Fprint prints the touch formatted text to the writer.
func (p Printer) Fprint(w io.Writer, n *node.Node) error {
	pp := printer{
		elements:   p.Elements,
		lineLength: p.LineLength,
	}
	if x, ok := w.(writer); ok {
		pp.w = &printerWriter{w: x}
		return pp.print(n)
	}

	buf := bufio.NewWriter(w)
	pp.w = &printerWriter{w: buf}
	if err := pp.print(n); err != nil {
		return err
	}
	return buf.Flush()
}

type printer struct {
	w          *printerWriter
	elements   parser.Elements
	lineLength int

	prefixes       []string // opened block prefixes
	lastPrefixLine int      // last line on which a prefix was written
	line           int      // current line
}

// print prints the node in its canonical form.
//
// Blocks are separated by single lines, except in groups, such as lists or
// stickies, where they are placed immediately one after another.
func (p *printer) print(n *node.Node) error {
	if n.Type == node.TypeError {
		return fmt.Errorf("error node (%s)", n)
	}

	if n.IsBlock() || (n.Type == node.TypeContainer && !isInlineContainer(n)) {
		if n.PreviousSibling != nil {
			if n.Parent != nil && isGroup(n.Parent) {
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

	if p.lineLength > 0 && n.IsInline() && n.Type != node.TypeText && hasDirectPreviousSibling(n) &&
		n.Parent != nil && isGroup(n.Parent) && isFirstChild(n) && containsUnwrappableInline(n.Parent) {
		// In a group (like sticky) that contains an unwrappable element
		// (on the first element in the group):
		// 	If a wrappable element in the group can wrap on its own
		// 	(is in the valid line range), let it wrap. Otherwise,
		// 	wrap the whole group here.
		//
		// Note: This was designed with stickies in mind (so 2 children
		//       only). Any groups with more children were not tested
		//       and will probably not work.
		parent := nodeClone(n.Parent) // clone because we will modify
		x := firstUnwrappableInline(parent)
		if x == nil {
			// must be non-nil because of containsUnwrap... in condition
			panic("nil unwrappable inline")
		}
		if x == parent.FirstChild {
			// unwrappable before
			// pass the line length we want to test (-1 for space that would be added)
			if ln, _, err := p.screenPos(parent, p.lineLength-p.w.screenColumn-1); err != nil {
				return err
			} else if ln > 0 {
				p.newline()
				p.writePrefix(withTrailingSpacing)
			}
		} else {
			// unwrappable after
			for c := x; c != nil; c = c.NextSibling {
				// remove nodes from first unwrappable element
				// onward (including)
				parent.RemoveChild(c)
			}
			// get screen columns of the first sequence of wrappable
			// elements
			if ln, col, err := p.screenPos(parent, p.lineLength); err != nil {
				return err
			} else if ln == 0 && p.w.screenColumn+col+1 <= p.lineLength { // +1 is for space that would be added
				// The first sequence of wrappable elements are
				// too short to wrap themselves, check if the
				// whole group needs to be wrapped (if it
				// exceeds the line length).
				//
				// Note: n.Parent=group so there are no spaces
				//       to account for between the elements.
				pa := nodeClone(n.Parent) // without clone we never end and crash
				if ln, col, err := p.screenPos(pa, p.lineLength); err != nil {
					return err
				} else if ln == 0 && p.w.screenColumn+col+1 > p.lineLength { // +1 is for space that would be added
					p.newline()
					p.writePrefix(withTrailingSpacing)
				}
			}
		}
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
		p.w.WriteString(e.Delimiter)
		if t := n.TextContent(); t != "" {
			p.w.WriteString(strings.TrimRight(t, " \t"))
		}
	case node.TypeWalled:
		defer p.addPrefix(e.Delimiter)()
		p.w.WriteString(e.Delimiter)

		if x := searchFirstNonContainer(n.FirstChild); x != nil {
			p.w.WriteByte(' ')
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if err := p.print(c); err != nil {
					return err
				}
			}
		}
	case node.TypeVerbatimWalled:
		defer p.addPrefix(e.Delimiter)()
		p.w.WriteString(e.Delimiter)
		lines := strings.Split(n.TextContent(), "\n")
		for i, line := range lines {
			if i > 0 {
				p.newline()
				p.writePrefix(withoutTrailingSpacing)
			}
			p.w.WriteString(strings.TrimRight(line, " \t"))
		}
	case node.TypeHanging:
		defer p.addPrefix(" ")()
		p.w.WriteString(e.Delimiter)

		if x := searchFirstNonContainer(n.FirstChild); x != nil {
			p.w.WriteByte(' ')
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

		p.w.WriteString(delimiter)

		if x := searchFirstNonContainer(n.FirstChild); x != nil {
			p.w.WriteByte(' ')
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if err := p.print(c); err != nil {
					return err
				}
			}
		}
	case node.TypeFenced:
		// opening delimiter
		p.w.WriteString(e.Delimiter)
		text := n.TextContent()
		needsEscape := fencedNeedsEscape(text, e.Delimiter)
		if needsEscape {
			p.w.WriteByte('\\')
		}

		if v, ok := n.Data[parser.KeyOpeningText]; ok {
			// opening text on the opening delimiter line
			openingText, isString := v.(string)
			if !isString {
				return fmt.Errorf("openingText is not string (%T %s)", n.Data[parser.KeyOpeningText], n)
			}
			p.w.WriteString(openingText)
		}
		p.newline()
		p.writePrefix(withTrailingSpacing)

		if text != "" {
			lines := strings.Split(text, "\n")
			for _, line := range lines {
				p.w.WriteString(line)
				p.newline()
				p.writePrefix(withTrailingSpacing)
			}
		}

		// closing delimiter
		if needsEscape {
			p.w.WriteByte('\\')
		}
		p.w.WriteString(e.Delimiter)

	case node.TypeLeaf:
		if p.needBlockEscape(n) {
			p.w.WriteByte('\\')
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if err := p.print(c); err != nil {
				return err
			}
		}
	case node.TypeUniform:
		if hasDirectPreviousSibling(n) { // only direct as don't want to separate elements in a group
			// +2 is for delimiter, +1 for space that would be added
			// no need to use rune count for delimiter as it's always 2 chars
			if p.lineLength > 0 && p.w.screenColumn+3 > p.lineLength {
				p.newline()
				p.writePrefix(withTrailingSpacing)
			} else if !p.atStart() {
				p.w.WriteByte(' ')
			}
		}
		p.w.WriteString(e.Delimiter + e.Delimiter)

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if err := p.print(c); err != nil {
				return err
			}
		}

		escapeClashing := p.hasEscapeClashingElementAtEnd(n)
		dd := 2 // closing sequence (delimiter+optional space) length
		if escapeClashing {
			dd++
		}
		if p.lineLength > 0 && hasPreviousSibling(n) && p.w.screenColumn+dd > p.lineLength {
			// doesn't need to be a direct sibling as the closing
			// delimiter cannot separate elements in a group
			p.newline()
			p.writePrefix(withTrailingSpacing)
		}
		if escapeClashing {
			// otherwise `**\` would return `**\**`
			p.w.WriteByte(' ')
		}
		counter := counterpartInString(e.Delimiter)
		p.w.WriteString(counter + counter)

	case node.TypeEscaped:
		text := n.TextContent()
		needsEscape := strings.Contains(text, e.Delimiter)
		// +4 is for delimiters, +1 for space that would be added
		// no need to use rune count for delimiter as it's always 2+2 chars
		ll := p.w.screenColumn + utf8.RuneCountInString(text) + 4 + 1
		if needsEscape {
			ll += 2 // for opening and closing delimiter
		}
		if hasDirectPreviousSibling(n) {
			if p.lineLength > 0 && ll > p.lineLength {
				p.newline()
				p.writePrefix(withTrailingSpacing)
			} else if !p.atStart() {
				p.w.WriteByte(' ')
			}
		}

		p.w.WriteString(e.Delimiter + e.Delimiter)
		if needsEscape {
			p.w.WriteByte('\\')
		}
		p.w.WriteString(text)
		if needsEscape {
			p.w.WriteByte('\\')
		}
		counter := counterpartInString(e.Delimiter)
		p.w.WriteString(counter + counter)

	case node.TypePrefixed:
		t := n.TextContent()
		if hasDirectPreviousSibling(n) {
			dc := utf8.RuneCountInString(e.Delimiter)
			dt := utf8.RuneCountInString(t)
			ll := p.w.screenColumn + dc + dt + 1 // +1 is for space that would be added
			if p.lineLength > 0 && ll > p.lineLength {
				p.newline()
				p.writePrefix(withTrailingSpacing)
			} else if !p.atStart() {
				p.w.WriteByte(' ')
			}
		}
		p.w.WriteString(e.Delimiter)
		p.w.WriteString(t)
	case node.TypeText:
		p.writeText(n)

	default:
		return fmt.Errorf("unexpected node type %v (%s)", n.Type, n)
	}

	return nil
}

// screenPos prints the given node and returns the line and screen column on
// which the printer finished. It is used to determine whether the given node
// needs to be wrapped.
func (p printer) screenPos(n *node.Node, lineLength int) (int, int, error) {
	var b strings.Builder
	pp := &printer{
		w: &printerWriter{
			w: &b,
		},
		elements:   p.elements,
		lineLength: lineLength,
	}
	if err := pp.print(n); err != nil {
		return 0, 0, err
	}
	return pp.line, pp.w.screenColumn, nil
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

func hasDirectPreviousSibling(n *node.Node) bool {
	if hasPreviousSibling(n) {
		if n.Parent != nil && isGroup(n.Parent) && !isFirstChild(n) {
			return false
		}
		return true
	}
	return false
}

// hasPreviousSibling reports if the current node has a previous sibling, even
// if it is placed inside a group.
func hasPreviousSibling(n *node.Node) bool {
	if n.Parent != nil && isGroup(n.Parent) {
		// in group like sticky
		return hasPreviousSibling(n.Parent)
	}
	return n.PreviousSibling != nil
}

func isFirstChild(n *node.Node) bool {
	return n.Parent != nil && n == n.Parent.FirstChild
}

func isGroup(n *node.Node) bool {
	return n.Type == node.TypeContainer && n.Element != ""
}

func containsUnwrappableInline(n *node.Node) bool {
	return firstUnwrappableInline(n) != nil
}

func firstUnwrappableInline(n *node.Node) *node.Node {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if !isWrappableInline(c) {
			return c
		}
	}
	return nil
}

func isWrappableInline(n *node.Node) bool {
	return n.IsInline() && n.Type != node.TypeEscaped && n.Type != node.TypePrefixed
}

func (p *printer) newline() {
	p.w.WriteString("\n")
	p.line++
	p.w.textColumn = 0
	p.w.screenColumn = 0
}

type prefixSpacing int

const (
	withTrailingSpacing prefixSpacing = iota
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
	case withTrailingSpacing:
		prefix += " "
	case withoutTrailingSpacing:
		prefix = strings.Trim(prefix, " \t")
	default:
		panic(fmt.Sprintf("printer: unexpected spacing state (%d)", spacing))
	}
	p.w.WriteString(prefix)
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

func (p *printer) writeText(n *node.Node) {
	if n.Type != node.TypeText {
		panic(fmt.Sprintf("printer: expected text node (%s)", n))
	}

	// By trimming the text we disregard spacing around the word:
	// 	s**tro**ng -> s **tro** ng	?
	v := strings.Trim(n.Value, " \t")
	if v == "" {
		return
	}

	// use strings.Builder instead of bytes.Buffer as it's safer (can't
	// access underlying bytes, no unreads); con: doesn't reuse the memory
	// on reset (but no premature optimizations)
	var buf strings.Builder
	var sep byte                             // last word separator character (0|' '|'\t'|'\n')
	prependSpace := n.PreviousSibling != nil // whether should add a space/newline before this text node
	onlyPunct := containsOnlyPunct(v)
	if onlyPunct {
		// only punctuation
		prependSpace = false
	}

	for i := 0; i < len(v); i++ {
		ch := v[i]

		if ch == ' ' || ch == '\t' || ch == '\n' {
			// no need to be included in below escape checks as ' ',
			// '\t', or '\n' cannot be a part of a delimiter
			if prependSpace && sep != 0 {
				panic("prependSpace can only be true once-before separator is set")
			}
			if p.lineLength <= 0 && ch == '\n' {
				// undefined line length
				if buf.Len() > 0 {
					s := buf.String()
					if !containsOnlyPunct(s) {
						if prependSpace {
							p.w.WriteByte(' ')
						} else if sep > 0 { // cannot be '\n' as we catch it here and set sep=0
							p.w.WriteByte(sep)
						}
					}
					p.w.WriteString(s)
					buf.Reset()
				}
				if prependSpace {
					prependSpace = false
				}
				p.newline()
				if p.needsBlockEscape(string(v[i+1:])) {
					p.w.WriteByte('\\')
				}
				p.writePrefix(withTrailingSpacing)
				sep = 0
				continue
			}

			if buf.Len() > 0 {
				s := buf.String()
				if !containsOnlyPunct(s) {
					if prependSpace {
						if p.lineLength > 0 {
							// defined line length
							p.wrap(' ', s)
						} else {
							// undefined line length
							p.w.WriteByte(' ')
						}
					} else if p.lineLength > 0 {
						// defined line length
						p.wrap(sep, s)
					} else if sep > 0 {
						p.w.WriteByte(sep)
					}
				}
				p.w.WriteString(s)
				buf.Reset()
			}
			if prependSpace {
				prependSpace = false
			}
			sep = ch
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

		if escape {
			buf.WriteByte('\\')
		}
		if escape2 {
			buf.WriteByte('\\')
		}
		buf.WriteByte(ch)
	}

	if prependSpace && sep != 0 {
		panic("prependSpace can only be true once-before separator is set")
	}
	if buf.Len() > 0 {
		s := buf.String()
		if !containsOnlyPunct(s) {
			if prependSpace {
				if buf.Len() == 0 {
					// must be something buffered as v != "" and a separator has not
					// been reached
					panic("buffered is empty")
				}
				if p.lineLength > 0 {
					// defined line linegth
					p.wrap(' ', s)
				} else {
					// undefined line length
					p.w.WriteByte(' ')
				}
			} else if p.lineLength > 0 {
				// defined line length
				p.wrap(sep, s)
			} else if sep > 0 {
				p.w.WriteByte(sep)

			}
		}
		p.w.WriteString(s)
		buf.Reset()
	}
	if prependSpace {
		prependSpace = false // not used anymore but still want to keep correct state
	}
}

// containsOnlyPunct reports whether the given string contains only Unicode
// punctuation characters.
func containsOnlyPunct(s string) bool {
	for _, r := range s {
		if !unicode.IsPunct(r) {
			return false
		}
	}
	return true
}

// wrap breaks the line if it determines that adding the last word (+ a space if
// not a single line) would result in the current line being over the given line
// length.
//
// Words are never split (a word that is longer than the line length will not be
// split and will stay as is).
func (p *printer) wrap(sep byte, word string) {
	n := utf8.RuneCountInString(word) // last word
	if sep > 0 {
		n++
	}
	m := p.w.screenColumn + n
	if m > p.lineLength && !p.atStart() {
		// !atStart so '*a' !=> '*\n a' (lineLength=1)
		p.newline()
		p.writePrefix(withTrailingSpacing)
		if p.needsBlockEscape(word) {
			// escape the content that will be flushed below
			p.w.WriteByte('\\')
		}
	} else if sep > 0 {
		p.w.WriteByte(' ') // flush previous space (in lineLength mode newline=space)
	}
}

// atStart reports whether the current column is right after the current line's
// prefix. It assumes that the current line's prefix was withTrailingSpacing
// (only TypeVerbatimWalled does not use withTrailingSpacing).
func (p printer) atStart() bool {
	pl := 0
	if len(p.prefixes) > 0 {
		pl = len(strings.Join(p.prefixes, " ")) + 1 // +1 is for trailing space
	}
	return p.w.textColumn == pl
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

type printerWriter struct {
	w            writer
	textColumn   int // byte number (zero-based)
	screenColumn int // utf8 number (zero-based)
}

func (w *printerWriter) Write(p []byte) (int, error) {
	n, err := w.w.Write(p)
	if err != nil {
		return n, err
	}
	w.textColumn += len(p)
	w.screenColumn += utf8.RuneCount(p)
	return n, nil
}

func (w *printerWriter) WriteString(s string) (int, error) {
	n, err := w.w.WriteString(s)
	if err != nil {
		return n, err
	}
	w.textColumn += len(s)
	w.screenColumn += utf8.RuneCountInString(s)
	return n, nil
}

func (w *printerWriter) WriteByte(b byte) error {
	if err := w.w.WriteByte(b); err != nil {
		return err
	}
	w.textColumn++
	w.screenColumn++
	return nil
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

// nodeClone deep clones the given node but is detached from the parent.
func nodeClone(n *node.Node) *node.Node {
	if n == nil {
		return nil
	}
	nn := &node.Node{
		Element:  n.Element,
		Type:     n.Type,
		Data:     n.Data,
		Value:    n.Value,
		Location: n.Location,
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		nn.AppendChild(nodeClone(c))
	}
	return nn
}
