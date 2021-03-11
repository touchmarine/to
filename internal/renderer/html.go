package renderer

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"to/internal/node"
)

var ElementTagNames = map[string][]string{
	"Blockquote":      []string{"blockquote"},
	"DescriptionList": []string{"dl", "dt", "dd"},
	"CodeBlock":       []string{"pre><code"},
	"Heading":         []string{"h"},

	"Emphasis": []string{"em"},
	"Strong":   []string{"strong"},
	"Code":     []string{"code"},
	"Link":     []string{"a"},
}

func Render(nodes ...node.Node) string {
	var b strings.Builder
	RenderTo(&b, nodes...)
	return b.String()
}

func RenderTo(w io.Writer, nodes ...node.Node) {
	r := renderer{w: w, nodes: nodes}
	r.render()
}

type renderer struct {
	w     io.Writer
	nodes []node.Node
	pos   int

	n        node.Node
	tagNames []string
	isLast   bool
	indent   int
}

func (r *renderer) render() {
	for r.next() {
		switch m := r.n.(type) {
		case *node.Line:
			if onlyLineComments(m.InlineChildren()) {
				continue
			}
		}

		if r.tagName() != "" {
			r.enter()
		}

		r.inside()

		if r.tagName() != "" {
			r.leave()
		}
	}
}

func (r *renderer) next() bool {
	if r.pos > len(r.nodes)-1 {
		return false
	}

	n := r.nodes[r.pos]
	tagNames := ElementTagNames[n.Node()]
	isLast := r.pos == len(r.nodes)-1

	r.n = n
	r.tagNames = tagNames
	r.isLast = isLast
	r.pos++
	return true
}

func (r *renderer) enter() {
	switch r.n.(type) {
	case node.Block:
		r.writei([]byte("<" + r.tagName()))

		if r.isRanked() {
			r.write([]byte(uintToStr(r.rank())))
		}

		r.attrs()
		r.write([]byte(">"))

		switch r.n.(type) {
		case node.InlineChildren:
		case node.BlockChildren, node.HeadBody:
			r.write([]byte("\n"))
			r.indent++
		default:
			panic(fmt.Sprintf("renderer.enter: unexpected block type %T", r.n))
		}
	case node.Inline:
		r.write([]byte("<" + r.tagName()))
		r.attrs()
		r.write([]byte(">"))
	default:
		panic(fmt.Sprintf("renderer.enter: unexpected type %T", r.n))
	}
}

func (r *renderer) attrs() {
	switch r.n.Node() {
	case "Link":
		if link, ok := r.n.(node.Content); ok {
			r.write([]byte(` href="`))
			r.write(link.Content())
			r.write([]byte(`" `))
		} else {
			panic("renderer.attrs: link does not implement Content interface")
		}
	}
}

func (r *renderer) inside() {
	switch m := r.n.(type) {
	case node.ContentInlineChildren:
		r.write(m.Content())
		r.write([]byte(", "))
		if ic := m.InlineChildren(); ic != nil {
			RenderTo(r.w, node.InlinesToNodes(ic)...)
		}

	case node.HeadBody:
		r.writei([]byte("Head: "))
		r.write(m.Head())
		r.write([]byte(",\n"))

		r.writei([]byte("Body: "))
		r.write(m.Body())
		r.write([]byte(","))

	case node.BlockChildren:
		RenderTo(r.w, node.BlocksToNodes(m.BlockChildren())...)

	case node.InlineChildren:
		RenderTo(r.w, node.InlinesToNodes(m.InlineChildren())...)

	case node.Content:
		r.write(m.Content())

	default:
		panic(fmt.Sprintf("renderer.inside: unexpected type %T", r.n))
	}

	if _, ok := r.n.(*node.Line); ok && !r.isLast {
		r.write([]byte("<br>"))
	}
}

func (r *renderer) leave() {
	switch r.n.(type) {
	case node.Block:
		switch r.n.(type) {
		case node.InlineChildren:
		case node.BlockChildren, node.HeadBody:
			r.write([]byte("\n"))
			r.indent--
		default:
			panic(fmt.Sprintf("renderer.leave: unexpected block type %T", r.n))
		}

	case node.Inline:
	default:
		panic(fmt.Sprintf("renderer.leave: unexpected type %T", r.n))
	}

	r.write([]byte("</" + r.tagName()))
	if r.isRanked() {
		r.write([]byte(uintToStr(r.rank())))
	}
	r.write([]byte(">"))
}

func (r *renderer) isRanked() bool {
	ranked, ok := r.n.(node.Ranked)
	return ok && ranked.Rank() > 0
}

func (r *renderer) rank() uint {
	ranked, ok := r.n.(node.Ranked)
	if !ok {
		panic(fmt.Sprintf("renderer.rank: node %T does not implement node.Ranked", r.n))
	}
	return ranked.Rank()
}

func (r *renderer) tagName() string {
	if len(r.tagNames) == 0 {
		return ""
	}

	return r.tagNames[0]
}

func (r *renderer) writei(p []byte) {
	r.write(append(bytes.Repeat([]byte("\t"), r.indent), p...))
}

func (r *renderer) write(p []byte) {
	r.w.Write(p)
}

func onlyLineComments(inlines []node.Inline) bool {
	for _, n := range inlines {
		if _, ok := n.(node.LineComment); !ok {
			return false
		}
	}
	return true
}

func uintToStr(u uint) string {
	return strconv.FormatUint(uint64(u), 10)
}
