package stringifier

import (
	"fmt"
	"github.com/touchmarine/to/node"
	"io"
	"strings"
)

func Stringify(n *node.Node) string {
	var b strings.Builder
	StringifyTo(&b, n)
	return b.String()
}

func StringifyTo(w io.StringWriter, n *node.Node) {
	var s stringifier
	s.w = w
	s.stringify(n)
}

type stringifier struct {
	w            io.StringWriter
	indent       int
	freshNewline bool
}

func (s *stringifier) stringify(n *node.Node) {
	s.writef("%s(%s)(", n.Type.String()[len("Type"):], n.Name)
	if !isEmpty(n) {
		s.newline()
		s.indent++
	}

	defer func() {
		if !isEmpty(n) {
			s.newline()
			s.indent--
		}
		s.write(")")
	}()

	switch n.Type {
	case node.TypeContainer, node.TypeLeaf, node.TypeWalled,
		node.TypeHanging, node.TypeRankedHanging, node.TypeUniform:
		i := 0
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if i > 0 {
				s.write(",")
				s.newline()
			}

			s.stringify(c)
			i++
		}
	case node.TypeVerbatimLine, node.TypeText, node.TypeEscaped, node.TypePrefixed:
		s.write(n.Data)
	case node.TypeVerbatimWalled, node.TypeFenced:
		s.write(n.Data)
	default:
		panic("stringifier: unexpected node type " + n.Type.String())
	}
}

func isEmpty(n *node.Node) bool {
	switch n.Type {
	case node.TypeContainer, node.TypeLeaf, node.TypeWalled,
		node.TypeHanging, node.TypeRankedHanging, node.TypeUniform:
		return n.FirstChild == nil
	case node.TypeVerbatimLine, node.TypeText, node.TypeEscaped, node.TypePrefixed,
		node.TypeVerbatimWalled, node.TypeFenced:
		return n.Data == ""
	default:
		panic("stringifier: unexpected node type " + n.Type.String())
	}
}

func (s *stringifier) newline() {
	s.w.WriteString("\n")
	s.freshNewline = true
}

func (s *stringifier) writef(format string, v ...interface{}) {
	s.write(fmt.Sprintf(format, v...))
}

func (s *stringifier) write(a string) {
	if s.freshNewline {
		s.w.WriteString(strings.Repeat("\t", s.indent) + a)
		s.freshNewline = false
	} else {
		s.w.WriteString(a)
	}
}
