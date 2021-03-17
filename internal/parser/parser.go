package parser

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"to/internal/config"
	"to/internal/node"
	"unicode/utf8"
)

const trace = false

const tabWidth = 8

func Parse(r io.Reader) ([]node.Block, []error) {
	return ParseCustom(r, config.Default.Elements)
}

func ParseCustom(r io.Reader, elements []config.Element) ([]node.Block, []error) {
	var p parser
	p.register(elements)
	p.init(r)
	return p.parse(), p.errors
}

// parser holds the parsing state.
type parser struct {
	errors      []error
	scnr        *bufio.Scanner          // line scanner
	blockElems  map[rune]config.Element // map of block elements by delimiter
	inlineElems map[rune]config.Element // map of inline elements by delimiter

	// parsing
	ln    []byte // current line excluding EOL
	ch    rune   // current character
	atEOF bool   // at end of file

	blocks []rune // open blocks
	lead   []rune // blocks on current line

	inlines [][2]rune // open inlines

	// tracing
	tindent int // trace indentation
}

func (p *parser) register(elems []config.Element) {
	if p.blockElems == nil {
		p.blockElems = make(map[rune]config.Element)
	}
	if p.inlineElems == nil {
		p.inlineElems = make(map[rune]config.Element)
	}

	for _, el := range elems {
		switch categ := node.TypeCategory(el.Type); categ {
		case node.CategoryBlock:
			if _, ok := p.blockElems[el.Delimiter]; ok {
				log.Fatalf(
					"parser.register: block delimiter %q is already registered",
					el.Delimiter,
				)
			}

			p.blockElems[el.Delimiter] = el

		case node.CategoryInline:
			if _, ok := p.inlineElems[el.Delimiter]; ok {
				log.Fatalf(
					"parser.register: inline delimiter %q is already registered",
					el.Delimiter,
				)
			}

			p.inlineElems[el.Delimiter] = el

		default:
			panic(fmt.Sprintf(
				"parser.register: unexpected node category %s (element=%q, delimiter=%q)",
				categ.String(),
				el.Name,
				el.Delimiter,
			))
		}
	}
}

// The argument l is whether only line children are allowed.
func (p *parser) parse() []node.Block {
	if trace {
		defer p.trace("parse")()
	}

	var blocks []node.Block
	for {
		if p.atEOL() {
			for p.atEOL() && !p.atEOF {
				// skip empty lines
				p.nextln()
				p.nextch()
			}

			if p.atEOF {
				break
			}

			p.parseLead()
		}

		if p.atEOF {
			break
		}

		block := p.parseBlock()
		if block == nil {
			panic("parser.parse: nil block")
		}

		blocks = append(blocks, block)
	}

	return blocks
}

func (p *parser) parseBlock() node.Block {
	if trace {
		defer p.trace("parseBlock")()
	}

	if p.ch == '|' {
		// escape block
		p.nextch()
	} else {
		el, ok := p.blockElems[p.ch]
		if ok {
			switch el.Type {
			case node.TypeLine:
				return p.parseLine(el.Name)
			case node.TypeWalled:
				return p.parseWalled(el.Name)
			case node.TypeHanging:
				if el.MinRank <= 1 || p.consecutive() >= el.MinRank {
					return p.parseHanging(el.Name, el.Ranked)
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

func (p *parser) parseWalled(name string) node.Block {
	if trace {
		defer p.tracef("parseWalled (%s)", name)()
	}

	p.addLead(p.ch)
	defer p.open(p.ch)()

	p.nextch() // consume delimiter

	reqdBlocks := p.blocks
	children := p.parseChildren(reqdBlocks)

	return &node.Walled{name, children}
}

func (p *parser) parseHanging(name string, ranked bool) node.Block {
	if trace {
		defer p.tracef("parseHanging (%s, ranked=%t)", name, ranked)()
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
		p.addLead(' ')
		p.nextch() // consume delimiter
	}

	newBlocks := diff(p.blocks, p.lead)
	defer p.open(newBlocks...)()

	reqdBlocks := p.blocks
	children := p.parseChildren(reqdBlocks)

	return &node.Hanging{name, rank, children}
}

func (p *parser) parseChildren(reqdBlocks []rune) []node.Block {
	if trace {
		defer p.trace("parseChildren")()
	}

	var blocks []node.Block
	for {
		if p.atEOL() {
			p.nextln()
			if !p.nextch() {
				// empty line
				break
			}

			p.parseLead()

			continue
		}

		if p.ch == ' ' || p.ch == '\t' {
			p.parseLead()
		}

		if !p.continues(reqdBlocks) {
			break
		}

		blocks = append(blocks, p.parseBlock())
	}

	return blocks
}

func (p *parser) continues(blocks []rune) bool {
	if trace {
		defer p.trace("continues")()
		p.printBlocks("reqd", p.blocks)
		p.printBlocks("lead", p.lead)
	}

	var i, j int
	for {
		if i > len(blocks)-1 {
			if trace {
				p.printf("return true")
			}
			return true
		}

		if j > len(p.lead)-1 {
			if trace {
				p.print("return false (not enough blocks)")
			}
			return false
		}

		if blocks[i] == ' ' || blocks[i] == '\t' || p.lead[j] == ' ' || p.lead[j] == '\t' {
			n, m := i, j
			for ; i < len(p.blocks); i++ {
				if p.blocks[i] == ' ' || p.blocks[i] == '\t' {
					continue
				}
			}

			for ; j < len(p.lead); j++ {
				if p.lead[j] == ' ' || p.lead[j] == '\t' {
					continue
				}
			}

			x := countSpacing(spacingSeq(p.blocks[n:i]))
			y := countSpacing(spacingSeq(p.lead[m:j]))

			if y < x {
				if trace {
					p.printf("return false (lesser ident)")
				}
				return false
			}

			continue
		}

		if blocks[i] != p.lead[j] {
			if trace {
				p.printf("return false (%q != %q, i=%d, j=%d)", blocks[i], p.lead[j], i, j)
			}
			return false
		}

		i++
		j++
	}
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
				if p.atEOL() {
					p.nextln()
					p.nextch()
					p.parseLead()
				} else {
					p.nextch()
				}
				break OuterLoop
			}

			b.WriteRune(p.ch)

			if !p.nextch() {
				break
			}
		}
	}

	return &node.Fenced{name, lines}
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

func (p *parser) parseLead() {
	if trace {
		defer p.trace("parseLead")()
		p.printBlocks("reqd", p.blocks)
	}

	var lead []rune
	var i int
	for {
		if p.ch == ' ' || p.ch == '\t' {
			goto cont
		}

		if i > len(p.blocks)-1 {
			break
		}

		if p.ch != p.blocks[i] {
			break
		}

	cont:
		lead = append(lead, p.ch)

		if !p.nextch() {
			break
		}
		i++
	}

	p.addLead(lead...)

	if trace {
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
	var i int
	for i = len(new) - 1; i >= 0; i-- {
		if i < len(old) {
			break
		}
	}
	return new[i+1:]
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

		if blocks[i] == '\t' {
			b.WriteString(" (hanging)")
		}
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
