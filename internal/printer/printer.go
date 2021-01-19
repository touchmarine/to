package printer

import (
	"fmt"
	"strings"
	"to/internal/node"
)

type Printer struct {
	w      strings.Builder
	indent int
}

func New() *Printer {
	return &Printer{}
}

func (p *Printer) Reset() {
	p.w.Reset()
}

func (p *Printer) Print(blocks []node.Block) string {
	p.printBlocks(blocks)
	return p.w.String()
}

func (p *Printer) printBlocks(blocks []node.Block) {
	for _, block := range blocks {
		p.printBlock(block)
	}
}

func (p *Printer) printBlock(block node.Block) {
	switch n := block.(type) {
	case *node.Paragraph:
		p.printParagraph(n)
	case node.Lines:
		p.printLines(n)
	default:
		panic(fmt.Sprintf("printer.printBlock: unsupported block type %T", block))
	}
}

func (p *Printer) printParagraph(para *node.Paragraph) {
	defer p.open("Paragraph")()
	p.printBlocks(para.Children)
}

func (p *Printer) printLines(lines node.Lines) {
	for _, line := range lines {
		p.printf("\"%s\"", line)
	}
}

func (p *Printer) open(name string) func() {
	p.println(name + "(")
	p.indent++

	return func() {
		p.indent--
		p.println("")
		p.print(")")
	}
}

func (p *Printer) printf(format string, a interface{}) {
	p.print(fmt.Sprintf(format, a))
}

func (p *Printer) println(s string) {
	p.print(s + "\n")
}

func (p *Printer) print(s string) {
	p.w.WriteString(strings.Repeat("\t", p.indent) + s)
}
