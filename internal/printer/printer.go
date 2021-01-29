package printer

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"to/internal/node"
)

// Print returns a pretty string representation of blocks.
func Print(blocks []node.Block) string {
	var b strings.Builder
	New(&b).Print(blocks)
	return b.String()
}

// PrintIndented is like Print but whole string is indented to the given indent.
func PrintIndented(blocks []node.Block, indent uint) string {
	var b strings.Builder
	p := New(&b)
	p.indent = indent
	p.Print(blocks)
	return b.String()
}

// PrintInlines returns a pretty string representation of inlines.
func PrintInlines(inlines []node.Inline) string {
	var b strings.Builder
	New(&b).PrintInlines(inlines)
	return b.String()
}

// Printer holds the printing state.
type Printer struct {
	w      io.StringWriter
	indent uint
}

// New returns a new Printer.
func New(w io.StringWriter) *Printer {
	return &Printer{w: w}
}

// Print writes a pretty string representation of blocks to the printer writer.
func (p *Printer) Print(blocks []node.Block) {
	p.printBlocks(blocks)
}

// PrintInlines writes a pretty string representation of inlines to the printer
// writer.
func (p *Printer) PrintInlines(inlines []node.Inline) {
	p.printInlines(inlines)
}

func (p *Printer) printBlocks(blocks []node.Block) {
	for i, block := range blocks {
		isLast := i == len(blocks)-1
		p.printBlock(block, isLast)
	}
}

func (p *Printer) printBlock(block node.Block, isLast bool) {
	switch n := block.(type) {
	case *node.CodeBlock:
		p.printCodeBlock(n, isLast)
	case *node.ListItem:
		p.printListItem(n, isLast)
	case *node.Blockquote:
		p.printBlockquote(n, isLast)
	case *node.Paragraph:
		p.printParagraph(n, isLast)
	case *node.Line:
		p.printLine(n, isLast)
	default:
		panic(fmt.Sprintf("printer.printBlock: unsupported block type %T", n))
	}
}

func (p *Printer) printCodeBlock(cb *node.CodeBlock, isLast bool) {
	defer p.openi("CodeBlock", false, isLast)()
	p.println("Head: " + strconv.Quote(cb.Head))
	p.printi("Body: " + strconv.Quote(cb.Body))
}

func (p *Printer) printListItem(li *node.ListItem, isLast bool) {
	defer p.openi("ListItem", len(li.Children) == 0, isLast)()
	p.printBlocks(li.Children)
}

func (p *Printer) printBlockquote(bq *node.Blockquote, isLast bool) {
	defer p.openi("Blockquote", len(bq.Children) == 0, isLast)()
	p.printBlocks(bq.Children)
}

func (p *Printer) printParagraph(para *node.Paragraph, isLast bool) {
	defer p.openi("Paragraph", len(para.Children) == 0, isLast)()
	p.printBlocks(para.Children)
}

func (p *Printer) printLine(line *node.Line, isLast bool) {
	defer p.open("Line")()
	p.printInlines(line.Children)
}

func (p *Printer) printInlines(inlines []node.Inline) {
	for i, inline := range inlines {
		if i > 0 {
			p.print(" ")
		}
		p.printInline(inline)
	}
}

func (p *Printer) printInline(inline node.Inline) {
	switch n := inline.(type) {
	case *node.Emphasis:
		p.printEmphasis(n)
	case node.Text:
		p.print(strconv.Quote(string(n)))
	default:
		panic(fmt.Sprintf("printer.printInline: unsupported inline type %T", n))
	}
}

func (p *Printer) printEmphasis(em *node.Emphasis) {
	defer p.open("Emphasis")()
	p.printInlines(em.Children)
}

// openi is is like open but indented.
func (p *Printer) openi(name string, isEmpty, isLast bool) func() {
	if isEmpty {
		p.printi(name + "()")
		return func() {}
	}

	p.println(name + "(")
	p.indent++

	return func() {
		p.indent--
		p.printi("\n") // printed separately because of indentation
		p.printi(")")
		if !isLast {
			p.printi("\n")
		}
	}
}

func (p *Printer) open(name string) func() {
	p.printi(name + "(")
	return func() {
		p.print(")")
	}
}

// printi is like print but indented.
func (p *Printer) printi(s string) {
	p.print(strings.Repeat("\t", int(p.indent)) + s)
}

func (p *Printer) println(s string) {
	p.printi(s + "\n")
}

func (p *Printer) print(s string) {
	p.w.WriteString(s)
}
