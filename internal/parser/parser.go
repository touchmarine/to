package parser

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"to/internal/node"
	"unicode/utf8"
)

const trace = false

type Element struct {
	Name             string
	Type             node.Type
	Delimiter        rune
	OnlyLineChildren bool
}

var DefaultElements = []Element{
	// block
	{"Paragraph", node.TypeWalled, '|', true},
	{"Blockquote", node.TypeWalled, '>', false},
	{"DescriptionList", node.TypeHanging, '*', false},
	{"CodeBlock", node.TypeFenced, '`', false},

	// inline
	{"Emphasis", node.TypeUniform, '_', false},
	{"Strong", node.TypeUniform, '*', false},
	{"Code", node.TypeEscaped, '`', false},
	{"Link", node.TypeForward, '<', false},
}

func Parse(r io.Reader) ([]node.Block, []error) {
	var p parser
	p.register(DefaultElements)
	p.init(r)
	return p.parse(false), p.errors
}

// parser holds the parsing state.
type parser struct {
	errors      []error
	scnr        *bufio.Scanner   // line scanner
	blockElems  map[rune]Element // map of block elements by delimiter
	inlineElems map[rune]Element // map of inline elements by delimiter

	// parsing
	ln           []byte    // current line excluding EOL
	ch           rune      // current character
	isFirst      bool      // is first character
	atEOL        bool      // at end of line
	atEOF        bool      // at end of file
	blocks       []rune    // open blocks
	lnBlocks     []rune    // blocks on current line
	spacing      int       // current spacing
	inlines      [][2]rune // current inline closing delimiters
	closeInlines bool

	// tracing
	tindent int // trace indentation
}

