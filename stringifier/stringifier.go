package stringifier

import (
	"bytes"
	"fmt"
	"github.com/touchmarine/to/node"
	"io"
	"strconv"
	"strings"
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
		case node.Lines:
			verbatim := bytes.Join(m.Lines(), []byte("\n"))
			s.writei([]byte(strconv.Quote(string(verbatim))))

			if k, ok := n.(node.Boxed); ok {
				unboxed := k.Unbox()
				if unboxed == nil {
					s.write([]byte("\n"))
				} else {
					s.write([]byte(",\n"))
					s.stringify([]node.Node{unboxed})
				}
			} else if k, ok := n.(node.TrailingText); ok {
				t := k.TrailingText()
				if t != nil && len(t) > 0 {
					s.write([]byte(",\n"))
					s.writei([]byte(strconv.Quote(string(t))))
				}

				s.write([]byte("\n"))
			}

		case node.TrailingText:
			s.write(m.TrailingText())

		case node.BlockChildren:
			s.stringify(node.BlocksToNodes(m.BlockChildren()))

		case node.InlineChildren:
			s.stringify(node.InlinesToNodes(m.InlineChildren()))

		case node.Content:
			s.write([]byte(strconv.Quote(string(m.Content()))))

		case node.Composited:
			s.stringify(node.InlinesToNodes([]node.Inline{m.Primary(), m.Secondary()}))

		case node.Boxed:
			s.stringify([]node.Node{m.Unbox()})

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
		case node.InlineChildren, node.Content:
			s.write([]byte("("))
		case node.BlockChildren, node.Lines:
			s.write([]byte("(\n"))
			s.indent++
		default:
			panic(fmt.Sprintf("stringifier.enter: unexpected block type %T", n))
		}
	case node.Inline:
		s.write([]byte(n.Node() + "("))
	case node.Boxed:
		s.writei([]byte(n.Node() + "(\n"))
		s.indent++
	default:
		panic(fmt.Sprintf("stringifier.enter: unexpected type %T", n))
	}
}

func (s *stringifier) leave(n node.Node) {
	switch n.(type) {
	case node.Block:
		switch n.(type) {
		case node.InlineChildren, node.Content:
			s.write([]byte(")\n"))
		case node.BlockChildren, node.Lines:
			s.indent--
			s.writei([]byte(")\n"))
		default:
			panic(fmt.Sprintf("stringifier.leave: unexpected block type %T", n))
		}
	case node.Inline:
		s.write([]byte(")"))
	case node.Boxed:
		s.indent--
		s.writei([]byte(")\n"))
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
