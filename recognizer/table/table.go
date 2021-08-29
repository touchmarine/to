package table

import (
	"bytes"
	"fmt"
	"github.com/touchmarine/to/config"
	"github.com/touchmarine/to/matcher"
	"github.com/touchmarine/to/node"
	"io"
	"strings"
	"unicode/utf8"
)

const trace = false

func Recognize(w io.Writer, src []byte, elements []config.Element) {
	var r recognizer
	r.w = w
	r.Matchers(matcher.Defaults())
	r.register(elements)
	r.init(src)
	r.recognize()
}

type recognizer struct {
	src            []byte                    // source
	inlineMap      map[string]config.Element // registered inline elements by delimiter
	specialEscapes []string                  // delimiters that do not start with a punctuation
	matcherMap     matcher.Map               // registered matchers by name

	// parsing
	ch         rune // current character
	offset     int  // character offset
	rdOffset   int  // position after current character
	lineOffset int  // current line offset

	inlines          []rune // open inlines
	separatorOffsets []int  // "|" separator offsets in current line

	// tracing
	indent int // trace indentation

	w io.Writer
}

func (r *recognizer) register(elems []config.Element) {
	if r.inlineMap == nil {
		r.inlineMap = make(map[string]config.Element)
	}

	for _, e := range elems {
		switch c := node.TypeCategory(e.Type); c {
		case node.CategoryBlock:
			// not needed

		case node.CategoryInline:
			r.inlineMap[e.Delimiter] = e

			ch, _ := utf8.DecodeRuneInString(e.Delimiter)
			if !isPunct(ch) {
				r.specialEscapes = append(r.specialEscapes, e.Delimiter)
			}

		default:
			panic("recognizer: unexpected node category " + c.String())
		}
	}
}

func (r *recognizer) recognize() {
	if trace {
		defer r.trace("recognize")()
	}

	for r.ch > 0 {
		if isSpacing(r.ch) {
			r.parseSpacing()
		} else if r.ch == '\n' {
			r.process()

			r.next()
			r.lineOffset = r.offset
		} else {
			r.parseInlines()
		}
	}

	r.process()
}

func (r *recognizer) process() {
	if trace {
		defer r.trace("process")()
	}

	if len(r.separatorOffsets) > 0 {
		// after newline or at EOF

		r.w.Write([]byte("<")) // row prefix

		// add end of line separator
		r.separatorOffsets = append(r.separatorOffsets, r.offset)

		i := 0
		offset := r.lineOffset

		for _, end := range r.separatorOffsets {
			if i > 0 {
				r.w.Write([]byte("\n "))
			}

			r.w.Write([]byte("&")) // cell prefix

			if trace {
				r.printf("write; offset=%d end=%d", offset, end)
			}
			r.w.Write(r.src[offset:end])

			// +1 to skip "|"
			offset = end + 1
			i++
		}

		if r.ch == '\n' {
			r.w.Write([]byte("\n"))
		}

		r.separatorOffsets = nil

		return
	}

	if trace {
		r.printf("write; offset=%d end=%d", r.lineOffset, r.offset)
	}

	r.w.Write(r.src[r.lineOffset:r.offset])
}

func (r *recognizer) parseInlines() {
	if trace {
		defer r.trace("parseInlines")()
	}

	for r.ch > 0 && r.ch != '\n' && r.closingDelimiter() == 0 {
		r.parseInline()
	}
}

func (r *recognizer) parseInline() {
	if trace {
		defer r.trace("parseInline")()
	}

	if !r.isEscape() {
		el, ok := r.matchInline()
		if ok {
			switch el.Type {
			case node.TypeUniform:
				r.parseUniform(el.Name)
				return
			case node.TypeEscaped:
				r.parseEscaped(el.Name)
				return
			case node.TypePrefixed:
				r.parsePrefixed(el.Name, el.Delimiter, el.Matcher)
				return
			default:
				panic(fmt.Sprintf("recognizer.parseInline: unexpected node type %s (%s)", el.Type, el.Name))
			}
		}
	}

	r.parseText()
	return
}

