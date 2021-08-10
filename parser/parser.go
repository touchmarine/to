package parser

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/touchmarine/to/config"
	"github.com/touchmarine/to/node"
	"log"
	"strings"
	"unicode/utf8"
)

const trace = false

const tabWidth = 8

var tabSpaces []rune

func init() {
	tabSpaces = make([]rune, tabWidth)
	for i := 0; i < tabWidth; i++ {
		tabSpaces[i] = ' '
	}
}

func Parse(src []byte) ([]node.Block, []error) {
	return ParseCustom(src, config.Default.Elements)
}

func ParseCustom(src []byte, elements []config.Element) ([]node.Block, []error) {
	var p parser
	p.register(elements)
	p.init(src)
	return p.parse(nil), p.errors
}

// parser holds the parsing state.
type parser struct {
	errors      []error
	src         []byte                  // source
	blockElems  []config.Element        // registered block elements
	inlineElems map[rune]config.Element // registered inline elements by delimiter

	// parsing
	ch       rune // current character
	offset   int  // character offset
	rdOffset int  // position after current character

	blocks []rune       // open blocks
	lead   []rune       // blocks on current line
	blank  bool         // whether the lead is blank
	ambis  int          // number of current ambiguous lines
	stage  []node.Block // ambiguous blocks

	inlines []rune // open inlines

	// tracing
	indent int // trace indentation
}

func (p *parser) register(elems []config.Element) {
	if p.inlineElems == nil {
		p.inlineElems = make(map[rune]config.Element)
	}

	for _, el := range elems {
		switch categ := node.TypeCategory(el.Type); categ {
		case node.CategoryBlock:
			p.blockElems = append(p.blockElems, el)

			//if els, ok := p.blockElems[el.Delimiter]; ok {
			//	if len(els) == 1 {
			//		e := els[0])

			//		if (el.Type == node.TypeHanging || el.Type == node.TypeRankedHanging) &&
			//		(e.Type == node.TypeHanging || e.Type == node.TypeRankedHanging) &&
			//		el.Type != e.Type {
			//			// hanging and ranked can use the same delimiter
			//			p.blockElems = append(p.blockElems, el)
			//			continue
			//		}
			//	}

			//	log.Fatalf(
			//		"parser: block delimiter %q is already registered",
			//		el.Delimiter,
			//	)
			//} else {
			//	p.blockElems[el.Delimiter] = []config.Element{el}
			//}

		case node.CategoryInline:
			delim, _ := utf8.DecodeRuneInString(el.Delimiter)
			if delim == utf8.RuneError {
				log.Fatal("parser: invalid UTF-8 encoding in delimiter")
			}

			if _, ok := p.inlineElems[delim]; ok {
				log.Fatalf(
					"parser: inline delimiter %q is already registered",
					delim,
				)
			}

			p.inlineElems[delim] = el

		default:
			panic(fmt.Sprintf(
				"parser: unexpected node category %s (element=%q, delimiter=%q)",
				categ.String(),
				el.Name,
				el.Delimiter,
			))
		}
	}
}

func (p *parser) parse(reqdBlocks []rune) []node.Block {
	if trace {
		defer p.trace("parse")()
	}

	var blocks []node.Block
Loop:
	for p.ch > 0 {
		if p.isSpacing() {
			p.parseSpacing()
			continue
		}

		//if p.ch == '\n' {
		//	if p.continues(reqdBlocks) == stop {
		//		break
		//	}

		//	p.next()
		//	p.parseLead()

		//	continue
		//}

		switch x := p.continues(reqdBlocks); x {
		case continues, maybe:
		//case continues:
		//	if p.stage != nil {
		//		if trace {
		//			p.printf("add %d ambiguous lines", p.ambis)
		//		}

		//		blocks = append(blocks, p.stage...)

		//		p.ambis = 0
		//		p.stage = nil
		//	} else if p.ambis > 0 {
		//		if trace {
		//			p.print("clear ambiguous lines")
		//		}

		//		p.ambis = 0
		//	}

		case stop:
			if len(blocks) > 0 && p.ambis > 0 && p.stage == nil {
				if trace {
					p.printf("stage %d ambiguous lines", p.ambis)
				}

				p.stage = append(p.stage, blocks[len(blocks)-p.ambis:]...)
				blocks = blocks[:len(blocks)-p.ambis]
			}
			break Loop

		//case maybe:
		//	if len(reqdBlocks) > 0 {
		//		// not at top level (we cannot carry ambiguous
		//		// lines any further up at top level)
		//		if trace {
		//			p.print("note ambiguous line")
		//		}
		//		p.ambis++
		//	}

		default:
			panic(fmt.Sprintf("parser: unexpected continues state %d", x))
		}

		if p.ch == '\n' {
			p.next()
			p.parseLead()

			continue
		}

		b := p.parseBlock()
		if b == nil {
			panic("parser: parseBlock() returned no block")
		}

		blocks = append(blocks, b)
	}

	return blocks
}

