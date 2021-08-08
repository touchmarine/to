package parser

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/touchmarine/to/config"
	"github.com/touchmarine/to/node"
	"io"
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

func Parse(r io.Reader) ([]node.Block, []error) {
	return ParseCustom(r, config.Default.Elements)
}

func ParseCustom(r io.Reader, elements []config.Element) ([]node.Block, []error) {
	var p parser
	p.register(elements)
	p.init(r)
	return p.parse(nil), p.errors
}

// parser holds the parsing state.
type parser struct {
	errors      []error
	scnr        *bufio.Scanner          // line scanner
	blockElems  []config.Element        // map of block elements by delimiter
	inlineElems map[rune]config.Element // map of inline elements by delimiter

	// parsing
	ln    []byte // current line excluding EOL
	ch    rune   // current character
	atEOF bool   // at end of file

	blocks []rune       // open blocks
	lead   []rune       // blocks on current line
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
			//		"block delimiter %q is already registered",
			//		el.Delimiter,
			//	)
			//} else {
			//	p.blockElems[el.Delimiter] = []config.Element{el}
			//}

		case node.CategoryInline:
			delim, _ := utf8.DecodeRuneInString(el.Delimiter)
			if delim == utf8.RuneError {
				log.Fatal("invalid UTF-8 encoding in delimiter")
			}

			if _, ok := p.inlineElems[delim]; ok {
				log.Fatalf(
					"inline delimiter %q is already registered",
					delim,
				)
			}

			p.inlineElems[delim] = el

		default:
			panic(fmt.Sprintf(
				"unexpected node category %s (element=%q, delimiter=%q)",
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
OuterLoop:
	for !p.atEOF {
		if p.ch == ' ' || p.ch == '\t' {
			p.parseSpacing()
		}

		if p.atEOL() {
			p.nextlnf()
			p.nextch()
			p.parseLead()
			continue
		}

		switch x := p.continues0(reqdBlocks); x {
		case 0: // continue
			if p.stage != nil {
				if trace {
					p.printf("add %d ambiguous lines", p.ambis)
				}

				blocks = append(blocks, p.stage...)

				p.ambis = 0
				p.stage = nil
			} else if p.ambis > 0 {
				if trace {
					p.print("clear ambiguous lines")
				}

				p.ambis = 0
			}
		case 1: // stop
			if len(blocks) > 0 && p.ambis > 0 && p.stage == nil {
				if trace {
					p.printf("stage %d ambiguous lines", p.ambis)
				}

				p.stage = append(p.stage, blocks[len(blocks)-p.ambis:]...)
				blocks = blocks[:len(blocks)-p.ambis]
			}
			break OuterLoop
		case 2: // maybe
			if len(reqdBlocks) > 0 {
				// not at top level (we cannot carry ambiguous
				// lines any further up at top level)
				if trace {
					p.print("note ambiguous line")
				}
				p.ambis++
			}
		default:
			panic(fmt.Sprintf("unexpected continues state %d", x))
		}

		b := p.parseBlock()
		if b != nil {
			blocks = append(blocks, b)
		}
	}

	return blocks
}

func (p *parser) parseBlock() node.Block {
	if trace {
		defer p.trace("parseBlock")()
	}

	if p.ch == '\\' {
		// escape block
		p.nextch()
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
				if peek := p.peek(); isValidRune(peek) && p.ch == peek {
					return p.parseFenced(el.Name)
				}
			default:
				panic(fmt.Sprintf("parseBlock: unexpected node type %s (%s)", el.Type, el.Name))
			}
		}
	}

	if p.atEOL() {
		return nil
	}

	return p.parseLine("Line")
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

			if el.Type == node.TypeHanging && p.consecutive() > 1 ||
				el.Type == node.TypeRankedHanging && p.consecutive() == 1 {
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

// hasPrefix determines whether b matches the p.ch and start of p.ln.
func (p *parser) hasPrefix(b []byte) bool {
	ln := append([]byte{byte(p.ch)}, p.ln...)
	return bytes.HasPrefix(ln, b)
}

// consecutive determines the number of consecutive characters.
func (p *parser) consecutive() int {
	if trace {
		defer p.trace("consecutive")()
	}

	ch := p.ch

	i := 1

	var offs int
	for {
		peek, w := utf8.DecodeRune(p.ln[offs:])
		if peek == utf8.RuneError {
			break
		}

		if peek != ch {
			break
		}

		i++
		offs += w
	}

	if trace {
		p.printf("return %d", i)
	}

	return i
}

func (p *parser) parseHat() node.Block {
	if trace {
		defer p.trace("parseHat")()
	}

	lines := p.parseHatLines()

	var nod node.Block
	if !p.atEOF {
		switch x := p.continues0(p.blocks); x {
		case 0, 2: // continue, maybe
			nod = p.parseBlock()
		case 1: // stop
		default:
			panic(fmt.Sprintf("unexpected continues state %d", x))
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

	p.nextch() // consume delimiter

	reqdBlocks := p.blocks

	var lines [][]byte

	var b strings.Builder
	for {
		if p.atEOL() {
			lines = append(lines, []byte(b.String()))
			b.Reset()

			p.nextln()
			p.nextch()
			p.parseLead()

			if p.atEOF {
				break
			}
		}

		if !p.continues(reqdBlocks) {
			break
		}

		for !p.atEOL() {
			b.WriteRune(p.ch)
			p.nextch()
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

	p.nextch() // consume delimiter

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
	p.nextchn(c)

	children := p.parseHanging0()
	return &node.Hanging{name, children}
}

func (p *parser) parseRankedHanging(name, delim string) node.Block {
	if trace {
		defer p.tracef("parseRankedHanging (%s, delim=%q)", name, delim)()
	}

	var rank int

	d := p.ch
	for p.ch == d {
		rank++

		p.nextch()
	}

	p.addLead([]rune(strings.Repeat(" ", rank))...)

	children := p.parseHanging0()
	return &node.RankedHanging{name, rank, children}
}

func (p *parser) parseHanging0() []node.Block {
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

func (p *parser) continues(blocks []rune) bool {
	return p.continues0(blocks) == 0
}

func (p *parser) continues0(blocks []rune) int {
	if trace {
		defer p.trace("continues0")()
		p.printBlocks("reqd", blocks)
		p.printBlocks("lead", p.lead)
	}

	var i, j int
	for {
		if i > len(blocks)-1 {
			if trace {
				p.print("return true")
			}
			return 0
		}

		if j > len(p.lead)-1 {
			if onlySpacing(blocks[i:]) && len(p.lead) > 0 &&
				(p.atEOL() || p.ch == ' ' || p.ch == '\t') {
				if trace {
					p.print("return maybe")
				}
				return 2
			}

			if trace {
				p.print("return false (not enough blocks)")
			}
			return 1
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
				return 1
			}

			continue
		}

		if blocks[i] != p.lead[j] {
			if trace {
				p.printf("return false (%q != %q, i=%d, j=%d)", blocks[i], p.lead[j], i, j)
			}
			return 1
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

		if !p.nextch() {
			break
		}
	}

	var lines [][]byte
	var trailingText []byte

	var b strings.Builder
OuterLoop:
	for {
		for p.atEOL() {
			lines = append(lines, []byte(b.String()))
			b.Reset()

			if !p.nextln() {
				break OuterLoop
			}

			p.nextch()
			p.parseLead()

			newSpacing := diffSpacing(openSpacing, p.spacing())
			for _, ch := range newSpacing {
				b.WriteRune(ch)
			}
		}

		if !p.continues(reqdBlocks) {
			break
		}

		var j int
		for {
			if p.ch == delim {
				j++
			} else {
				j = 0
			}

			if j == i && len(lines) > 0 {
				// closing delimiter
				b.Reset()

				for p.nextch() {
					b.WriteRune(p.ch)
				}
				trailingText = []byte(b.String())

				p.nextln()
				p.nextch()
				p.parseLead()
				break OuterLoop
			}

			b.WriteRune(p.ch)

			if !p.nextch() {
				break
			}
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

	p.nextchn(utf8.RuneCountInString(delim))

	var b bytes.Buffer
	for !p.atEOL() {
		b.WriteRune(p.ch)

		if !p.nextch() {
			break
		}
	}

	p.nextlnf()
	p.nextch()
	p.parseLead()

	return &node.VerbatimLine{name, b.Bytes()}
}

func (p *parser) parseLine(name string) node.Block {
	if trace {
		defer p.tracef("parseLine (%s)", name)()
	}

	children := p.parseInlines()

	if p.atEOL() {
		p.nextlnf()
		p.nextch()
		p.parseLead()
	}

	return &node.Line{name, children}
}

func (p *parser) parseInlines() []node.Inline {
	if trace {
		defer p.trace("parseInlines")()
	}

	var inlines []node.Inline
	for {
		if p.atEOL() {
			break
		}

		if p.closingDelimiter() > 0 {
			break
		}

		inline := p.parseInline()
		if inline == nil {
			panic("parser.parseInlines: nil inline")
		}

		inlines = append(inlines, inline)
	}
	return inlines
}

func (p *parser) parseInline() node.Inline {
	if trace {
		defer p.trace("parseInline")()
	}

	el, ok := p.openingDelimiter()
	if ok {
		switch el.Type {
		case node.TypeUniform:
			return p.parseUniform(el.Name)
		case node.TypeEscaped:
			return p.parseEscaped(el.Name)
		default:
			panic(fmt.Sprintf("parser.parseInline: unexpected node type %s (%s)", el.Type, el.Name))
		}
	}

	return p.parseText()
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
	if peek == utf8.RuneError {
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

func (p *parser) parseUniform(name string) node.Inline {
	if trace {
		defer p.tracef("parseUniform (%s)", name)()
		p.printInlines("inlines", p.inlines)
	}

	delim := p.ch
	p.nextchn(2)

	defer p.openInline(delim)()

	children := p.parseInlines()

	if p.closingDelimiter() == counterpart(delim) {
		p.nextchn(2)
	}

	return &node.Uniform{name, children}
}

func (p *parser) parseEscaped(name string) node.Inline {
	if trace {
		defer p.tracef("parseEscaped (%s)", name)()
	}

	delim := p.ch
	c := counterpart(delim)
	closing := []rune{c, c}

	p.nextchn(2)

	escaped := p.ch == '\\'
	if escaped {
		closing = append([]rune{'\\'}, closing...)
		p.nextch()
	}

	var content []byte

	if !p.atEOL() {
		var b bytes.Buffer
		for {
			var a []rune
			if escaped {
				a = append([]rune{p.ch}, p.peekn(2)...)
			} else {
				a = []rune{p.ch, p.peek()}
			}

			if cmpRunes(a, closing) {
				// closing delimiter
				p.nextchn(len(closing))
				break
			}

			b.WriteRune(p.ch)

			if !p.nextch() {
				break
			}
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

func (p *parser) parseText() node.Inline {
	if trace {
		defer p.trace("parseText")()
	}

	var b bytes.Buffer
	for {
		if p.isInlineEscape() {
			p.nextch()
			goto next
		}

		if p.closingDelimiter() > 0 {
			break
		}

		if _, ok := p.openingDelimiter(); ok {
			break
		}

	next:
		b.WriteRune(p.ch)

		if !p.nextch() {
			break
		}
	}

	txt := b.Bytes()

	if trace {
		defer p.printf("return %q", txt)
	}

	return node.Text(txt)
}

// openingDelimiter returns the element of inline delimiter if found.
func (p *parser) openingDelimiter() (config.Element, bool) {
	el, ok := p.inlineElems[p.ch]
	if ok && p.ch == p.peek() {
		switch el.Type {
		case node.TypeUniform, node.TypeEscaped:
			return el, true
		default:
			panic(fmt.Sprintf("parser.parseText: unexpected node type %s (%s)", el.Type, el.Name))
		}
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
		p.printf("return 0 (not closing delim)")
	}

	return 0
}

func (p *parser) init(r io.Reader) {
	p.scnr = bufio.NewScanner(r)
	p.nextlnf()
	ch, w := utf8.DecodeRune(p.ln)
	if ch == '\uFEFF' {
		// skip BOM if first character
		p.ln = p.ln[w:]
	}
	p.nextch()
	p.parseLead()
}

// parseLead parses spacing and block delimiters at the start of the line.
//
// Call at the start of the line as it consumes required block delimiters.
func (p *parser) parseLead() {
	if trace {
		defer p.trace("parseLead")()
		p.printBlocks("reqd", p.blocks)
	}

	a := stripSpacing(p.blocks)

	var lead []rune
	var i int
	for {
		if i < len(a) && p.ch == a[i] {
			i++
		} else if p.ch == ' ' || p.ch == '\t' {
		} else {
			break
		}

		lead = append(lead, p.ch)

		if !p.nextch() {
			break
		}
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
	for p.ch == ' ' || p.ch == '\t' {
		lead = append(lead, p.ch)

		if !p.nextch() {
			break
		}
	}

	p.addLead(lead...)

	if trace {
		p.printBlocks("new", lead)
		p.printBlocks("lead", p.lead)
	}
}

// nextlnf sets next non-blank (filled) line.
func (p *parser) nextlnf() bool {
	if trace {
		defer p.trace("nextlnf")()
	}

	cont := p.nextln()

	if onlySpacing([]rune(string(p.ln))) {
		return p.nextln()
	}

	return cont
}

func (p *parser) nextln() bool {
	if trace {
		defer p.trace("nextln")()
	}

	cont := p.scnr.Scan()
	p.ln = p.scnr.Bytes()

	if err := p.scnr.Err(); err != nil {
		switch err {
		case bufio.ErrTooLong:
			log.Fatal("line too long")
		default:
			panic(err)
		}
	}

	if !cont {
		p.atEOF = true
	}

	p.lead = nil

	if trace {
		if cont {
			p.printf("p.ln=%q", p.ln)
		} else {
			p.print("EOF")
		}
	}

	return cont
}

// Encoding errors
var (
	ErrInvalidUTF8Encoding = errors.New("invalid UTF-8 encoding")
	ErrIllegalNULL         = errors.New("illegal character NULL")
	ErrIllegalBOM          = errors.New("illegal byte order mark")
)

// nextchn calls nextch n times. It returns false if any call returns false.
func (p *parser) nextchn(n int) bool {
	for i := 0; i < n; i++ {
		if !p.nextch() {
			return false
		}
	}
	return true
}

// nextch reads the next character.
func (p *parser) nextch() bool {
	if trace {
		defer p.trace("nextch")()
	}

	r, w := utf8.DecodeRune(p.ln)

	var ch rune
	switch r {
	case utf8.RuneError: // encoding error
		if w == 0 {
			if trace {
				p.printf("EOL")
			}

			// empty p.ln
			p.ch = 0
			return false
		} else if w == 1 {
			p.error(ErrInvalidUTF8Encoding)
			ch = utf8.RuneError
		}
	case '\u0000': // NULL
		p.error(ErrIllegalNULL)
		ch = utf8.RuneError
	case '\uFEFF': // BOM
		p.error(ErrIllegalBOM)
		ch = utf8.RuneError
	default:
		ch = r
	}

	p.ch = ch
	p.ln = p.ln[w:]

	if trace {
		p.printf("p.ch=%q ", p.ch)
	}

	return true
}

func (p *parser) atEOL() bool {
	return p.ch == 0
}

// peek returns the next character but does not advance the parser.
func (p *parser) peek() rune {
	a := p.peekn(1)
	if len(a) != 1 {
		panic("unexpected number of peeks")
	}

	return a[0]
}

// peekn returns n next characters. Slice length is always n. Empty runes are
// set to 0.
func (p *parser) peekn(n int) []rune {
	a := make([]rune, n)

	var offs int
	for i := 0; i < n; i++ {
		r, w := utf8.DecodeRune(p.ln[offs:])
		if r == utf8.RuneError && w == 0 {
			// empty
			a[i] = 0
		} else {
			a[i] = r
		}

		offs += w
	}

	return a
}

// isValidRune determines whether rune is not empty and not RuneError.
func isValidRune(r rune) bool {
	return r > 0 && r != utf8.RuneError
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
