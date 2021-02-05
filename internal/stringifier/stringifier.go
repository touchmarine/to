package stringifier

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"to/internal/node"
)

func Stringify(nodes ...node.Node) string {
	var b strings.Builder
	StringifyTo(&b, nodes...)
	return b.String()
}

func StringifyTo(w io.Writer, nodes ...node.Node) {
	var s stringifier
	s.w = w
	s.stringify(nodes)
}

type stringifier struct {
	w      io.Writer
	indent uint
}

func (s *stringifier) stringify(nodes []node.Node) {
	for _, n := range nodes {
		s.enter(n)

		switch m := n.(type) {
		case node.NodeChildren:
			s.stringify(m.Children())

		case node.Text:
			s.write([]byte(strconv.Quote(string(m))))

		default:
			panic(fmt.Sprintf("stringifier.stringify: unexpected type %T", m))
		}

		s.leave(n)
	}
}

func (s *stringifier) enter(n node.Node) {
	name, _ := n.Node()
	switch n.(type) {
	case node.Text:
		s.write([]byte(name + "("))
	default:
		s.writei([]byte(name + "(\n"))
		s.indent++
	}
}

func (s *stringifier) leave(n node.Node) {
	switch n.(type) {
	case node.Text:
		s.write([]byte(")"))
	default:
		s.indent--
		s.writei([]byte(")\n"))
	}
}

func (s *stringifier) writei(p []byte) {
	s.write(append(bytes.Repeat([]byte("\t"), int(s.indent)), p...))
}

func (s *stringifier) write(p []byte) {
	s.w.Write(p)
}