func (p *parser) parseBlock() node.Block {
	if trace {
		defer p.trace("parseBlock")()
	}

	if p.ch == '\\' {
		// escape block
		p.next()
	} else if p.ch == '%' {
		return p.parseHat()
	} else {
		el, ok := p.matchBlock()
		if ok {
			switch el.Type {
			case node.TypeLine:
				return p.parseLine(el.Name)
			case node.TypeVerbatimLine:
				return p.parseVerbatimLine(el.Name, el.Delimiter)
			case node.TypeWalled:
				return p.parseWalled(el.Name)
			case node.TypeHanging:
				return p.parseHanging(el.Name, el.Delimiter)
			case node.TypeRankedHanging:
				return p.parseRankedHanging(el.Name, el.Delimiter)
			case node.TypeFenced:
				if peek := p.peek(); peek > 0 && peek != utf8.RuneError && p.ch == peek {
					return p.parseFenced(el.Name)
				}
			default:
				panic(fmt.Sprintf("parser.parseBlock: unexpected node type %s (%s)", el.Type, el.Name))
			}
		}
	}

	return p.parseTextBlock()
}

func (p *parser) matchBlock() (config.Element, bool) {
	if trace {
		defer p.trace("matchBlock")()
	}

	var block config.Element
	var found bool

	for _, el := range p.blockElems {
		if p.hasPrefix([]byte(el.Delimiter)) {
			block = el
			found = true

			if peek := p.peek(); el.Type == node.TypeHanging && p.ch == peek ||
				el.Type == node.TypeRankedHanging && p.ch != peek {
				// ambigous hanging and ranked-try searching for
				// the other pair otherwise use this one
				continue
			}

			break
		}
	}

	if trace {
		p.printf("return %t (%s)", found, block.Name)
	}

	return block, found
}

// hasPrefix determines whether b matches source from offset.
func (p *parser) hasPrefix(b []byte) bool {
	return bytes.HasPrefix(p.src[p.offset:], b)
}

func (p *parser) parseHat() node.Block {
	if trace {
		defer p.trace("parseHat")()
	}

	lines := p.parseHatLines()

	var nod node.Block
	if p.ch > 0 {
		switch x := p.continues(p.blocks); x {
		case continues, maybe:
			nod = p.parseBlock()
		case stop:
		default:
			panic(fmt.Sprintf("parser: unexpected continues state %d", x))
		}
	}

	return &node.Hat{lines, nod}
}

func (p *parser) parseHatLines() [][]byte {
	if trace {
		defer p.trace("parseHatLines")()
	}

	p.addLead(p.ch)
	defer p.open(p.ch)()

	p.next() // consume delimiter

	reqdBlocks := p.blocks

	var lines [][]byte

	var b strings.Builder
	for {
		if p.ch == 0 || p.ch == '\n' {
			lines = append(lines, []byte(b.String()))
			b.Reset()

			p.next()
			p.parseLead()

			if p.ch == 0 {
				break
			}
		}

		if p.continues(reqdBlocks) != continues {
			break
		}

		for p.ch > 0 && p.ch != '\n' {
			b.WriteRune(p.ch)
			p.next()
		}
	}

	return lines
}

