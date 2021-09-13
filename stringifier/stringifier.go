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
	switch n.TypeCategory() {
	case node.CategoryBlock:
		s.writef("%s(%s)(", n.Type, n.Name)
		s.newline()
		s.indent++

		defer func() {
			s.newline()
			s.indent--
			s.write(")")
		}()
	case node.CategoryInline:
		s.writef("%s(%s)(", n.Type, n.Name)
		defer func() {
			s.write(")")
		}()
	default:
		panic("stringifier: unexpected node category " + n.TypeCategory().String())
	}

	switch n.Type {
	case node.TypeContainer, node.TypeLeaf, node.TypeWalled,
		node.TypeHanging, node.TypeRankedHanging, node.TypeUniform:
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			s.stringify(c)
		}
	case node.TypeVerbatimLine, node.TypeText, node.TypeEscaped, node.TypePrefixed:
		s.write(n.Data)
	case node.TypeVerbatimWalled, node.TypeFenced:
		s.write(n.Data)
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
