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
		case node.BlockChildren:
			s.stringify(node.BlocksToNodes(m.BlockChildren()))

		case node.InlineChildren:
			s.stringify(node.InlinesToNodes(m.InlineChildren()))

		case node.Text:
			s.write([]byte(strconv.Quote(string(m))))

		default:
			panic(fmt.Sprintf("stringifier.stringify: unexpected type %T", n))
		}

		s.leave(n)
	}
}

func (s *stringifier) enter(n node.Node) {
	switch n.(type) {
	case node.BlockChildren:
		s.writei([]byte(n.Node() + "(\n"))
		s.indent++
	case node.InlineChildren, node.Inline:
		s.write([]byte(n.Node() + "("))
	default:
		panic(fmt.Sprintf("stringifier.enter: unexpected type %T", n))
	}
}

func (s *stringifier) leave(n node.Node) {
	switch n.(type) {
	case node.BlockChildren:
		s.write([]byte("\n"))
		s.indent--
		s.writei([]byte(")\n"))
	case node.InlineChildren, node.Inline:
		s.write([]byte(")"))
	default:
		panic(fmt.Sprintf("stringifier.leave: unexpected type %T", n))
	}
}

func (s *stringifier) writei(p []byte) {
	s.write(append(bytes.Repeat([]byte("\t"), int(s.indent)), p...))
}

func (s *stringifier) write(p []byte) {
	s.w.Write(p)
}