func (p *parser) parseWalled(name string) node.Block {
	if trace {
		defer p.tracef("parseWalled (%s)", name)()
	}

	p.addLead(p.ch)
	defer p.open(p.ch)()

	p.next() // consume delimiter

	reqdBlocks := p.blocks
	children := p.parse(reqdBlocks)

	return &node.Walled{name, children}
}

func (p *parser) parseHanging(name, delim string) node.Block {
	if trace {
		defer p.tracef("parseHanging (%s, delim=%q)", name, delim)()
	}

	c := utf8.RuneCountInString(delim)
	p.addLead([]rune(strings.Repeat(" ", c))...)

	// consume delimiter
	for i := 0; i < c; i++ {
		p.next()
	}

	children := p.parseHanging0()
	return &node.Hanging{name, children}
}

func (p *parser) parseRankedHanging(name, delim string) node.Block {
	if trace {
		defer p.tracef("parseRankedHanging (%s, delim=%q)", name, delim)()
	}

	var rank int

	// consume delimiter, count rank
	d := p.ch
	for p.ch == d {
		rank++

		p.next()
	}

	p.addLead([]rune(strings.Repeat(" ", rank))...)

	children := p.parseHanging0()
	return &node.RankedHanging{name, rank, children}
}

func (p *parser) parseHanging0() []node.Block {
	if trace {
		defer p.trace("parseHanging0")()
	}

	newBlocks := diff(p.blocks, p.lead)
	if trace {
		p.printBlocks("reqd", p.blocks)
		p.printBlocks("lead", p.lead)
		p.printBlocks("diff", newBlocks)
	}
	defer p.open(newBlocks...)()

	reqdBlocks := p.blocks
	return p.parse(reqdBlocks)
}

type continuesState int

const (
	continues continuesState = iota
	stop
	maybe
)

func (p *parser) continues(blocks []rune) continuesState {
	if trace {
		defer p.trace("continues")()
		p.printBlocks("reqd", blocks)
		p.printBlocks("lead", p.lead)
	}

	if p.blank && onlySpacing(blocks) {
		if trace {
			p.print("return maybe (blank)")
		}
		return maybe
	}

	var i, j int
	for {
		if i > len(blocks)-1 {
			if trace {
				p.print("return true")
			}
			return continues
		}

		if j > len(p.lead)-1 {
			if onlySpacing(blocks[i:]) && len(p.lead) > 0 &&
				(p.ch == 0 || p.ch == '\n' || p.ch == ' ' || p.ch == '\t') {
				if trace {
					p.print("return maybe")
				}
				return maybe
			}

			if trace {
				p.print("return false (not enough blocks)")
			}
			return stop
		}

		if blocks[i] == ' ' || blocks[i] == '\t' || p.lead[j] == ' ' || p.lead[j] == '\t' {
			n, m := i, j
			for i < len(blocks) {
				if blocks[i] == ' ' || blocks[i] == '\t' {
					i++
				} else {
					break
				}
			}

			for j < len(p.lead) {
				if p.lead[j] == ' ' || p.lead[j] == '\t' {
					j++
				} else {
					break
				}
			}

			x := countSpacing(spacingSeq(blocks[n:i]))
			y := countSpacing(spacingSeq(p.lead[m:j]))

			if y < x {
				if trace {
					p.print("return false (lesser ident)")
				}
				return stop
			}

			continue
		}

		if blocks[i] != p.lead[j] {
			if trace {
				p.printf("return false (%q != %q, i=%d, j=%d)", blocks[i], p.lead[j], i, j)
			}
			return stop
		}

		i++
		j++
	}
}

func onlySpacing(a []rune) bool {
	return len(spacingSeq(a)) == len(a)
}

func spacingSeq(a []rune) []rune {
	for i, v := range a {
		if v == ' ' || v == '\t' {
			continue
		}
		return a[:i]
	}
	return a
}

