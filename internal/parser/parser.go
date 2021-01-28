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
	case p.tok == token.GRAVEACCENTS:
		block = p.parseCodeBlock()
	case p.tok == token.VLINE:
		block = p.parseParagraph()
	case p.tok == token.GT:
		block = p.parseBlockquote()
	case p.tok == token.HYPEN:
		block = p.parseListItem()
	case p.tok == token.UNDERSCORES, p.tok == token.TEXT:
		block = p.parseLine()
	default:
		panic("parser.parseBlock: unsupported token " + p.tok.String())
	}

	cont := true
	for i, b := range blocks {
		if i >= len(p.openBlocks) {
			cont = false
			break
		}

		if b.tok != p.openBlocks[i].tok {
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

	p.next() // consume HYPEN

	var children []node.Block
	for {
		if p.tok == token.EOF {
			break
		}

		if p.tok == token.LINEFEED && !p.continues(false) {
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

	p.next() // consume GT

	var children []node.Block
	for {
		if p.tok == token.EOF {
			break
		}

		if p.tok == token.LINEFEED && !p.continues(false) {
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

	p.next() // consume VLINE

	var children []node.Block
	for {
		if p.tok == token.EOF {
			break
		}

		if p.tok == token.LINEFEED && !p.continues(false) {
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

func (p *Parser) parseCodeBlock() *node.CodeBlock {
	if trace {
		defer p.trace("parseCodeBlock")()
	}

	tok, lit := p.tok, p.lit // save delimiter
	p.next()                 // consume GRAVEACCENTS

	var literals []string
	for {
		if p.tok == token.EOF {
			break
		}

		if p.tok == tok && p.lit == lit {
			p.next() // consume closing delimiter
			break
		}

		literals = append(literals, p.lit)

		if p.tok == token.LINEFEED {
			if !p.continues(true) {
				break
			}
			continue
		}

		p.next()
	}

	// add to head until the first "\n"
	inHead := true
	var hb, bb strings.Builder
	for _, l := range literals {
		if inHead {
			if l == "\n" {
				inHead = false
				continue
			}
			hb.WriteString(l)
			continue
		}

		bb.WriteString(l)
	}

	head := hb.String()
	body := bb.String()

	if trace {
		p.print("return")
		p.indent++
		p.print("Head: " + strconv.Quote(head))
		p.print("Body: " + strconv.Quote(body))
		p.indent--
	}

	return &node.CodeBlock{
		Head: head,
		Body: body,
	}
}

// continues determines whether the current block continues.
//
// The raw argument tells continues that it is parsing raw content and as such
// only parent delimiters must be at the start of the line for the block to
// continue. Any following delimiters are then already part of the raw content.
//
// If no blocks are open, we are parsing lines and only inline delimiters return
// true. Otherwise, compare the current token to the bottom  of the open blocks:
//	If equal, compare the next token. Return true if there are no more open
//	blocks left.
//	If not equal, pop the current and following blocks from the open blocks,
//	and return false.
func (p *Parser) continues(raw bool) bool {
	if p.tok != token.LINEFEED {
		panic(fmt.Sprintf("parser.continues: token %s is not LINEFEED", p.tok))
	}

	p.next() // consume LINEFEED

	blocks := p.openBlocks
	if len(blocks) == 0 {
		if raw {
			return true
		}
		// only inline delims can continue lines
		return p.tok == token.UNDERSCORES || p.tok == token.TEXT
	}

	var i int
	for {
		if len(blocks) == 0 {
			return true
		}

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

		i++
		p.next()
	}
}

func (p *Parser) parseLine() *node.Line {
	if trace {
		defer p.trace("parseLine")()
		p.dumpOpenBlocks()
	}

	if len(p.openBlocks) > 1 {
		top1 := p.openBlocks[len(p.openBlocks)-1]
		top2 := p.openBlocks[len(p.openBlocks)-2]
		if top1.tok == token.INDENT && top2.tok == token.INDENT {
			p.openBlocks = p.openBlocks[:len(p.openBlocks)-1]
		}
	}

	var children []node.Inline
	for {
		if p.tok == token.EOF {
			break
		}

		if p.tok == token.LINEFEED {
			p.continues(false) // handle open blocks
			break
		}

		inline := p.parseInline()
		if inline == nil {
			panic("parse.parseLine: nil inline")
		}

		children = append(children, inline)
	}

	return &node.Line{
		Children: children,
	}
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

func (p *Parser) parseInline() node.Inline {
	if trace {
		p.trace("parseInline")()
	}

	var inline node.Inline
	switch p.tok {
	case token.UNDERSCORES:
		p.parseEmphasis()
	case token.TEXT:
		inline = node.Text(p.lit)
		p.next()
	default:
		panic("parser.parseLine: unsupported token " + p.tok.String())
	}

	return inline
}

func (p *Parser) parseEmphasis() *node.Emphasis {
	if trace {
		p.trace("parseEmphasis")()
	}

	var children []node.Inline
	for {
		if p.tok == token.LINEFEED || p.tok == token.EOF {
			break
		}

		if p.tok == token.UNDERSCORES {
			break
		}

		children = append(children, p.parseInline())
		p.next()
	}

	return &node.Emphasis{
		Children: children,
	}
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
