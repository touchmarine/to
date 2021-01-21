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

func (p *Printer) printBlocks(blocks []node.Block) {
	for i, block := range blocks {
		isLast := i == len(blocks)-1
		p.printBlock(block, isLast)
	}
}

func (p *Printer) printBlock(block node.Block, isLast bool) {
	switch n := block.(type) {
	case *node.Blockquote:
		p.printBlockquote(n, isLast)
	case *node.Paragraph:
		p.printParagraph(n, isLast)
	case node.Lines:
		p.printLines(n, isLast)
	default:
		panic(fmt.Sprintf("printer.printBlock: unsupported block type %T", block))
	}
}

func (p *Printer) printBlockquote(bq *node.Blockquote, isLast bool) {
	defer p.open("Blockquote", len(bq.Children) == 0, isLast)()
	p.printBlocks(bq.Children)
}

func (p *Printer) printParagraph(para *node.Paragraph, isLast bool) {
	defer p.open("Paragraph", len(para.Children) == 0, isLast)()
	p.printBlocks(para.Children)
}

func (p *Printer) printLines(lines node.Lines, isLast bool) {
	defer p.open("Lines", len(lines) == 0, isLast)()
	for i, line := range lines {
		if i > 0 {
			p.print("\n")
		}
		p.printf("%s", strconv.Quote(line))
	}
}

func (p *Printer) open(name string, isEmpty, isLast bool) func() {
	if isEmpty {
		p.println(name + "()")
		return func() {}
	}

	p.println(name + "(")
	p.indent++

	return func() {
		p.indent--
		p.print("\n") // printed separately because of indentation
		p.print(")")
		if !isLast {
			p.print("\n")
		}
	}
}

func (p *Printer) printf(format string, a interface{}) {
	p.print(fmt.Sprintf(format, a))
}

func (p *Printer) println(s string) {
	p.print(s + "\n")
}

func (p *Printer) print(s string) {
	p.w.WriteString(strings.Repeat("\t", int(p.indent)) + s)
}
