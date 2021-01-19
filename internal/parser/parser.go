package parser

import (
	"to/internal/node"
	"to/internal/scanner"
	"to/internal/token"
)

type Parser struct {
	scanner scanner.Scanner
	errors  []error

	tok token.Token
	lit string
}

func New(scanner scanner.Scanner) *Parser {
	p := &Parser{
		scanner: scanner,
	}
	p.next()
	return p
}

func (p *Parser) Parse() []node.Block {
	var blocks []node.Block
	for p.tok != token.EOF {
		block := p.parseBlock()
		if block == nil {
			panic("parse.Parse: nil block")
		}

		blocks = append(blocks, block)
	}
	return blocks
}

func (p *Parser) parseBlock() node.Block {
	var block node.Block
	switch {
	case p.tok == token.BlockDelim:
		switch p.lit {
		case ">":
			block = p.parseBlockquote()
		case "|":
			block = p.parseParagraph()
		default:
			panic("parser.parseBlock: unsupported block delimiter " + p.lit)
		}
	default:
		block = p.parseLines()
	}
	return block
}

func (p *Parser) parseBlockquote() *node.Blockquote {
	p.next() // consume ">"

	var children []node.Block
	for {
		if p.tok == token.Newline || p.tok == token.EOF {
			break
		}

		block := p.parseBlock()
		children = append(children, block)
	}

	return &node.Blockquote{
		Children: children,
	}
}

func (p *Parser) parseParagraph() *node.Paragraph {
	p.next() // consume "|"

	var children []node.Block
	for {
		if p.tok == token.Newline || p.tok == token.EOF {
			break
		}

		block := p.parseBlock()
		children = append(children, block)
	}

	return &node.Paragraph{
		Children: children,
	}
}

func (p *Parser) parseLines() node.Lines {
	var lines node.Lines
	for {
		if p.tok == token.BlockDelim || p.tok == token.EOF {
			break
		}

		line := p.parseLine()
		lines = append(lines, line)
	}
	return lines
}

func (p *Parser) parseLine() string {
	var line string
	for p.tok != token.EOF {
		line += p.lit
		p.next()

		if p.tok == token.Newline {
			break
		}
	}
	return line
}

func (p *Parser) next() {
	p.tok, p.lit = p.scanner.Scan()
}