func (r *recognizer) isEscape() bool {
	if trace {
		defer r.trace("isEscape")()
	}

	if r.ch == '\\' {
		if isPunct(r.peek()) {
			if trace {
				r.print("return true")
			}

			return true
		} else {
			for _, escape := range r.specialEscapes {
				if r.hasPrefix([]byte("\\" + escape)) {
					if trace {
						r.print("return true")
					}

					return true
				}
			}
		}
	}

	if trace {
		r.print("return false")
	}

	return false
}

// hasPrefix determines whether b matches source from offset.
func (r *recognizer) hasPrefix(b []byte) bool {
	return bytes.HasPrefix(r.src[r.offset:], b)
}

// isPunct determines whether ch is an ASCII punctuation character.
func isPunct(ch rune) bool {
	return ch >= 0x21 && ch <= 0x2F ||
		ch >= 0x3A && ch <= 0x40 ||
		ch >= 0x5B && ch <= 0x60 ||
		ch >= 0x7B && ch <= 0x7E
}

func (r *recognizer) parseUniform(name string) {
	if trace {
		defer r.tracef("parseUniform (%s)", name)()
		r.printDelims("inlines", r.inlines)
	}

	delim := r.ch

	// consume delimiter
	r.next()
	r.next()

	defer r.openInline(delim)()

	r.parseInlines()

	if r.closingDelimiter() == counterpart(delim) {
		// consume closing delimiter
		r.next()
		r.next()
	}
}

func (r *recognizer) parseEscaped(name string) {
	if trace {
		defer r.tracef("parseEscaped (%s)", name)()
	}

	delim := r.ch
	c := counterpart(delim)
	closing := []rune{c, c}

	// consume delimiter
	r.next()
	r.next()

	escaped := r.ch == '\\'
	if escaped {
		closing = append([]rune{'\\'}, closing...)
		r.next()
	}

	for r.ch > 0 && r.ch != '\n' {
		var a []rune
		if escaped {
			a = append([]rune{r.ch}, []rune{r.peek(), r.peek2()}...)
		} else {
			a = []rune{r.ch, r.peek()}
		}

		if cmpRunes(a, closing) {
			// closing delimiter

			for i := 0; i < len(closing); i++ {
				r.next()

			}

			break
		}

		r.next()
	}
}

