package parser

import (
	"fmt"
	"strings"
	"to/node"
)

const trace = false

// heading types
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

	for p.ch != 0 {
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
		Children:   p.parseInline(delimiters{}),
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
		// end paragragh if heading
		if p.ch == '=' {
			break
		}

		// end paragraph if numbered heading
		if p.ch == '#' && p.peek() == '#' {
			break
		}

		if trace {
			p.print(fmt.Sprintf("p.ch=%s, p.peek()=%s", char(p.ch), char(p.peek())))
		}

		para.Children = append(para.Children, p.parseInline(delimiters{})...)
		p.next() // eat EOL
	}

	return para
}

type delimiters struct {
	single []byte // <https://koala.test>
	double []byte // **strong**
}

// parseInline parses until one of the provided delims, EOL, or EOF.
func (p *Parser) parseInline(delims delimiters) []node.Inline {
	if trace {
		defer p.trace("parseInline")()
		p.print(fmt.Sprintf(
			"single delims=%s, double delims=%s",
			delims.single,
			delims.double,
		))
	}

	inlines := []node.Inline{}
	for p.ch != '\n' && p.ch != 0 {
		if contains(delims.single, p.ch) {
			break
		}

		if p.ch == p.peek() && contains(delims.double, p.ch) &&
			contains(delims.double, p.peek()) {
			break
		}

		if trace {
			p.print(fmt.Sprintf("p.ch=%s, p.peek()=%s", char(p.ch), char(p.peek())))
		}

		switch {
		case p.ch == '_' && p.peek() == '_':
			inlines = append(inlines, p.parseEmphasis(delims))
		case p.ch == '*' && p.peek() == '*':
			inlines = append(inlines, p.parseStrong(delims))
		case p.ch == '<':
			inlines = append(inlines, p.parseLink(delims))
		default:
			inlines = append(inlines, p.parseText())
		}

		// pointers are advanced by parslets
	}

	if trace {
		p.print("return " + node.String("Inline", map[string]interface{}{
			"Children": node.InlinesToNodes(inlines),
		}, p.indent+1))
	}

	return inlines
}

// parseEmphasis parses until a '__' and consumes the closing delimiter.
func (p *Parser) parseEmphasis(delims delimiters) *node.Emphasis {
	if trace {
		defer p.trace("parseEmphasis")()
	}

	// eat opening '__'
	p.next()
	p.next()

	// no possible duplicates because p.parseInline() returns on delim match
	delims.double = append(delims.double, '_')

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
func (p *Parser) parseStrong(delims delimiters) *node.Strong {
	if trace {
		defer p.trace("parseStrong")()
	}

	// eat opening '**'
	p.next()
	p.next()

	// no possible duplicates because p.parseInline() returns on delim match
	delims.double = append(delims.double, '*')

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

// parseLink parses link.
//
// Link can consist from one or two parts:
// <link-destination> | <link-text><link-destination>
//
// Link destination is plain text and is also used as link text if link text is
// no present.
// Link text is inline content.
func (p *Parser) parseLink(delims delimiters) *node.Link {
	if trace {
		defer p.trace("parseLink")()
	}

	p.next() // eat opening '<'

	link := &node.Link{}
	isTwoPartLink := p.isTwoPartLink()

	// parse link text if a two part link
	// <link-text><link-destination>
	if isTwoPartLink {
		delims.single = append(delims.single, '>')
		link.Children = p.parseInline(delims)

		p.next() // eat closing '>' of link text
		p.next() // eat opening '<' of link destination
	}

	// parse link destination which is also link text if no link text is present
	offs := p.offset
	for p.ch != '>' && p.ch != '\n' && p.ch != 0 {
		p.next()
	}

	text := p.src[offs:p.offset]
	p.next() // eat closing '>'

	link.Destination = text

	// use link destination as link text if one part link
	if !isTwoPartLink {
		link.Children = []node.Inline{
			&node.Text{
				Value: text,
			},
		}
	}

	if trace {
		p.print("return " + link.String(p.indent+1))
	}

	return link
}

// isTwoPartLink determnies whether link consists of two consecutive parts:
// <link-text><link-destination>
func (p *Parser) isTwoPartLink() bool {
	if trace {
		defer p.trace("isTwoPartLink")()
	}

	// opening '<' already consumed

	// reset pointers to where they were before calling isTwoPartLink
	defer func(offs int) {
		p.ch = '<'
		p.offset = offs
		p.rdOffset = offs + 1
		p.next() // eat opening '<' again
	}(p.offset - 1)

	for p.ch != '>' && p.ch != '\n' && p.ch != 0 {
		p.next()
	}

	isTwoPartLink := p.ch == '>' && p.peek() == '<'

	if trace {
		p.print(fmt.Sprintf("return %t", isTwoPartLink))
	}

	return isTwoPartLink
}

// parseText parses until a delimiter, EOL, or EOF.
func (p *Parser) parseText() *node.Text {
	if trace {
		defer p.trace("parseText")()
	}

	offs := p.offset
	for p.ch != '\n' && p.ch != 0 {
		if isSingleDelim(p.ch) {
			break
		}

		if p.ch == p.peek() && isDoubleDelim(p.ch) && isDoubleDelim(p.peek()) {
			break
		}

		p.next()
	}

	text := p.src[offs:p.offset]

	if trace {
		p.print("return " + text)
	}

	return &node.Text{
		Value: text,
	}
}

func isSingleDelim(ch byte) bool {
	return ch == '<' || ch == '>'
}

func isDoubleDelim(ch byte) bool {
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

// contains determines whether needle is in haystack.
func contains(haystack []byte, needle byte) bool {
	for _, v := range haystack {
		if v == needle {
			return true
		}
	}
	return false
}
