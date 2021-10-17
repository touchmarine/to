package node

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

func Stringify(n *Node) (string, error) {
	var b strings.Builder
	if err := Fprint(&b, n); err != nil {
		return "", err
	}
	return b.String(), nil
}

// StringifyDetailed is like Stringify but includes more information, such as
// the node's Data.
func StringifyDetailed(n *Node) (string, error) {
	var b strings.Builder
	if err := fprint(&b, true, n); err != nil {
		return "", err
	}
	return b.String(), nil
}

func Print(n *Node) error {
	return Fprint(os.Stdout, n)
}

func Fprint(w io.Writer, n *Node) (err error) {
	return fprint(w, false, n)
}

type writer interface {
	io.Writer
	io.StringWriter
}

func fprint(w io.Writer, detailed bool, n *Node) error {
	if x, ok := w.(writer); ok {
		return fprint0(x, detailed, n)
	}

	buf := bufio.NewWriter(w)
	if err := fprint0(buf, detailed, n); err != nil {
		return err
	}
	return buf.Flush()
}

func fprint0(w writer, detailed bool, n *Node) error {
	p := printer{
		w:        w,
		detailed: detailed,
	}
	return p.print(n)
}

type printer struct {
	w        writer
	detailed bool // whether to print more information

	indent       int
	afterNewline bool
}

func (p *printer) print(n *Node) error {
	p.writef("%s(%s)", n.Type.String()[len("Type"):], n.Element)
	if p.detailed && n.Data != nil {
		b, err := json.Marshal(n.Data)
		if err != nil {
			panic(err)
		}
		if len(b) > 0 {
			p.writef("<%s>", string(b))
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
