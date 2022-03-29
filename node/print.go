package node

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// PrinterMode controls the printing.
type PrinterMode int

//go:generate stringer -type=PrinterMode
const (
	PrintData     PrinterMode = 1 << iota // print node.Data
	PrintOffsets                          // print node.Start and node.End
	PrintLocation                         // print node.Location
)

// UnmarshalText decodes the given text into *PrinterMode (case-insensitive).
// It returns an error on unexpected text.
func (m *PrinterMode) UnmarshalText(text []byte) error {
	s := strings.ToLower(string(text))
	if mm, ok := validModes[s]; ok {
		*m = mm
	} else {
		return fmt.Errorf("unexpected PrinterMode value: %q", text)
	}
	return nil
}

var validModes = map[string]PrinterMode{
	strings.ToLower(PrintData.String()):     PrintData,
	strings.ToLower(PrintOffsets.String()):  PrintOffsets,
	strings.ToLower(PrintLocation.String()): PrintLocation,
}

// Print prints the string representation of the given node tree to stdout.
func Print(n *Node) error {
	return Fprint(os.Stdout, n)
}

// Fprint prints the string representation of the node tree to the writer.
func Fprint(w io.Writer, n *Node) error {
	return Printer{}.Fprint(w, n)
}

// Printer prints the string representation of node trees. Output depends on the
// values in this struct.
type Printer struct {
	Mode PrinterMode
}

type writer interface {
	io.Writer
	io.StringWriter
}

// Fprint prints the string representation of the the node tree to the writer.
func (p Printer) Fprint(w io.Writer, n *Node) error {
	if x, ok := w.(writer); ok {
		return p.fprint(x, n)
	}

	buf := bufio.NewWriter(w)
	if err := p.fprint(buf, n); err != nil {
		return err
	}
	return buf.Flush()
}

func (p Printer) fprint(w writer, n *Node) error {
	pp := printer{
		w:    w,
		mode: p.Mode,
	}
	return pp.print(n)
}

type printer struct {
	w    writer
	mode PrinterMode

	indent       int
	afterNewline bool
}

func (p *printer) print(n *Node) error {
	if p.mode&PrintOffsets != 0 {
		p.writef("%d-%d: ", n.Start, n.End)
	}
	if p.mode&PrintLocation != 0 {
		s := n.Location.Range.Start
		e := n.Location.Range.End
		p.writef("%d:%d#%d-%d:%d#%d: ", s.Line, s.Column, s.Offset, e.Line, e.Column, e.Offset)
	}
	p.writef("%s(%s)", n.Type.String(), n.Element)
	if p.mode&PrintData != 0 && n.Data != nil {
		if s := p.prettyJSON(n.Data); s != "" {
			p.writef("<%s>", s)
		}
	}
	p.write("(")

	if !isEmpty(n) {
		p.newline()
		p.indent++
	}

	defer func() {
		if !isEmpty(n) {
			p.newline()
			p.indent--
		}
		p.write(")")
	}()

	if n.Value != "" && n.FirstChild != nil {
		return fmt.Errorf("node has data and children (%s)", n)
	}

	if n.Value != "" {
		p.write(n.Value)
	} else if n.FirstChild != nil {
		i := 0
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if i > 0 {
				p.write(",")
				p.newline()
			}

			if err := p.print(c); err != nil {
				return err
			}
			i++
		}
	}

	return nil
}

func (p printer) prettyJSON(v interface{}) string {
	b, err := json.MarshalIndent(v, strings.Repeat("\t", p.indent), "\t")
	if err != nil {
		panic(err)
	}
	if len(bytes.Split(b, []byte("\n"))) <= 3 {
		b, err = json.Marshal(v)
		if err != nil {
			panic(err)
		}
	}
	return string(b)
}

func isEmpty(n *Node) bool {
	return n.FirstChild == nil && n.Value == ""
}

func (p *printer) newline() {
	p.w.WriteString("\n")
	p.afterNewline = true
}

func (p *printer) writef(format string, v ...interface{}) {
	p.write(fmt.Sprintf(format, v...))
}

func (p *printer) write(s string) {
	if p.afterNewline {
		p.w.WriteString(strings.Repeat("\t", p.indent) + s)
		p.afterNewline = false
	} else {
		p.w.WriteString(s)
	}
}