func (p *parser) register(elems []Element) {
	if p.blockElems == nil {
		p.blockElems = make(map[rune]Element)
	}
	if p.inlineElems == nil {
		p.inlineElems = make(map[rune]Element)
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
func (p *parser) parse(l bool) []node.Block {
	if trace {
		defer p.trace("parse")()
	}

	var blocks []node.Block
	for {
		for p.atEOL {
			// blank line
			p.nextln()
		}

		p.parseSpacing()

		if p.atEOF {
			break
		}

		block := p.parseBlock(l)
		if block == nil {
			panic("parser.parse: nil block")
		}

		blocks = append(blocks, block)
	}
	return blocks
}

func (p *parser) parseSpacing() {
	var i int
Loop:
	for {
		switch p.ch {
		case '\t':
			i += 8
		case ' ':
			i++
		default:
			break Loop
		}

		if !p.nextch() {
			if !p.nextln() {
				break
			}
			i = 0
		}
	}

	p.spacing = i
}

func (p *parser) parseBlock(l bool) node.Block {
	if trace {
		defer p.trace("parseBlock")()
	}

	el, ok := p.blockElems[p.ch]
	if ok && !l {
		switch el.Type {
		case node.TypeWalled:
			return p.parseWalled(el.Name, el.OnlyLineChildren)
		case node.TypeHanging:
			return p.parseHanging(el.Name)
		case node.TypeFenced:
			if peek, _ := utf8.DecodeRune(p.ln); peek == p.ch {
				return p.parseFenced(el.Name)
			}
		default:
			panic(fmt.Sprintf("parser.parseBlock: unexpected node type %s (%s)", el.Type, el.Name))
		}
	}

	line := p.parseLine()
	p.nextln()
	return line
}

func (p *parser) parseWalled(name string, l bool) node.Block {
	if trace {
		defer p.tracef("parseWalled (%s, onlyLineChildren=%t)", name, l)()
	}

	p.open(p.ch)
	defer p.close(p.ch)

	p.nextch() // consume delimiter
	reqBlocks := p.blocks
	children := p.parseChildren(l, reqBlocks)

	return &node.Walled{name, children}
}

func (p *parser) parseHanging(name string) node.Block {
	if trace {
		defer p.tracef("parseHanging (%s)", name)()
	}

	p.open('\t')
	defer p.close('\t')

	p.nextch() // consume delimiter
	reqBlocks := p.blocks
	children := p.parseChildren(false, reqBlocks)

	return &node.Hanging{name, children}
}

func (p *parser) parseChildren(l bool, reqBlocks []rune) []node.Block {
	if trace {
		defer p.trace("parseChildren")()
	}

	var blocks []node.Block

	for {
		if p.atEOL {
			if !p.nextln() {
				break
			}
			if p.atEOL {
				// blank line
				break
			}
		}

		p.parseSpacing()

		if p.atEOF {
			break
		}

		if !p.continues(reqBlocks) {
			break
		}

		blocks = append(blocks, p.parseBlock(l))
	}

	return blocks
}

func (p *parser) continues(blocks []rune) bool {
	if trace {
		defer p.trace("continues")()
		p.printBlocks("reqBlocks", blocks)
		p.dumpLnBlocks()
	}

	for i := 0; i < len(blocks); i++ {
		if i > len(p.lnBlocks)-1 {
			if trace {
				p.print("return false, more required blocks")
			}
			return false
		}

		if blocks[i] != p.lnBlocks[i] {
			if trace {
				p.print("return false, not matching")
			}
			return false
		}
	}

	if trace {
		p.print("return true")
	}

	return true
}

func (p *parser) parseFenced(name string) node.Block {
	if trace {
		defer p.tracef("parseFenced (%s)", name)()
	}

	reqBlocks := p.blocks
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
		for p.atEOL {
			lines = append(lines, []byte(b.String()))
			b.Reset()

			if !p.nextln() {
				break OuterLoop
			}

			n := p.spacing
			for n > 0 {
				switch p.ch {
				case '\t':
					n -= 8
				case ' ':
					n--
				default:
					break
				}

				if !p.nextch() {
					break
				}
			}
		}

		if !p.continues(reqBlocks) {
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
				if p.atEOL {
					p.nextln()
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

func (p *parser) parseLine() node.Block {
	if trace {
		defer p.trace("parseLine")()
	}

	children := p.parseInlines()
	return &node.Line{"Line", children}
}

func (p *parser) parseInlines() []node.Inline {
	if trace {
		defer p.trace("parseInlines")()
	}

	var inlines []node.Inline
	for {
		if p.atEOL {
			break
		}

		if p.tryCloseInline() {
			break
		}

		if p.closeInlines {
			p.closeInlines = false
			break
		}

		if trace {
			p.dumpInlines()
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

	el, ok := p.inlineElems[p.ch]
	if ok {
		switch el.Type {
		case node.TypeUniform:
			if peek, _ := utf8.DecodeRune(p.ln); peek == p.ch {
				return p.parseUniform(el.Name)
			}
		case node.TypeEscaped:
			if peek, _ := utf8.DecodeRune(p.ln); isPunct(peek) {
				return p.parseEscaped(el.Name)
			}
		case node.TypeForward:
			return p.parseForward(el.Name)
		default:
			panic(fmt.Sprintf("parser.parseInline: unexpected node type %s (%s)", el.Type, el.Name))
		}
	}

	text := p.parseText()
	return text
}

func isPunct(ch rune) bool {
	// no space
	return ch >= 0x21 && ch <= 0x2F ||
		ch >= 0x3A && ch <= 0x40 ||
		ch >= 0x5B && ch <= 0x60 ||
		ch >= 0x7B && ch <= 0x7E
}

func (p *parser) parseUniform(name string) node.Inline {
	if trace {
		defer p.tracef("parseUniform (%s)", name)()
	}

	delim := p.ch
	p.nextch()
	escape := p.ch
	p.nextch()

	p.openInline(delim, escape)
	defer p.closeInline(delim, escape)
	children := p.parseInlines()

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

	if !p.atEOL {
		var b bytes.Buffer
		b.WriteRune(p.ch)
		for p.nextch() {
			peek, _ := utf8.DecodeRune(p.ln)
			if p.ch == escape && peek == delim {
				p.nextch()
				p.nextch()
				break
			}

			b.WriteRune(p.ch)
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

		if offs == 0 && p.ch == counterpart(delim) {
			if r1 == delim {
				isTwoPart = true
			}
			break
		}

		r2, _ := utf8.DecodeRune(p.ln[offs+w:])
		if r1 == counterpart(delim) {
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
		p.openInline(counterpart(delim), 0)
		children = p.parseInlines()
		p.closeInline(counterpart(delim), 0)

		p.nextch()
	}

	if !p.atEOL {
		var b bytes.Buffer
		b.WriteRune(p.ch)

		for p.nextch() {
			if p.ch == counterpart(delim) {
				p.nextch()
				break
			}

			b.WriteRune(p.ch)
		}

		content = b.Bytes()
	}

	return &node.Forward{name, content, children}
}

func counterpart(ch rune) rune {
	c, ok := counterpartPunct[ch]
	if ok {
		return c
	}
	return ch
}

var counterpartPunct = map[rune]rune{
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
	b.WriteRune(p.ch)
OuterLoop:
	for p.nextch() {
		if el, ok := p.inlineElems[p.ch]; ok {
			switch el.Type {
			case node.TypeUniform:
				if peek, _ := utf8.DecodeRune(p.ln); peek == p.ch {
					break OuterLoop
				}
			case node.TypeEscaped:
				if peek, _ := utf8.DecodeRune(p.ln); isPunct(peek) {
					break OuterLoop
				}
			case node.TypeForward:
				break OuterLoop
			default:
				panic(fmt.Sprintf("parser.parseText: unexpected node type %s (%s)", el.Type, el.Name))
			}
		}

		if p.tryCloseInline() {
			break
		}

		b.WriteRune(p.ch)
	}

	txt := b.Bytes()

	if trace {
		defer p.printf("return %q", txt)
	}

	return node.Text(txt)
}

func (p *parser) tryCloseInline() bool {
	for _, pair := range p.inlines {
		delim, escape := pair[0], pair[1]

		if escape == 0 {
			if p.ch == delim {
				p.nextch()
				p.closeInlines = true
				return true
			}
			continue
		}

		peek, _ := utf8.DecodeRune(p.ln)
		if p.ch == escape && peek == delim {
			p.nextch()
			p.nextch()
			return true
		}
	}
	return false
}

func (p *parser) init(r io.Reader) {
	p.scnr = bufio.NewScanner(r)
	p.isFirst = true
	p.nextln()
}

// nextln returns false if no lines left
func (p *parser) nextln() bool {
	if trace {
		defer p.trace("nextln")()
	}

	p.atEOL = false
	p.lnBlocks = nil

	if !p.nextln0() {
		p.atEOF = true
		return false
	}
	if !p.nextch() {
		return true
	}

	p.parseContinuations()
	return true
}

func (p *parser) parseContinuations() {
	if trace {
		defer p.trace("parseContinuations")()
		p.dumpBlocks()
		defer p.dumpLnBlocks()
	}

	for i := 0; i < len(p.blocks); i++ {
		if p.blocks[i] == '\t' && p.ch == ' ' {
			// 8 spaces equals one tab
			var i int
			for p.ch == ' ' {
				i++

				if i%8 == 0 {
					p.addLnBlock('\t')
				}

				if !p.nextch() {
					break
				}
			}

			if i >= 8 {
				continue
			}

			break
		}

		if p.ch != p.blocks[i] {
			break
		}

		p.addLnBlock(p.ch)

		if !p.nextch() {
			break
		}
	}
}

func (p *parser) nextln0() bool {
	if trace {
		defer p.trace("nextln0")()
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

skip:
	r, w := utf8.DecodeRune(p.ln)

	var ch rune
	switch r {
	case utf8.RuneError: // encoding error
		if w == 0 {
			if trace {
				p.printf("EOL")
			}

			// empty p.ln
			p.atEOL = true
			return false
		} else if w == 1 {
			p.error(ErrInvalidUTF8Encoding)
			ch = utf8.RuneError
		}
	case '\u0000': // NULL
		p.error(ErrIllegalNULL)
		ch = utf8.RuneError
	case '\uFEFF': // BOM
		if p.isFirst {
			// skip BOM if first character
			p.ln = p.ln[w:]
			goto skip
		} else {
			p.error(ErrIllegalBOM)
			ch = utf8.RuneError
		}
	default:
		ch = r
	}

	if p.isFirst {
		p.isFirst = false
	}

	p.ch = ch
	p.ln = p.ln[w:]

	if trace {
		p.printf("p.ch=%q ", p.ch)
	}

	return true
}

func (p *parser) open(ch rune) {
	p.blocks = append(p.blocks, ch)
	p.addLnBlock(ch)
}

func (p *parser) addLnBlock(ch rune) {
	p.lnBlocks = append(p.lnBlocks, ch)
}

func (p *parser) close(ch rune) {
	for i := len(p.blocks) - 1; i > -1; i-- {
		if ch == p.blocks[i] {
			p.blocks = p.blocks[:i]
			break
		}
	}
}

func (p *parser) openInline(delim rune, escape rune) {
	p.inlines = append(p.inlines, [2]rune{delim, escape})
}

func (p *parser) closeInline(delim rune, escape rune) {
	for i := len(p.inlines) - 1; i > -1; i-- {
		if escape == 0 {
			if delim == p.inlines[i][0] {
				p.inlines = p.inlines[:i]
				break
			}

			continue
		}

		if delim == p.inlines[i][0] && escape == p.inlines[i][1] {
			p.inlines = p.inlines[:i]
			break
		}
	}
}

func (p *parser) error(err error) {
	p.errors = append(p.errors, err)
}

func (p *parser) dumpBlocks() {
	p.printBlocks("p.blocks", p.blocks)
}

func (p *parser) dumpLnBlocks() {
	p.printBlocks("p.lnBlocks", p.lnBlocks)
}

func (p *parser) printBlocks(name string, blocks []rune) {
	p.print(name + "=" + joinBlocks(blocks))
}

func joinBlocks(blocks []rune) string {
	var b strings.Builder
	b.WriteString("[")
	for i, v := range blocks {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(fmt.Sprintf("%q", v))
	}
	b.WriteString("]")
	return b.String()
}

func (p *parser) dumpInlines() {
	var b strings.Builder
	b.WriteString("[")
	for i, v := range p.inlines {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(fmt.Sprintf("%q %q", v[0], v[1]))
	}
	b.WriteString("]")
	p.print("p.inlines=" + b.String())
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
