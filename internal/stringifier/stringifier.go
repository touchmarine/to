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
	indent int
}

func (s *stringifier) stringify(nodes []node.Node) {
	for _, n := range nodes {
		s.enter(n)

		switch m := n.(type) {
		case node.ContentInlineChildren:
			s.write([]byte(strconv.Quote(string(m.Content())) + ", "))
			if ic := m.InlineChildren(); ic != nil {
				s.stringify(node.InlinesToNodes(ic))
			}

		case node.HeadBody:
			s.writei([]byte("Head: " + strconv.Quote(string(m.Head())) + ",\n"))
			s.writei([]byte("Body: " + strconv.Quote(string(m.Body())) + ","))

		case node.BlockChildren:
			s.stringify(node.BlocksToNodes(m.BlockChildren()))

		case node.InlineChildren:
			s.stringify(node.InlinesToNodes(m.InlineChildren()))

		case node.Content:
			s.write([]byte(strconv.Quote(string(m.Content()))))

		default:
			panic(fmt.Sprintf("stringifier.stringify: unexpected type %T", n))
		}

		s.leave(n)
	}
}

func (s *stringifier) enter(n node.Node) {
	switch n.(type) {
	case node.Block:
		s.writei([]byte(n.Node()))

		if ranked, ok := n.(node.Ranked); ok && ranked.Rank() > 0 {
			s.write([]byte(strconv.FormatUint(uint64(ranked.Rank()), 10)))
		}

		switch n.(type) {
		case node.InlineChildren:
			s.write([]byte("("))
		case node.BlockChildren, node.HeadBody:
			s.write([]byte("(\n"))
			s.indent++
		default:
			panic(fmt.Sprintf("stringifier.enter: unexpected block type %T", n))
		}
	case node.Inline:
		s.write([]byte(n.Node() + "("))
	default:
		panic(fmt.Sprintf("stringifier.enter: unexpected type %T", n))
	}
}

func (s *stringifier) leave(n node.Node) {
	switch n.(type) {
	case node.Block:
		switch n.(type) {
		case node.InlineChildren:
			s.write([]byte(")"))
		case node.BlockChildren, node.HeadBody:
			s.write([]byte("\n"))
			s.indent--
			s.writei([]byte(")\n"))
		default:
			panic(fmt.Sprintf("stringifier.leave: unexpected block type %T", n))
		}
	case node.Inline:
		s.write([]byte(")"))
	default:
		panic(fmt.Sprintf("stringifier.leave: unexpected type %T", n))
	}
}

func (s *stringifier) writei(p []byte) {
	s.write(append(bytes.Repeat([]byte("\t"), s.indent), p...))
}

func (s *stringifier) write(p []byte) {
	s.w.Write(p)
}