func (p *parser) parseFenced(name string) node.Block {
	if trace {
		defer p.tracef("parseFenced (%s)", name)()
	}

	openSpacing := p.spacing()
	reqdBlocks := p.blocks
	delim := p.ch

	var i int
	for p.ch == delim {
		i++

		p.next()
	}

	var lines [][]byte
	var trailingText []byte

	var b strings.Builder
OuterLoop:
	for {
		for p.ch == 0 || p.ch == '\n' {
			lines = append(lines, []byte(b.String()))
			b.Reset()

			if p.ch == 0 {
				break OuterLoop
			}

			p.next()
			p.parseLead()

			newSpacing := diffSpacing(openSpacing, p.spacing())
			for _, ch := range newSpacing {
				b.WriteRune(ch)
			}
		}

		if p.continues(reqdBlocks) != continues {
			break
		}

		var j int
		for p.ch > 0 && p.ch != '\n' {
			if p.ch == delim {
				j++
			} else {
				j = 0
			}

			if j == i && len(lines) > 0 {
				// closing delimiter
				b.Reset()

				p.next() // consume last closing character

				// save trailing text
				for p.ch > 0 && p.ch != '\n' {
					b.WriteRune(p.ch)

					p.next()
				}

				trailingText = []byte(b.String())

				p.next()
				p.parseLead()

				break OuterLoop
			}

			b.WriteRune(p.ch)

			p.next()
		}
	}

	return &node.Fenced{name, lines, trailingText}
}

// a=old spacing
// b=new spacing
func diffSpacing(a, b []rune) []rune {
	x := countSpacing(a)
	y := countSpacing(b)

	if y == x {
		return nil
	} else if y > x {
		var c []rune

		n := y - x
		for i := len(b) - 1; i >= 0; i-- {
			if n <= 0 {
				break
			}

			w := countSpacing([]rune{b[i]})
			if w > n {
				for j := 0; j < n; j++ {
					c = append(c, ' ')
				}
				break
			}
			c = append(c, b[i])
			n -= w
		}

		return c
	}

	return nil
}

func countSpacing(s []rune) int {
	var i int
	for _, ch := range s {
		switch ch {
		case ' ':
			i++
		case '\t':
			i += tabWidth
		default:
			panic(fmt.Sprintf("countSpacing: got %q, want ' ' or '\t'", ch))
		}
	}
	return i
}

func (p *parser) spacing() []rune {
	if trace {
		defer p.trace("spacing")()
	}

	a := p.lead
	for i := len(a) - 1; i >= 0; i-- {
		if a[i] != ' ' && a[i] != '\t' {
			a = a[i+1:]
			break
		}
	}

	if trace {
		p.printBlocks("spacing", a)
	}

	return a
}

func (p *parser) parseVerbatimLine(name, delim string) node.Block {
	if trace {
		defer p.tracef("parseVerbatimLine (%s, delim=%q)", name, delim)()
	}

	for i := 0; i < utf8.RuneCountInString(delim); i++ {
		p.next()

	}

	var b bytes.Buffer
	for p.ch > 0 && p.ch != '\n' {
		b.WriteRune(p.ch)

		p.next()
	}

	p.next()
	p.parseLead()

	return &node.VerbatimLine{name, b.Bytes()}
}

func (p *parser) parseTextBlock() node.Block {
	if trace {
		defer p.trace("parseTextBlock")()
	}

	children := p.parseInlines(false)

	return &node.Line{"TextBlock", children}
}

func (p *parser) parseLine(name string) node.Block {
	if trace {
		defer p.tracef("parseLine (%s)", name)()
	}

	children := p.parseInlines(false)

	if p.ch == 0 || p.ch == '\n' {
		p.next()
		p.parseLead()
	}

	return &node.Line{name, children}
}

func (p *parser) parseInlines(afterNewline bool) []node.Inline {
	if trace {
		defer p.tracef("parseInlines (afterNewline=%t)", afterNewline)()
	}

	var inlines []node.Inline
	for p.ch > 0 {
		if p.closingDelimiter() > 0 {
			break
		}

		inline, cont := p.parseInline(afterNewline)
		if inline == nil {
			panic("parser.parseInlines: nil inline")
		}

		inlines = append(inlines, inline)

		if !cont {
			break
		}
	}
	return inlines
}

