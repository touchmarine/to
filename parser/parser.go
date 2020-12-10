package parser

import (
	"fmt"
	"strings"
	"to/node"
)

const trace = false

const (
	unnumberedHeading = iota
	numberedHeading
)

type Parser struct {
	// immutable stat
	src string

	// scanning state
	ch       byte // current character
	offset   int  // character offset
	rdOffset int  // reading offset (position after current character)

	indent int // trace indentation level
}

func New(src string) *Parser {
	p := &Parser{src: src}
	// initialize ch, offset, and rdOffset
	p.next()
	return p
}

// next reads the next character into p.ch.
// p.ch < 0 means end-of-file.
func (p *Parser) next() {
	if p.rdOffset < len(p.src) {
		p.ch = p.src[p.rdOffset]
	} else {
		p.ch = 0 // eof
	}
	p.offset = p.rdOffset
	p.rdOffset += 1
}

// peek returns the byte following the most recently read character without
// advancing the parser. If the parser is at EOF, peek returns 0.
func (p *Parser) peek() byte {
	if p.rdOffset < len(p.src) {
		return p.src[p.rdOffset]
	}
	return 0
}

func (p *Parser) ParseDocument() *node.Document {
	if trace {
		defer p.trace("ParseDocument")()
	}

	doc := &node.Document{}

	for {
		block := p.parseBlock()
		if block == nil {
			break
		}

		doc.Children = append(doc.Children, block)
		// pointers are advaced by p.parseBlock()
	}

	return doc
}

func (p *Parser) parseBlock() node.Node {
	if trace {
		defer p.trace("parseBlock")()
	}

	// skip blank lines
	for p.ch == '\t' || p.ch == '\n' || p.ch == ' ' {
		if trace {
			p.print(char(p.ch) + ", skip")
		}

		p.next()
	}

	switch p.ch {
	case 0:
		return nil
	case '=':
		return p.parseHeading(unnumberedHeading)
	default:
		if p.ch == '#' && p.peek() == '#' {
			return p.parseHeading(numberedHeading)
		}

		return p.parseParagraph()
	}
}

func (p *Parser) parseHeading(typ int) *node.Heading {
	var isNumbered bool
	var delim byte

	// determine heading type we are parsing
	switch typ {
	case unnumberedHeading:
		isNumbered = false
		delim = '='
	case numberedHeading:
		isNumbered = true
		delim = '#'
	default:
		panic("unsupported heading type")
	}

	if trace {
		if isNumbered {
			defer p.trace("parseNumberedHeading")()
		} else {
			defer p.trace("parseHeading")()
		}
	}

	// count heading level by counting consecutive delimiters
	level := 0
	for p.ch == delim {
		level++
		p.next()
	}

	// skip whitespace
	for p.ch == '\t' || p.ch == ' ' {
		p.next()
	}

	h := &node.Heading{
		Level:      level,
		IsNumbered: isNumbered,
		Children:   p.parseInline(nil),
	}
	// pointers are advanced by p.parseInline()

	return h
}

func (p *Parser) parseParagraph() *node.Paragraph {
	if trace {
		defer p.trace("parseParagraph")()
	}

	para := &node.Paragraph{}

	for p.ch != '\n' && p.ch != 0 {
		if p.ch == '=' {
			break
		}

		if trace {
			p.print(fmt.Sprintf("p.ch=%s, p.peek()=%s", char(p.ch), char(p.peek())))
		}

		para.Children = append(para.Children, p.parseInline(nil)...)
		p.next() // eat EOL
	}

	return para
}

// parseInline parses until one of the provided delims, EOL, or EOF.
func (p *Parser) parseInline(delims []byte) []node.Inline {
	if trace {
		defer p.trace("parseInline")()
		p.print(fmt.Sprintf("delims=%s", delims))
	}

	inline := []node.Inline{}
	for p.ch != '\n' && p.ch != 0 {
		if p.ch == p.peek() && byteContains(delims, p.ch) && byteContains(delims, p.peek()) {
			break
		}

		if trace {
			p.print(fmt.Sprintf("p.ch=%s, p.peek()=%s", char(p.ch), char(p.peek())))
		}

		switch {
		case p.ch == '_' && p.peek() == '_':
			inline = append(inline, p.parseEmphasis(delims))
		case p.ch == '*' && p.peek() == '*':
			inline = append(inline, p.parseStrong(delims))
		default:
			inline = append(inline, p.parseText())
		}

		// pointers are advanced by parslets
	}

	if trace {
		p.print("return " + node.String(node.InlinesToNodes(inline), "Inline", p.indent+1))
	}

	return inline
}

// parseEmphasis parses until a '__' and consumes the closing delimiter.
func (p *Parser) parseEmphasis(delims []byte) *node.Emphasis {
	if trace {
		defer p.trace("parseEmphasis")()
	}

	// eat opening '__'
	p.next()
	p.next()

	// no possible duplicates because parseInline returns on delim match
	delims = append(delims, '_')

	em := &node.Emphasis{
		Children: p.parseInline(delims),
	}

	// eat closing '__' if it is the closing delimiter
	if p.ch == '_' && p.peek() == '_' {
		p.next()
		p.next()
	}

	if trace {
		p.print("return " + em.String(p.indent+1))
	}

	return em
}

// parseStrong parses until a '**' and consumes it.
func (p *Parser) parseStrong(delims []byte) *node.Strong {
	if trace {
		defer p.trace("parseStrong")()
	}

	// eat opening '**'
	p.next()
	p.next()

	// no possible duplicates because parseInline returns on delim match
	delims = append(delims, '*')

	em := &node.Strong{
		Children: p.parseInline(delims),
	}

	// eat closing '**' if it is the closing delimiter
	if p.ch == '*' && p.peek() == '*' {
		p.next()
		p.next()
	}

	if trace {
		p.print("return " + em.String(p.indent+1))
	}

	return em
}

// parseText parses until a delimiter, EOL, or EOF.
func (p *Parser) parseText() *node.Text {
	if trace {
		defer p.trace("parseText")()
	}

	offs := p.offset
	for p.ch != '\n' && p.ch != 0 {
		if p.ch == p.peek() && isDelimiter(p.ch) && isDelimiter(p.peek()) {
			break
		}

		p.next()
	}

	if trace {
		p.print("return " + p.src[offs:p.offset])
	}

	return &node.Text{Value: p.src[offs:p.offset]}
}

func isDelimiter(ch byte) bool {
	return ch == '_' || ch == '*'
}

func (p *Parser) trace(msg string) func() {
	p.print(msg + " (")
	p.indent++

	return func() {
		p.indent--
		p.print(")")
	}
}

func (p *Parser) print(msg string) {
	fmt.Println(strings.Repeat(".   ", p.indent) + msg)
}

// char returns a string representation of a character.
func char(ch byte) string {
	s := string(ch)

	switch ch {
	case 0:
		s = "EOF"
	case '\t':
		s = "\\t"
	case '\n':
		s = "\\n"
	}

	return "'" + s + "'"
}

// byteContains determines whether needle is in haystack.
func byteContains(haystack []byte, needle byte) bool {
	for _, v := range haystack {
		if v == needle {
			return true
		}
	}
	return false
}