// cmpRunes determines whether a and b have the same values.
func cmpRunes(a, b []rune) bool {
	if len(a) != len(b) {
		return false
	}

	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func counterpart(ch rune) rune {
	c, ok := leftRightChars[ch]
	if ok {
		return c
	}
	return ch
}

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

func (r *recognizer) parsePrefixed(name, prefix string, matcher string) {
	if trace {
		defer r.tracef("parsePrefixed (%s, prefix=%q, matcher=%q)", name, prefix, matcher)()
	}

	// consume prefix
	for i := 0; i < len(prefix); i++ {
		r.next()
	}

	if matcher == "" {
		return
	}

	m, ok := r.matcherMap[matcher]
	if !ok {
		panic("recognizer: matcher " + matcher + " not found")
	}

	w := m.Match(r.src[r.offset:])
	end := r.offset + w

	// consume match
	for r.offset < end {
		r.next()
	}
}

func (r *recognizer) parseText() {
	if trace {
		defer r.trace("parseText")()
	}

	for r.ch > 0 && r.ch != '\n' {
		if r.isEscape() {
			r.next()
		} else {
			if r.closingDelimiter() > 0 {
				break
			}

			if _, ok := r.matchInline(); ok {
				break
			}

			if r.ch == '|' {
				if trace {
					r.printf("mark separator at %d", r.offset)
				}

				r.separatorOffsets = append(r.separatorOffsets, r.offset)
			}
		}

		r.next()
	}
}

func (r *recognizer) matchInline() (config.Element, bool) {
	if trace {
		defer r.trace("matchInline")()
	}

	for d, e := range r.inlineMap {
		runes := utf8.RuneCountInString(d)

		if runes > 0 && e.Type == node.TypePrefixed {
			if r.hasPrefix([]byte(e.Delimiter)) {
				if trace {
					r.printf("return true (%s)", e.Name)
				}

				return e, true
			}
		} else if runes == 1 {
			if d == string(r.ch) && r.ch == r.peek() {
				if trace {
					r.printf("return true (%s)", e.Name)
				}

				return e, true
			}
		} else {
			panic("recognizer: unvalid inline delimiter " + d)
		}
	}

	if trace {
		r.print("return false")
	}

	return config.Element{}, false
}

// closingDelimiter returns the closing delimiter if found, otherwise 0.
func (r *recognizer) closingDelimiter() rune {
	if trace {
		defer r.trace("closingDelimiter")()
	}

	for i := len(r.inlines) - 1; i >= 0; i-- {
		delim := r.inlines[i]
		c := counterpart(delim)

		if r.ch == c && r.peek() == c {
			if trace {
				r.printf("return %q (is closing delim)", c)
			}

			return c
		}
	}

	if trace {
		r.print("return 0 (not closing delim)")
	}

	return 0
}

func (r *recognizer) Matchers(m matcher.Map) {
	if r.matcherMap == nil {
		r.matcherMap = make(matcher.Map)
	}

	for k, v := range m {
		r.matcherMap[k] = v
	}
}

func (r *recognizer) init(src []byte) {
	r.src = src

	r.next()
	if r.ch == '\uFEFF' {
		// skip BOM at file beginning
		r.next()
	}
}

func isSpacing(r rune) bool {
	return r == ' ' || r == '\t'
}

// parseSpacing is like parses spacing.
func (r *recognizer) parseSpacing() {
	if trace {
		defer r.trace("parseSpacing")()
	}

	var spacing []rune
	for isSpacing(r.ch) {
		spacing = append(spacing, r.ch)

		r.next()
	}

	if trace {
		r.printDelims("spacing", spacing)
	}
}

// next reads the next character.
func (r *recognizer) next() {
	if trace {
		defer r.trace("next")()
	}

	if r.rdOffset < len(r.src) {
		ch, w := utf8.DecodeRune(r.src[r.rdOffset:])

		r.ch = validRune(ch, w)
		r.offset = r.rdOffset
		r.rdOffset += w
	} else {
		r.ch = 0
		r.offset = len(r.src)
	}

	if trace {
		r.printf("r.ch=%q ", r.ch)
	}
}

func (r *recognizer) peek() rune {
	if r.rdOffset < len(r.src) {
		return validRune(utf8.DecodeRune(r.src[r.rdOffset:]))
	}
	return 0
}

func (r *recognizer) peek2() rune {
	if r.peek() > 0 {
		l := utf8.RuneLen(r.peek())

		if r.rdOffset+l < len(r.src) {
			return validRune(utf8.DecodeRune(r.src[r.rdOffset+l:]))
		}
	}

	return 0
}

func validRune(r rune, w int) rune {
	if r == utf8.RuneError && w == 0 {
		panic("recognizer: cannot decode empty slice")
	}
	return r
}

func (r *recognizer) openInline(delim rune) func() {
	size := len(r.inlines)
	r.inlines = append(r.inlines, delim)

	return func() {
		r.inlines = r.inlines[:size]
	}
}

func (r *recognizer) printDelims(name string, blocks []rune) {
	r.print(name + "=" + fmtBlocks(blocks))
}

func fmtBlocks(blocks []rune) string {
	var b strings.Builder
	b.WriteString("[")

	for i := 0; i < len(blocks); i++ {
		if i > 0 {
			b.WriteString(", ")
		}

		j := 1
		for i+j < len(blocks) && blocks[i+j-1] == blocks[i+j] {
			j++
		}

		if j > 2 {
			b.WriteString(fmt.Sprintf("%dx", j))
			i += j - 1
		}

		b.WriteString(fmt.Sprintf("%q", blocks[i]))
	}

	b.WriteString("]")
	return b.String()
}

func (r *recognizer) tracef(format string, v ...interface{}) func() {
	return r.trace(fmt.Sprintf(format, v...))
}

func (r *recognizer) trace(msg string) func() {
	r.printf("%q -> %s (", r.ch, msg)
	r.indent++

	return func() {
		r.indent--
		r.print(")")
	}
}

func (r *recognizer) printf(format string, v ...interface{}) {
	r.print(fmt.Sprintf(format, v...))
}

func (r *recognizer) print(msg string) {
	fmt.Println(strings.Repeat("\t", r.indent) + msg)
}
