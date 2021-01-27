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

	tok        token.Token // current token
	lit        string      // current literal
	openBlocks []tl        // current open block tokens and literals

	indent uint // tracing indentation
}

// tl is a token-literal pair.
type tl struct {
	tok token.Token
	lit string
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

skip:
	var block node.Block
	switch {
	case p.tok == token.INDENT:
		p.openBlocks = append(p.openBlocks, tl{p.tok, p.lit})
		p.next()
		goto skip
	case p.tok == token.VLINE:
		block = p.parseParagraph()
	case p.tok == token.GT:
		block = p.parseBlockquote()
	case p.tok == token.HYPEN:
		block = p.parseListItem()
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

func (p *Parser) parseListItem() *node.ListItem {
	var top *tl
	if len(p.openBlocks) > 0 {
		top = &p.openBlocks[len(p.openBlocks)-1]
	}
	if top == nil || top.tok != token.INDENT {
		p.openBlocks = append(p.openBlocks, tl{token.INDENT, ""})
	}

	if trace {
		defer p.trace("parseListItem")()
		p.dumpOpenBlocks()
	}

	p.next() // consume "-"

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

	return &node.ListItem{
		Children: children,
	}
}

func (p *Parser) parseBlockquote() *node.Blockquote {
	p.openBlocks = append(p.openBlocks, tl{p.tok, p.lit})

	if trace {
		defer p.trace("parseBlockquote")()
		p.dumpOpenBlocks()
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
	p.openBlocks = append(p.openBlocks, tl{p.tok, p.lit})

	if trace {
		defer p.trace("parseParagraph")()
		p.dumpOpenBlocks()
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
		p.dumpOpenBlocks()
	}

	if len(p.openBlocks) > 1 {
		top1 := p.openBlocks[len(p.openBlocks)-1]
		top2 := p.openBlocks[len(p.openBlocks)-2]
		if top1.tok == token.INDENT && top2.tok == token.INDENT {
			p.openBlocks = p.openBlocks[:len(p.openBlocks)-1]
		}
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
		if bot.tok != p.tok {
			p.openBlocks = p.openBlocks[:i]
			return false
		}

		if bot.tok == token.INDENT {
			botIndent := countIndent(bot.lit)
			curIndent := countIndent(p.lit)
			if curIndent <= botIndent {
				p.openBlocks = p.openBlocks[:i]
				return false
			}
		}

		blocks = blocks[1:]
		p.next()
	}

	return true
}

func countIndent(s string) uint {
	var indent uint
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '\t':
			indent += 8
		case ' ':
			indent++
		default:
			panic("parser.countIndent: unexpected char in indent")
		}
	}
	return indent
}

func (p *Parser) parseLine() string {
	var line string
	for p.tok == token.TEXT {
		line += p.lit
		p.next()
	}
	return line
}

func (p *Parser) dumpOpenBlocks() {
	var b strings.Builder
	b.WriteString("[")
	for i, ob := range p.openBlocks {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(ob.tok.String() + " " + strconv.Quote(ob.lit))
	}
	b.WriteString("]")
	p.print(b.String())
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