func (p *parser) parseInline(afterNewline bool) (node.Inline, bool) {
	if trace {
		defer p.trace("parseInline")()
	}

	el, ok := p.openingDelimiter()
	if ok {
		switch el.Type {
		case node.TypeUniform:
			return p.parseUniform(el.Name)
		case node.TypeEscaped:
			return p.parseEscaped(el.Name), true
		default:
			panic(fmt.Sprintf("parser.parseInline: unexpected node type %s (%s)", el.Type, el.Name))
		}
	}

	return p.parseText(afterNewline)
}

func (p *parser) isInlineEscape() bool {
	if trace {
		defer p.trace("isInlineEscape")()
	}

	if p.ch != '\\' {
		if trace {
			p.print("return false")
		}
		return false
	}

	peek := p.peek()
	if peek == 0 || peek == utf8.RuneError {
		if trace {
			p.print("return false")
		}
		return false
	}

	if peek == '\\' || peek == '/' {
		if trace {
			p.print("return true")
		}
		return true
	}

	_, ok := p.inlineElems[peek]
	if trace {
		p.printf("return %t", ok)
	}
	return ok
}

func (p *parser) parseUniform(name string) (node.Inline, bool) {
	if trace {
		defer p.tracef("parseUniform (%s)", name)()
		p.printInlines("inlines", p.inlines)
	}

	delim := p.ch

	// consume delimiter
	p.next()
	p.next()

	afterNewline := false
	if p.ch == '\n' {
		p.next()
		p.parseLead()

		afterNewline = true

		_, ok := p.matchBlock()
		if p.ch == '%' || ok || p.continues(p.blocks) != continues {
			return &node.Uniform{name, nil}, false
		}
	}

	defer p.openInline(delim)()

	children := p.parseInlines(afterNewline)

	if p.closingDelimiter() == counterpart(delim) {
		// consume closing delimiter
		p.next()
		p.next()
	}

	return &node.Uniform{name, children}, true
}

