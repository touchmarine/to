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
	Name                  string
	Type                  node.Type
	Delimiter             rune
	DisallowBlockChildren bool
}

var DefaultElements = []Element{
	{"Paragraph", node.TypeWalled, '|', true},
	{"Blockquote", node.TypeWalled, '>', false},
}

func Parse(r io.Reader) ([]node.Node, []error) {
	var p parser
	p.register(DefaultElements)
	p.init(r)
	return p.parse(), p.errors
}

// parser holds the parsing state.
type parser struct {
	errors []error
	scnr   *bufio.Scanner // line scanner

	// parsing
	ln []byte // current line excluding EOL
	ch rune   // current character

	blockElems map[rune]Element // map of block elements by delimiter

	// tracing
	tindent uint // trace indentation
}

func (p *parser) register(elems []Element) {
	if p.blockElems == nil {
		p.blockElems = make(map[rune]Element)
	}

	for _, el := range elems {
		switch categ := node.Category(el.Type); categ {
		case node.CategoryBlock:
			if _, ok := p.blockElems[el.Delimiter]; ok {
				log.Fatalf(
					"parser.register: delimiter %q is already registered once",
					el.Delimiter,
				)
			}

			p.blockElems[el.Delimiter] = el
		default:
			panic("parser.register: unexpected node category " + categ.String())
		}
	}
}

func (p *parser) parse() []node.Node {
	var nodes []node.Node

	for p.nextln() {
		if el, ok := p.blockElems[p.ch]; ok {
			var n node.Node
			switch el.Type {
			case node.TypeWalled:
				n = p.parseWalled(el.Name)
			default:
				panic("parser.parse: unexpected node type " + el.Type.String())
			}
			nodes = append(nodes, n)
			continue
		}

		line := p.parseLine()
		nodes = append(nodes, line)
	}

	return nodes
}

func (p *parser) parseWalled(name string) node.Node {
	if trace {
		defer p.trace("parse" + name + " (Walled)")()
	}

	p.nextch() // consume delimiter

	var children []node.Node
	children = append(children, p.parse()...)
	return node.NewBlock(name, node.TypeWalled, children)
}

func (p *parser) parseLine() node.Text {
	if trace {
		defer p.trace("parseLine")()
	}

	var b bytes.Buffer
	for p.nextch() {
		b.WriteRune(p.ch)
	}
	txt := b.Bytes()

	if trace {
		p.printf("return %q", txt)
	}

	return node.Text(txt)
}

func (p *parser) init(r io.Reader) {
	p.scnr = bufio.NewScanner(r)
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

var ErrInvalidUTF8Encoding = errors.New("invalid UTF-8 encoding")

func (p *parser) nextch() bool {
	r, w := utf8.DecodeRune(p.ln)
	if r == utf8.RuneError {
		if w == 0 {
			// empty p.ln
			return false
		} else if w == 1 {
			p.error(ErrInvalidUTF8Encoding)
		}
	}
	p.ch = r
	p.ln = p.ln[w:]
	return true
}

func (p *parser) error(err error) {
	p.errors = append(p.errors, err)
}

func (p *parser) trace(msg string) func() {
	p.printf("%q -> %s (", p.ch, msg)
	p.tindent++

	return func() {
		p.tindent--
		p.print(")")
	}
}

func (p *parser) printf(format string, a ...interface{}) {
	p.print(fmt.Sprintf(format, a...))
}

func (p *parser) print(msg string) {
	fmt.Println(strings.Repeat("\t", int(p.tindent)) + msg)
}
