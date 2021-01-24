package parser

import (
	"fmt"
	"strconv"
	"strings"
	"to/internal/node"
	"to/internal/printer"
	"to/internal/scanner"
	"to/internal/token"
)

const trace = false

// Parser holds the parsing state.
type Parser struct {
	scanner scanner.Scanner

	tok        token.Token   // current token
	lit        string        // current literal
	openBlocks []token.Token // current open block tokens

	indent uint // tracing indentation
}

// New returns a new Parser and prepares it for parsing by setting the initial
// state.
func New(scanner scanner.Scanner) *Parser {
	p := &Parser{
		scanner: scanner,
	}
	p.next()
	return p
}

// Parse the tokens provided by the scanner until token.EOF.
func (p *Parser) Parse() []node.Block {
	var blocks []node.Block
	for p.tok != token.EOF {
		block, _ := p.parseBlock()
		if block == nil {
			panic("parse.Parse: nil block")
		}

		blocks = append(blocks, block)
	}
	return blocks
}

func (p *Parser) parseBlock() (node.Block, bool) {
	blocks := p.openBlocks // save

	var block node.Block
	switch {
	case p.tok == token.VLINE:
		block = p.parseParagraph()
	case p.tok == token.GT:
		block = p.parseBlockquote()
	case p.tok == token.TEXT:
		block = p.parseLines()
	default:
		panic("parser.parseBlock: unsupported token " + p.tok.String())
	}

	cont := true
	for i, b := range blocks {
		if i >= len(p.openBlocks) {
			cont = false
			break
		}

		if b != p.openBlocks[i] {
			cont = false
			break
		}
	}

	if trace {
		if cont {
			p.print("continue")
		} else {
			p.print("stop")
		}
	}

	return block, cont
}

func (p *Parser) parseBlockquote() *node.Blockquote {
	p.openBlocks = append(p.openBlocks, p.tok)

	if trace {
		defer p.trace("parseBlockquote")()
		p.printf("openBlocks %s", p.openBlocks)
	}

	p.next() // consume ">"

	var children []node.Block
	for {
		if p.tok == token.EOF {
			break
		}

		if p.tok == token.LINEFEED && !p.continues() {
			break
		}

		block, cont := p.parseBlock()
		children = append(children, block)
		if !cont {
			break
		}
	}

	if trace {
		p.dump("return\n", children)
	}

	return &node.Blockquote{
		Children: children,
	}
}

func (p *Parser) parseParagraph() *node.Paragraph {
	p.openBlocks = append(p.openBlocks, p.tok)

	if trace {
		defer p.trace("parseParagraph")()
		p.printf("openBlocks %s", p.openBlocks)
	}

	p.next() // consume "|"

	var children []node.Block
	for {
		if p.tok == token.EOF {
			break
		}

		if p.tok == token.LINEFEED && !p.continues() {
			break
		}

		block, cont := p.parseBlock()
		children = append(children, block)
		if !cont {
			break
		}
	}

	if trace {
		p.dump("return\n", children)
	}

	return &node.Paragraph{
		Children: children,
	}
}

func (p *Parser) parseLines() node.Lines {
	if trace {
		defer p.trace("parseLines")()
	}

	var lines node.Lines
	for {
		if p.tok == token.EOF {
			break
		}

		if p.tok == token.LINEFEED {
			if !p.continues() {
				break
			}

			// only text can continue lines
			if p.tok != token.TEXT {
				break
			}
		}

		line := p.parseLine()
		lines = append(lines, line)
	}

	if trace {
		p.dump("return\n", []node.Block{lines})
	}

	return lines
}

// continues determines whether the current block continues.
//
// If there are no open blocks, only text returns true as only the Lines block
// does not have a delimiter. Otherwise, compare the current token to the bottom
// of the open blocks:
//	If equal, compare the next token. Return true if there are no more open
//	blocks left.
//	If not equal, pop the current and following blocks from the open blocks,
//	and return false.
func (p *Parser) continues() bool {
	if p.tok != token.LINEFEED {
		panic(fmt.Sprintf("parser.continues: token %s is not LINEFEED", p.tok))
	}

	p.next() // consume LINEFEED

	// only text can continue the lines block
	blocks := p.openBlocks
	if len(blocks) == 0 {
		return p.tok == token.TEXT
	}

	for i := 0; len(blocks) > 0; i++ {
		bot := blocks[0]
		if p.tok != bot {
			p.openBlocks = p.openBlocks[:i]
			return false
		}

		blocks = blocks[1:]
		p.next()
	}

	return true
}

func (p *Parser) parseLine() string {
	var line string
	for p.tok == token.TEXT {
		line += p.lit
		p.next()
	}
	return line
}

func (p *Parser) next() {
	p.tok, p.lit = p.scanner.Scan()
}

func (p *Parser) trace(msg string) func() {
	p.printf("%s %s -> %s (", p.tok, strconv.Quote(p.lit), msg)
	p.indent++

	return func() {
		p.indent--
		p.print(")")
	}
}

func (p *Parser) dump(msg string, blocks []node.Block) {
	p.print(msg + printer.PrintIndented(blocks, p.indent+1))
}

func (p *Parser) printf(format string, a ...interface{}) {
	p.print(fmt.Sprintf(format, a...))
}

func (p *Parser) print(msg string) {
	fmt.Println(strings.Repeat("\t", int(p.indent)) + msg)
}