func (p *parser) parseEscaped(name string) node.Inline {
	if trace {
		defer p.tracef("parseEscaped (%s)", name)()
	}

	delim := p.ch
	c := counterpart(delim)
	closing := []rune{c, c}

	// consume delimiter
	p.next()
	p.next()

	escaped := p.ch == '\\'
	if escaped {
		closing = append([]rune{'\\'}, closing...)
		p.next()
	}

	var content []byte

	if p.ch > 0 && p.ch != '\n' {
		var b bytes.Buffer
		for p.ch > 0 && p.ch != '\n' {
			var a []rune
			if escaped {
				a = append([]rune{p.ch}, []rune{p.peek(), p.peek2()}...)
			} else {
				a = []rune{p.ch, p.peek()}
			}

			if cmpRunes(a, closing) {
				// consume closing delimiter
				for i := 0; i < len(closing); i++ {
					p.next()

				}
				break
			}

			b.WriteRune(p.ch)

			p.next()
		}

		content = b.Bytes()
	}

	if trace {
		defer p.printf("return %q", content)
	}

	return &node.Escaped{name, content}
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

func (p *parser) parseText(afterNewline bool) (node.Inline, bool) {
	if trace {
		defer p.trace("parseText")()
	}

	cont := true

	var b bytes.Buffer
	for p.ch > 0 {
		if p.ch == '\n' {
			if afterNewline {
				cont = false
				break
			}

			p.next()
			p.parseLead()

			afterNewline = true

			_, ok := p.matchBlock()
			if p.ch == '%' || ok || p.continues(p.blocks) != continues {
				cont = false
				break
			}

			continue
		}

		if p.isInlineEscape() {
			p.next()
		} else {
			if p.closingDelimiter() > 0 {
				break
			}

			if _, ok := p.openingDelimiter(); ok {
				break
			}
		}

		if afterNewline {
			line := b.Bytes()
			trailingSpacing := len(line) - len(bytes.TrimRight(line, " \t"))

			if trailingSpacing > 0 {
				// line has trailing spacing
				b.Truncate(trailingSpacing)
			}

			b.WriteByte(' ') // newline separator

			afterNewline = false
		}

		b.WriteRune(p.ch)

		p.next()
	}

	txt := b.Bytes()

	if trace {
		defer p.printf("return %q", txt)
	}

	return node.Text(txt), cont
}

// openingDelimiter returns the element of inline delimiter if found.
func (p *parser) openingDelimiter() (config.Element, bool) {
	if trace {
		defer p.trace("openingDelimiter")()
	}

	el, ok := p.inlineElems[p.ch]
	if ok && p.ch == p.peek() {
		switch el.Type {
		case node.TypeUniform, node.TypeEscaped:
			if trace {
				p.printf("return true (%s)", el.Name)
			}

			return el, true
		default:
			panic(fmt.Sprintf("parser.parseText: unexpected node type %s (%s)", el.Type, el.Name))
		}
	}

	if trace {
		p.print("return false")
	}

	return config.Element{}, false
}

// closingDelimiter returns the closing delimiter if found, otherwise 0.
func (p *parser) closingDelimiter() rune {
	if trace {
		defer p.trace("closingDelimiter")()
	}

	for i := len(p.inlines) - 1; i >= 0; i-- {
		delim := p.inlines[i]
		c := counterpart(delim)

		if p.ch == c && p.peek() == c {
			if trace {
				p.printf("return %q (is closing delim)", c)
			}

			return c
		}
	}

	if trace {
		p.print("return 0 (not closing delim)")
	}

	return 0
}

func (p *parser) init(src []byte) {
	p.src = src

	p.next()
	if p.ch == '\uFEFF' {
		// skip BOM at file beginning
		p.next()
	}
}

// parseLead parses spacing and block delimiters at the start of the line.
//
// Call at the start of the line as it consumes required block delimiters.
func (p *parser) parseLead() {
	if trace {
		defer p.trace("parseLead")()
		p.printBlocks("reqd", p.blocks)
	}

	p.lead = nil
	p.blank = true

	a := stripSpacing(p.blocks)

	var lead []rune
	var i int
	for p.ch > 0 {
		if p.ch != '\n' && p.ch != ' ' && p.ch != '\t' {
			p.blank = false
		}

		if i < len(a) && p.ch == a[i] {
			i++
		} else if p.isSpacing() {
		} else {
			break
		}

		lead = append(lead, p.ch)

		p.next()
	}

	p.addLead(lead...)

	if trace {
		p.printBlocks("new", lead)
		p.printBlocks("lead", p.lead)
	}
}

func stripSpacing(a []rune) []rune {
	var b []rune
	for _, c := range a {
		if c == ' ' || c == '\t' {
			continue
		}
		b = append(b, c)
	}
	return b
}

// parseSpacing is like parseLead but only parses spacing and can be used in the
// middle of the line.
func (p *parser) parseSpacing() {
	if trace {
		defer p.trace("parseSpacing")()
	}

	var lead []rune
	for p.isSpacing() {
		lead = append(lead, p.ch)

		p.next()
	}

	p.addLead(lead...)

	if trace {
		p.printBlocks("new", lead)
		p.printBlocks("lead", p.lead)
	}
}

func (p *parser) isSpacing() bool {
	return p.ch == ' ' || p.ch == '\t'
}

// Encoding errors
var (
	ErrInvalidUTF8Encoding = errors.New("invalid UTF-8 encoding")
	ErrIllegalNULL         = errors.New("illegal character NULL")
	ErrIllegalBOM          = errors.New("illegal byte order mark")
)

// next reads the next character.
func (p *parser) next() {
	if trace {
		defer p.trace("next")()
	}

	if p.rdOffset < len(p.src) {
		p.offset = p.rdOffset

		r, w := utf8.DecodeRune(p.src[p.rdOffset:])

		switch r {
		case utf8.RuneError: // encoding error
			if w == 0 {
				// EOF
				p.ch = 0
			} else if w == 1 {
				p.error(ErrInvalidUTF8Encoding)
				p.ch = utf8.RuneError
			}
		case '\u0000': // NULL
			p.error(ErrIllegalNULL)
			p.ch = utf8.RuneError
		case '\uFEFF': // BOM
			if p.offset == 0 {
				// skip in p.init
				p.ch = r
			} else {
				p.error(ErrIllegalBOM)
				p.ch = utf8.RuneError
			}
		default:
			p.ch = r
		}

		p.rdOffset += w
	} else {
		p.ch = 0
		p.offset = len(p.src)
	}

	if trace {
		p.printf("p.ch=%q ", p.ch)
	}
}

func (p *parser) peek() rune {
	if p.rdOffset < len(p.src) {
		r, w := utf8.DecodeRune(p.src[p.rdOffset:])

		switch r {
		case utf8.RuneError:
			if w == 0 {
				// EOF
				return 0
			} else if w == 1 {
				return utf8.RuneError
			}
		case '\u0000', '\uFEFF': // encoding error, NULL, or BOM
			return utf8.RuneError
		default:
			return r
		}
	}

	return 0
}

func (p *parser) peek2() rune {
	if p.rdOffset < len(p.src) {
		_, w := utf8.DecodeRune(p.src[p.rdOffset:])
		if w == 0 {
			return 0
		}

		r, _ := utf8.DecodeRune(p.src[p.rdOffset+w:])

		switch r {
		case utf8.RuneError:
			if w == 0 {
				// EOF
				return 0
			} else if w == 1 {
				return utf8.RuneError
			}
		case '\u0000', '\uFEFF': // encoding error, NULL, or BOM
			return utf8.RuneError
		default:
			return r
		}
	}

	return 0
}

func (p *parser) open(blocks ...rune) func() {
	size := len(p.blocks)
	p.blocks = append(p.blocks, blocks...)

	return func() {
		p.blocks = p.blocks[:size]
	}
}

func (p *parser) addLead(blocks ...rune) {
	p.lead = append(p.lead, blocks...)
}

func diff(old, new []rune) []rune {
	if len(old) == 0 {
		return new
	}

	n := expandTabs(new)

	a := trailingSpacing(old)
	if len(a) < len(old) {
		b := trailingSpacing(n)
		if len(a) > 0 && len(b) > 0 {
			x := countSpacing(a)
			y := countSpacing(b)
			if y > x {
				// different trailing spacing
				z := y - x
				return n[len(n)-z:]
			}
		}
	}

	var i int
	for i = len(n) - 1; i >= 0; i-- {
		if i < len(old) {
			break
		}

		c := n[i]
		if c != ' ' && c != '\t' && c == old[len(old)-1] {
			break
		}
	}

	return n[i+1:]
}

func trailingSpacing(a []rune) []rune {
	var i int
	for i = len(a) - 1; i >= 0; i-- {
		c := a[i]
		if c != ' ' && c != '\t' {
			break
		}
	}
	return a[i+1:]
}

func expandTabs(a []rune) []rune {
	n := make([]rune, 0, len(a))
	for _, c := range a {
		if c == '\t' {
			n = append(n, tabSpaces...)
		} else {
			n = append(n, c)
		}
	}
	return n
}

func (p *parser) openInline(delim rune) func() {
	size := len(p.inlines)
	p.inlines = append(p.inlines, delim)

	return func() {
		p.inlines = p.inlines[:size]
	}
}

func (p *parser) error(err error) {
	p.errors = append(p.errors, err)
}

func (p *parser) printBlocks(name string, blocks []rune) {
	p.print(name + "=" + fmtBlocks(blocks))
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

func (p *parser) printInlines(name string, inlines []rune) {
	p.printf("%s=%q", name, inlines)
}

func (p *parser) tracef(format string, v ...interface{}) func() {
	return p.trace(fmt.Sprintf(format, v...))
}

func (p *parser) trace(msg string) func() {
	p.printf("%q -> %s (", p.ch, msg)
	p.indent++

	return func() {
		p.indent--
		p.print(")")
	}
}

func (p *parser) printf(format string, v ...interface{}) {
	p.print(fmt.Sprintf(format, v...))
}

func (p *parser) print(msg string) {
	fmt.Println(strings.Repeat("\t", p.indent) + msg)
}
