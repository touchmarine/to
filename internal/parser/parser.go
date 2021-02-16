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

func (p *parser) parse(l bool) []node.Block {
	var blocks []node.Block
	for p.nextln() {
		blocks = append(blocks, p.parse0(l)...)
	}
	return blocks
}

// The argument l is whether only line children are allowed.
func (p *parser) parse0(l bool) []node.Block {
	var blocks []node.Block

	for p.nextch() {
		var block node.Block
		if el, ok := p.blockElems[p.ch]; ok && !l {
			switch el.Type {
			case node.TypeWalled:
				block = p.parseWalled(el.Name, el.OnlyLineChildren)
			default:
				panic("parser.parse: unexpected node type " + el.Type.String())
			}
		} else {
			block = p.parseLine()
		}

		blocks = append(blocks, block)
	}

	return blocks
}

func (p *parser) parseContinues(l bool) []node.Block {
	var blocks []node.Block

	blocks = append(blocks, p.parse0(l)...)

	// if continues

	for p.nextln() {
		blocks = append(blocks, p.parse0(l)...)
	}

	return blocks
}

func (p *parser) parseWalled(name string, l bool) node.Block {
	if trace {
		defer p.tracef("parseWalled (%s)", name)()
	}

	children := p.parseContinues(l)
	return &node.Walled{name, children}
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
		p.printf("return %q", txt)
	}

	var children []node.Inline
	children = append(children, node.Text(txt))

	return &node.Line{"Line", children}
}

func (p *parser) init(r io.Reader) {
	p.scnr = bufio.NewScanner(r)
	p.isFirst = true
}

func (p *parser) nextln() bool {
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

	return cont
}

// Encoding errors
var (
	ErrInvalidUTF8Encoding = errors.New("invalid UTF-8 encoding")
	ErrIllegalNULL         = errors.New("illegal character NULL")
	ErrIllegalBOM          = errors.New("illegal byte order mark")
)

func (p *parser) nextch() bool {
skip:
	r, w := utf8.DecodeRune(p.ln)

	var ch rune
	switch r {
	case utf8.RuneError: // encoding error
		if w == 0 {
			// p.ln is empty if r == utf8.RuneError && w == 0
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
	return true
}

func (p *parser) error(err error) {
	p.errors = append(p.errors, err)
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
