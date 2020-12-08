package parser

import (
	"fmt"
	"strings"
	"to/node"
)

const trace = false

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
		doc.Children = append(doc.Children, p.parse())
		p.next()
	}

	return doc
}

func (p *Parser) parse() node.Node {
	if trace {
		defer p.trace("parse")()
	}

	return p.parseParagraph()
}

func (p *Parser) parseParagraph() *node.Paragraph {
	if trace {
		defer p.trace("parseParagraph")()
	}

	para := &node.Paragraph{}

	for p.ch != 0 {
		para.Children = append(para.Children, p.parseInline(nil)...)
		p.next()
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
			p.print(fmt.Sprintf("p.ch=%s, p.peek()=%s", string(p.ch), string(p.peek())))
		}

		switch {
		case p.ch == '_' && p.peek() == '_':
			inline = append(inline, p.parseEmphasis(delims))
		case p.ch == '*' && p.peek() == '*':
			inline = append(inline, p.parseStrong(delims))
		default:
			inline = append(inline, p.parseText())
		}

		// next is called by parslets
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

	// eat closing delimiter ('__', EOL, EOF) if not other node's delimiter - so
	// parent nodes are also closed
	if p.ch == '_' && p.peek() == '_' || !isDelimiter(p.ch) && !isDelimiter(p.peek()) {
		p.next()

		// eat both '_'
		if p.ch == '_' {
			p.next()
		}
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

	// eat closing delimiter ('**', EOL, EOF) if not other node's delimiter - so
	// parent nodes are also closed
	if p.ch == '*' && p.peek() == '*' || !isDelimiter(p.ch) && !isDelimiter(p.peek()) {
		p.next()

		// eat both '*'
		if p.ch == '*' {
			p.next()
		}
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

// byteContains determines whether needle is in haystack.
func byteContains(haystack []byte, needle byte) bool {
	for _, v := range haystack {
		if v == needle {
			return true
		}
	}
	return false
}
