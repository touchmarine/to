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
	scnr        *bufio.Scanner            // line scanner
	blockElems  map[string]config.Element // map of block elements by delimiter
	inlineElems map[rune]config.Element   // map of inline elements by delimiter

	// parsing
	ln    []byte // current line excluding EOL
	ch    rune   // current character
	atEOF bool   // at end of file

	blocks []rune       // open blocks
	lead   []rune       // blocks on current line
	blank  bool         // whether the lead is blank
	ambis  int          // number of current ambiguous lines
	stage  []node.Block // ambiguous blocks

	inlines [][2]rune // open inlines

	// tracing
	tindent int // trace indentation
}

func (p *parser) register(elems []config.Element) {
	if p.blockElems == nil {
		p.blockElems = make(map[string]config.Element)
	}
	if p.inlineElems == nil {
		p.inlineElems = make(map[rune]config.Element)
	}

	for _, el := range elems {
		switch categ := node.TypeCategory(el.Type); categ {
		case node.CategoryBlock:
			if _, ok := p.blockElems[el.Delimiter]; ok {
				log.Fatalf(
					"parser: block delimiter %q is already registered",
					el.Delimiter,
				)
			}

			p.blockElems[el.Delimiter] = el

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

// The argument l is whether only line children are allowed.
func (p *parser) parse(reqdBlocks []rune) []node.Block {
	if trace {
		defer p.trace("parse")()
	}

	var blocks []node.Block
L:
	for !p.atEOF {
		if p.ch == ' ' || p.ch == '\t' {
			p.parseSpacing()
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
			break L
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
			panic(fmt.Sprintf("parser: unexpected continues state %d", x))
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
		p.nextch()
	} else if p.ch == '%' {
		return p.parseHat()
	} else {
		el, ok := p.matchBlock()
		if ok {
			switch el.Type {
			case node.TypeLine:
				return p.parseLine(el.Name)
			case node.TypeWalled:
				return p.parseWalled(el.Name)
			case node.TypeHanging:
				if el.MinRank <= 1 || p.consecutive() >= el.MinRank {
					return p.parseHanging(
						el.Name,
						el.Delimiter,
						el.Ranked,
						el.Verbatim,
					)
				}
			case node.TypeFenced:
				if p.peekEquals(p.ch) {
					return p.parseFenced(el.Name)
				}
			default:
				panic(fmt.Sprintf("parser.parseBlock: unexpected node type %s (%s)", el.Type, el.Name))
			}
		}
	}

	return p.parseLine("Line")
}

func (p *parser) matchBlock() (config.Element, bool) {
	if trace {
		defer p.trace("matchBlock")()
	}

OuterLoop:
	for name, el := range p.blockElems {
		var offs int
		for i, r := range name {
			if i == 0 {
				if p.ch != r {
					// skip
					continue OuterLoop
				}
			} else {
				if offs > len(p.ln)-1 {
					// skip
					continue OuterLoop
				}

				rr, w := utf8.DecodeRune(p.ln[offs:])
				if rr != r {
					// skip
					continue OuterLoop
				}
				offs += w
			}
		}

		if trace {
			p.printf("return true (%s)", el.Name)
		}
		return el, true
	}

	if trace {
		p.print("return false")
	}
	return config.Element{}, false
}

func (p *parser) consecutive() uint {
	if trace {
		defer p.trace("consecutive")()
	}

	ch := p.ch

	var i uint = 1
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

func (p *parser) isLineComment() bool {
	if trace {
		defer p.trace("isLineComment")()
	}

	t := p.ch == '/' && p.peekEquals('/')

	if trace {
		p.printf("return %t", t)
	}
	return t
}

func (p *parser) parseLineComment() node.Inline {
	if trace {
		defer p.trace("parseLineComment")()
	}

	p.nextch()
	p.nextch()

	var b bytes.Buffer
	for {
		if p.atEOL() {
			break
		}

		b.WriteRune(p.ch)
		p.nextch()
	}

	txt := b.Bytes()

	if trace {
		p.printf("return %q", txt)
	}

	return node.LineComment(txt)
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

func (p *parser) parseHanging(name string, delim string, ranked, verbatim bool) node.Block {
	if trace {
		defer p.tracef(
			"parseHanging (%s, delim=%q, ranked=%t, verbatim=%t)",
			name, delim, ranked, verbatim,
		)()
	}

	var rank uint
	if ranked {
		delim := p.ch
		for p.ch == delim {
			rank++
			p.addLead(' ')
			p.nextch()
		}
	} else {
		for i := 0; i < utf8.RuneCountInString(delim); i++ {
			p.addLead(' ')
			p.nextch() // consume delimiter
		}
	}

	newBlocks := diff(p.blocks, p.lead)
	if trace {
		p.printBlocks("reqd", p.blocks)
		p.printBlocks("lead", p.lead)
		p.printBlocks("diff", newBlocks)
	}
	defer p.open(newBlocks...)()

	reqdBlocks := p.blocks

	if verbatim {
		lines := p.parseLines(reqdBlocks)
		return &node.HangingVerbatim{name, rank, lines}
	}

	children := p.parse(reqdBlocks)
	return &node.Hanging{name, rank, children}
}

func (p *parser) parseLines(reqdBlocks []rune) [][]byte {
	if trace {
		defer p.trace("parseLines")()
	}

	var lines [][]byte

	var b strings.Builder
	for {
		n := 0 // blank lines index

		if p.atEOL() {
			for p.atEOL() && !p.atEOF {
				lines = append(lines, []byte(b.String()))
				b.Reset()

				// skip empty lines
				p.nextln()
				p.nextch()
				p.parseLead()

				n++
			}

			if p.atEOF {
				break
			}
		}

		if !p.continues(reqdBlocks) {
			if n > 1 {
				blankLines := n - 1
				var blocks []node.Block
				for i := 0; i < blankLines; i++ {
					blocks = append(blocks, &node.Line{"Line", nil})
				}
				p.ambis = n - 1
				p.stage = blocks
				lines = lines[:len(lines)-blankLines]
			}
			break
		}

		for {
			b.WriteRune(p.ch)

			if !p.nextch() {
				break
			}
		}
	}

	return lines
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

	if p.blank && onlySpacing(blocks) {
		if trace {
			p.print("return maybe (blank)")
		}
		return 2
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
			} else if j != 0 {
				j = 0
			}

			if j == i {
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

	var a []rune
	for i := len(p.lead) - 1; i >= 0; i-- {
		if p.lead[i] != ' ' && p.lead[i] != '\t' {
			a = p.lead[i+1:]
			break
		}
	}

	if trace {
		p.printBlocks("spacing", a)
	}

	return a
}

func (p *parser) parseLine(name string) node.Block {
	if trace {
		defer p.tracef("parseLine (%s)", name)()
	}

	children := p.parseInlines()

	if p.atEOL() {
		p.nextln()
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

		if p.isClosingDelimiter() {
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

func inlineEquals(opening, closing [2]rune) bool {
	if opening[1] == 0 {
		// no escape character
		return opening[0] == counterpart(closing[1])
	}
	return opening[0] == closing[1] && opening[1] == counterpart(closing[0])
}

func (p *parser) parseInline() node.Inline {
	if trace {
		defer p.trace("parseInline")()
	}

	if p.isLineComment() {
		return p.parseLineComment()
	}

	el, ok := p.inlineElems[p.ch]
	if ok {
		switch el.Type {
		case node.TypeUniform:
			if p.peekEquals(p.ch) {
				return p.parseUniform(el.Name)
			}
		case node.TypeEscaped:
			if p.isEscaped() {
				return p.parseEscaped(el.Name)
			}
		case node.TypeForward:
			return p.parseForward(el.Name)
		default:
			panic(fmt.Sprintf("parser.parseInline: unexpected node type %s (%s)", el.Type, el.Name))
		}
	}

	return p.parseText()
}

func (p *parser) isEscaped() bool {
	if p.peekEquals(p.ch) {
		return true
	}

	peek, _ := utf8.DecodeRune(p.ln)
	if peek != utf8.RuneError {
		_, ok := leftRightChars[peek]
		return ok
	}
	return false
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

	peek, _ := utf8.DecodeRune(p.ln)
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
	p.nextch()
	p.nextch()

	defer p.openInline(delim, delim)()

	children := p.parseInlines()
	if inlineEquals([2]rune{delim, delim}, p.closingDelimiter()) {
		p.nextch()
		p.nextch()
	}

	return &node.Uniform{name, children}
}

func (p *parser) parseEscaped(name string) node.Inline {
	if trace {
		defer p.tracef("parseEscaped (%s)", name)()
	}

	delim := p.ch
	p.nextch()
	escape := p.ch
	p.nextch()

	var content []byte

	if !p.atEOL() {
		var b bytes.Buffer
		for {
			if p.ch == counterpart(escape) && p.peekEquals(delim) {
				p.nextch()
				p.nextch()
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

func (p *parser) parseForward(name string) node.Inline {
	if trace {
		defer p.tracef("parseForward (%s)", name)()
		p.printInlines("inlines", p.inlines)
	}

	delim := p.ch
	p.nextch()

	var isTwoPart bool
	var offs int
	for {
		if offs > len(p.ln)-1 {
			break
		}

		r1, w := utf8.DecodeRune(p.ln[offs:])
		if r1 == utf8.RuneError {
			offs += w
			continue
		}

		if offs == 0 && p.ch == counterpart(delim) {
			if r1 == delim {
				isTwoPart = true
			}
			break
		}

		r2, _ := utf8.DecodeRune(p.ln[offs+w:])
		if r2 != utf8.RuneError && r1 == counterpart(delim) {
			if r2 == delim {
				isTwoPart = true
			}
			break
		}

		offs += w
	}

	if trace {
		p.printf("isTwoPart=%t", isTwoPart)
	}

	var content []byte
	var children []node.Inline

	if isTwoPart {
		defer p.openInline(delim, 0)()
		children = p.parseInlines()

		if inlineEquals([2]rune{delim, 0}, p.closingDelimiter()) {
			p.nextch()
		} else {
			return &node.Forward{name, content, children}
		}

		p.nextch() // consume opening delimiter of second part
	}

	if !p.atEOL() {
		var b bytes.Buffer
		for {
			if p.ch == counterpart(delim) {
				p.nextch()
				break
			}

			b.WriteRune(p.ch)

			if !p.nextch() {
				break
			}
		}

		content = b.Bytes()
	}

	return &node.Forward{name, content, children}
}

func counterpart(ch rune) rune {
	c, ok := leftRightChars[ch]
	if ok {
		return c
	}
	return ch
}

var leftRightChars = map[rune]rune{
	'(': ')',
	')': '(',
	'<': '>',
	'>': '<',
	'[': ']',
	']': '[',
	'{': '}',
	'}': '{',
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

		if p.isLineComment() {
			break
		}

		if p.isClosingDelimiter() {
			break
		}

		if p.isInlineDelimiter() {
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

func (p *parser) isClosingDelimiter() bool {
	if trace {
		defer p.trace("isClosingDelimiter")()
	}

	var ok bool
	for i := len(p.inlines) - 1; i >= 0; i-- {
		delim, escape := p.inlines[i][0], p.inlines[i][1]
		if escape == 0 && p.ch == counterpart(delim) {
			ok = true
			break
			return true
		}

		if p.ch == counterpart(escape) && p.peekEquals(delim) {
			ok = true
			break
		}
	}

	if trace {
		p.printf("return %t", ok)
	}

	return ok
}

func (p *parser) closingDelimiter() [2]rune {
	for i := len(p.inlines) - 1; i >= 0; i-- {
		delim, escape := p.inlines[i][0], p.inlines[i][1]
		if escape == 0 && p.ch == counterpart(delim) {
			return [2]rune{0, p.ch}
		}

		if p.ch == counterpart(escape) && p.peekEquals(delim) {
			return [2]rune{p.ch, delim}
		}
	}

	return [2]rune{0, 0}
}

func (p *parser) isInlineDelimiter() bool {
	el, ok := p.inlineElems[p.ch]
	if ok {
		switch el.Type {
		case node.TypeUniform:
			if p.peekEquals(p.ch) {
				return true
			}
		case node.TypeEscaped:
			if p.isEscaped() {
				return true
			}
		case node.TypeForward:
			return true
		default:
			panic(fmt.Sprintf("parser.parseText: unexpected node type %s (%s)", el.Type, el.Name))
		}
	}

	return false
}

func (p *parser) init(r io.Reader) {
	p.scnr = bufio.NewScanner(r)
	p.nextln()
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
		if !p.atEOL() && p.ch != ' ' && p.ch != '\t' {
			p.blank = false
		}

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
	p.blank = true

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

func (p *parser) peekEquals(ch rune) bool {
	r, _ := utf8.DecodeRune(p.ln)
	if r == utf8.RuneError {
		return false
	}

	return r == ch
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

func (p *parser) openInline(delim rune, escape rune) func() {
	size := len(p.inlines)
	p.inlines = append(p.inlines, [2]rune{delim, escape})

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

func (p *parser) printInlines(name string, inlines [][2]rune) {
	p.print(name + "=" + fmtInlines(inlines))
}

func fmtInlines(inlines [][2]rune) string {
	var b strings.Builder
	b.WriteString("[")
	for i, v := range inlines {
		if i > 0 {
			b.WriteString(", ")
		}

		b.WriteString(fmt.Sprintf("[del=%q esc=%q]", v[0], v[1]))
	}

	b.WriteString("]")
	return b.String()
}

func reverseInline(i [2]rune) [2]rune {
	return [2]rune{i[1], i[0]}
}

func (p *parser) tracef(format string, v ...interface{}) func() {
	return p.trace(fmt.Sprintf(format, v...))
}

func (p *parser) trace(msg string) func() {
	p.printf("%q -> %s (", p.ch, msg)
	p.tindent++

	return func() {
		p.tindent--
		p.print(")")
	}
}

func (p *parser) printf(format string, v ...interface{}) {
	p.print(fmt.Sprintf(format, v...))
}

func (p *parser) print(msg string) {
	fmt.Println(strings.Repeat("\t", p.tindent) + msg)
}
