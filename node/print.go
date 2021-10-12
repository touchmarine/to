package node

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

func Stringify(n *Node) (string, error) {
	var b strings.Builder
	err := Fprint(&b, n)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

// Dump is like Stringify but includes more information, such as the node's
// Data.
func Dump(n *Node) (string, error) {
	var b strings.Builder
	err := fprint(&b, true, n)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

func Print(n *Node) error {
	return Fprint(os.Stdout, n)
}

func Fprint(w io.StringWriter, n *Node) (err error) {
	return fprint(w, false, n)
}

func fprint(w io.StringWriter, detailed bool, n *Node) (err error) {
	p := printer{
		w:        w,
		detailed: detailed,
	}

	defer func() {
		if e := recover(); e != nil {
			err = e.(localError).err // repanic if not localError
		}
	}()

	p.print(n)
	return
}

type printer struct {
	w        io.StringWriter
	detailed bool // whether to print more information

	indent       int
	afterNewline bool
}

func (p *printer) print(n *Node) {
	p.writef("%s(%s)", n.Type.String()[len("Type"):], n.Element)
	if p.detailed && n.Data != nil {
		b, err := json.Marshal(n.Data)
		if err != nil {
			panic(localError{err})
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
		panic("printer: has data and children")
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

			p.print(c)
			i++
		}
	}
}

func isEmpty(n *Node) bool {
	return n.FirstChild == nil && n.Value == ""
}

func (p *printer) newline() {
	_, err := p.w.WriteString("\n")
	if err != nil {
		panic(localError{err})
	}
	p.afterNewline = true
}

func (p *printer) writef(format string, v ...interface{}) {
	p.write(fmt.Sprintf(format, v...))
}

func (p *printer) write(a string) {
	if p.afterNewline {
		_, err := p.w.WriteString(strings.Repeat("\t", p.indent) + a)
		if err != nil {
			panic(localError{err})
		}
		p.afterNewline = false
	} else {
		_, err := p.w.WriteString(a)
		if err != nil {
			panic(localError{err})
		}
	}
}

// localError wraps local errors so we can distinguish them from genuine panics.
type localError struct {
	err error
}
