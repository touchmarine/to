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
	{"Paragraph", node.TypeWalled, '|', true},
	{"Blockquote", node.TypeWalled, '>', false},
}

func Parse(r io.Reader) ([]node.Block, []error) {
	var p parser
	p.register(DefaultElements)
	p.init(r)
	return p.parse(false), p.errors
}

// parser holds the parsing state.
type parser struct {
	errors []error
	scnr   *bufio.Scanner // line scanner

	// parsing
	ln      []byte // current line excluding EOL
	ch      rune   // current character
	isFirst bool   // is first character
	atEOL   bool   // at end of line
	atEOF   bool   // at end of file
	blocks  []rune // open blocks

	blockElems map[rune]Element // map of block elements by delimiter

	// tracing
	tindent int // trace indentation
}

func (p *parser) register(elems []Element) {
	if p.blockElems == nil {
		p.blockElems = make(map[rune]Element)
	}

	for _, el := range elems {
		switch categ := node.TypeCategory(el.Type); categ {
		case node.CategoryBlock:
			if _, ok := p.blockElems[el.Delimiter]; ok {
				log.Fatalf(
					"parser.register: delimiter %q is already registered once",
					el.Delimiter,
				)
			}

			p.blockElems[el.Delimiter] = el

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

func (p *parser) parseBlock(l bool) node.Block {
	if trace {
		defer p.trace("parseBlock")()
	}

	el, ok := p.blockElems[p.ch]
	if ok && !l {
		switch el.Type {
		case node.TypeWalled:
			return p.parseWalled(el.Name, el.OnlyLineChildren)
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

	children := p.parseChildren(l)
	return &node.Walled{name, children}
}

func (p *parser) parseChildren(l bool) []node.Block {
	if trace {
		defer p.trace("parseChildren")()
	}

	var blocks []node.Block

	if p.nextch() { // consume delimiter
		blocks = append(blocks, p.parseBlock(l))
	} else {
		// at end of line, go to next line
		if !p.nextln() {
			return blocks
		}
	}

	for {
		if !p.continues() {
			break
		}

		if !p.atEOL {
			blocks = append(blocks, p.parseBlock(l))
		}

		if !p.nextln() {
			break
		}
	}

	return blocks
}

func (p *parser) continues() bool {
	if trace {
		defer p.trace("continues")()
		p.dumpBlocks()
	}

	for i := 0; i < len(p.blocks); i++ {
		if p.ch != p.blocks[i] {
			if trace {
				p.print("return false")
			}
			return false
		}

		if !p.nextch() {
			if i < len(p.blocks)-1 {
				// not last block
				if trace {
					p.print("return false (atEOL)")
				}
				return false
			}
			break
		}
	}

	if trace {
		p.print("return true")
	}

	return true
}

func (p *parser) parseLine() node.Block {
	if trace {
		defer p.trace("parseLine")()
	}

	var b bytes.Buffer
	b.WriteRune(p.ch)
	for p.nextch() {
		b.WriteRune(p.ch)
	}
	txt := b.Bytes()

	if trace {
		defer p.printf("return %q", txt)
	}

	var children []node.Inline
	children = append(children, node.Text(txt))
	return &node.Line{"Line", children}
}

func (p *parser) init(r io.Reader) {
	p.scnr = bufio.NewScanner(r)
	// TODO: skip possible BOM at beginning here, remove p.isFirst
	p.isFirst = true
	p.nextln()
}

func (p *parser) nextln() bool {
	contln := p.nextln0()
	contch := p.nextch()
	if !contln && !contch {
		p.atEOF = true
		return false
	}
	return true
}

func (p *parser) nextln0() bool {
	if trace {
		defer p.trace("nextln0")()
	}

	cont := p.scnr.Scan()
	p.ln = p.scnr.Bytes()
	p.atEOL = false

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
}

func (p *parser) close(ch rune) {
	for i := len(p.blocks) - 1; i > -1; i++ {
		if ch == p.blocks[i] {
			p.blocks = p.blocks[:i]
			break
		}
	}
}

func (p *parser) error(err error) {
	p.errors = append(p.errors, err)
}

func (p *parser) dumpBlocks() {
	var b strings.Builder
	b.WriteString("[")
	for i, bl := range p.blocks {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(fmt.Sprintf("%q", bl))
	}
	b.WriteString("]")

	p.print("p.blocks=" + b.String())
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
