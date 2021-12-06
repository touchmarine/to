package source

import (
	"fmt"
	"sort"
	"unicode/utf8"
)

// NewMap makes a new source map.
func NewMap(src []byte) *Map {
	return &Map{
		src: src,
		location: Location{
			Range: Range{
				Start: Position{
					Offset: 0,
					Line:   0,
					Column: 0,
				},
			},
		},
		lines: []int{0},
		leads: [][]int{nil},
	}
}

// Map holds source's line positions.
type Map struct {
	src      []byte
	location Location
	lines    []int   // lines offsets (index=line, value=offset)
	leads    [][]int // lead offsets (index=line, value=lead offsets)
}

// NodeRanges reports line ranges of the node that starts and ends at the given
// offsets.
func (m Map) NodeRanges(start, end int) []Range {
	lnStart, lnEnd := m.line(start), m.line(end)
	if lnStart < 0 || lnEnd < 0 {
		return nil
	}

	var ranges []Range
	d := m.delimiter(lnStart)
	for i := lnStart; i <= lnEnd; i++ {
		x := m.lineStart(i, d)
		if x < 0 {
			panic(fmt.Sprintf("line start not found (line=%d)", i))
		}
		var y int
		if i+1 < len(m.lines) {
			y = m.lines[i+1] - 1
		} else {
			// at last line -> must span until the end
			y = len(m.src)
		}
		ranges = append(ranges, Range{
			Start: m.Position(x),
			End:   m.Position(y),
		})
	}
	return ranges
}

func (m Map) delimiter(ln int) rune {
	leads := m.leads[ln]
	if len(leads) == 0 {
		return -1
	}
	offs := leads[0]
	r, _ := utf8.DecodeRune(m.src[offs:])
	if r == utf8.RuneError {
		return -1
	}
	return r
}

func (m Map) lineStart(ln int, d rune) int {
	leads := m.leads[ln]
	if d < 0 {
		// no delimiter
		if len(leads) == 0 {
			return 0
		}
		offs := leads[len(leads)-1]
		_, w := utf8.DecodeRune(m.src[offs:])
		return offs + w
	}
	for _, offs := range leads {
		ch, _ := utf8.DecodeRune(m.src[offs:])
		if d != ' ' && ch == ' ' {
			// allow spaces in-between for non-hanging elements
			continue
		}
		if ch == d {
			// found
			return offs
		}
	}
	// not found, element must have already ended
	return -1
}

// Position reports the offset's line and column in the source. Offset, line,
// and column are 0-based. If the line and column are not found, only the given
// offset is set.
func (m Map) Position(offs int) Position {
	p := Position{Offset: offs}
	if i := m.line(offs); i >= 0 {
		p.Line = i
		p.Column = offs - m.lines[i]
	}
	return p
}

func (m Map) line(offs int) int {
	return sort.Search(len(m.lines), func(i int) bool { return m.lines[i] > offs }) - 1
}

// AddLine registers a line offset. It is used by the parser to mark line
// positions in the source.
func (m *Map) AddLine(offs int) {
	m.location.Range.End = m.Position(offs)
	m.lines = append(m.lines, offs)
	m.leads = append(m.leads, nil)
}

// AddLead registers a lead offset. It is used by the parser to mark the starts
// of elements across multiple lines.
func (m *Map) AddLead(offs int) {
	l := len(m.lines)
	if l <= 0 {
		panic("cannot add lead, no lines yet")
	}
	ln := l - 1
	m.leads[ln] = append(m.leads[ln], offs)
}

// Location represents a location inside a resource, such as a line inside a
// text file.
type Location struct {
	URI   DocumentURI
	Range Range
}

// DocumentURI is an URI of a document.
//
// 	  foo://example.com:8042/over/there?name=ferret#nose
// 	  \_/   \______________/\_________/ \_________/ \__/
// 	   |           |            |            |        |
// 	scheme     authority       path        query   fragment
// 	   |   _____________________|__
// 	  / \ /                        \
// 	  urn:example:animal:ferret:nose
//
// https://microsoft.github.io/language-server-protocol/specifications/specification-3-17/#uri
// https://datatracker.ietf.org/doc/html/rfc3986
type DocumentURI string

// Range is like a selection in an editor (zero-based).
type Range struct {
	Start, End Position
}

// String returns a string representation of the range. It is intended for
// debugging and the output should not be relied upon.
func (r Range) String() string {
	s, e := r.Start, r.End
	return fmt.Sprintf(
		"%d:%d#%d-%d:%d#%d",
		s.Line, s.Column, s.Offset, e.Line, e.Column, e.Offset,
	)
}

// Position is like an 'insert' cursor in an editor (zero-based).
type Position struct {
	Offset int // zero-based
	Line   int // zero-based
	Column int // zero-based, byte-count
}
