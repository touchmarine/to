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
		lines:  []int{0},
		leads:  [][]int{nil},
		blanks: [][]bool{nil},
	}
}

// Map holds source's line positions.
type Map struct {
	src      []byte
	location Location
	lines    []int    // line offsets (index=line)
	leads    [][]int  // lead columns (index=line)
	blanks   [][]bool // leads overlay; lead=' ' if true, otherwise false
}

// NodeRanges reports line ranges of the node that starts and ends at the given
// offsets.
func (m Map) NodeRanges(start, end int) []Range {
	lnStart, lnEnd := m.line(start), m.line(end)
	if lnStart < 0 || lnEnd < 0 {
		return nil
	}

	var ranges []Range
	col := start - m.lines[lnStart]
	leadIx := m.leadIndex(lnStart, col)
	d := m.delimiter(lnStart, start)
	blank := false // whether lead=' '
	if leadIx >= 0 {
		blank = m.blanks[lnStart][leadIx]
	}
	for ln := lnStart; ln <= lnEnd; ln++ {
		var s int
		if ln == lnStart {
			// on first line, use the given start; needed for
			// inlines (e.g. for 'b' in '**a**b')
			s = start - m.lines[ln]
		} else if ln > lnStart && blank {
			s = m.startColumn(ln, leadIx, ' ')
		} else {
			s = m.startColumn(ln, leadIx, d)
		}
		if s < 0 {
			panic(fmt.Sprintf("start column not found (line=%d)", ln))
		}
		var e int
		if ln == lnEnd {
			// on last line, use the given end; needed for fenced
			// and the like (e.g. to no include 'a' in '`\n`a')
			e = end - m.lines[ln]
		} else {
			e = m.endColumn(ln)
		}
		if e < s {
			panic(fmt.Sprintf("end column less than start column (%d<%d)", e, s))
		}
		lnOffs := m.lines[ln]
		ranges = append(ranges, Range{
			Start: m.Position(lnOffs + s),
			End:   m.Position(lnOffs + e),
		})
	}
	return ranges
}

func (m Map) delimiter(ln, offs int) rune {
	col := offs - m.lines[ln]
	if m.leadIndex(ln, col) < 0 {
		// can't have a delimiter if there is no lead
		return -1
	}
	if r, _ := utf8.DecodeRune(m.src[offs:]); r != utf8.RuneError {
		return r
	}
	return -1
}

func (m Map) leadIndex(ln, col int) int {
	if i := sort.SearchInts(m.leads[ln], col); i < len(m.leads[ln]) && m.leads[ln][i] == col {
		return i
	}
	return -1
}

func (m Map) startColumn(ln, leadIx int, d rune) int {
	leads := m.leads[ln]
	if leadIx < 0 || d < 0 {
		if leadIx >= 0 || d >= 0 {
			panic(fmt.Sprintf("inconsistent lead index and delimiter; they should both be negative (leadIx=%d, d=%q)", leadIx, d))
		}
		// no delimiter
		if len(leads) == 0 {
			return 0
		}
		col := leads[len(leads)-1]
		_, w := utf8.DecodeRune(m.src[m.lines[ln]+col:])
		return col + w
	}
	trueIx := leadIx
	if d != ' ' {
		// non-hanging element
		for _, col := range leads[leadIx:] {
			if ch, _ := utf8.DecodeRune(m.src[m.lines[ln]+col:]); ch == ' ' {
				// allow spaces in-between for non-hanging elements
				trueIx++
				continue
			}
			break
		}
	}
	for _, col := range leads[trueIx:] {
		if ch, _ := utf8.DecodeRune(m.src[m.lines[ln]+col:]); ch == d {
			return col
		}
	}
	// not found, element must have already ended
	return -1
}

func (m Map) endColumn(ln int) int {
	if ln+1 < len(m.lines) {
		return m.lines[ln+1] - 1 - m.lines[ln]
	}
	// at last line -> must span until the end
	return len(m.src) - m.lines[ln]
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
	m.location.Range.End = m.Position(len(m.src))
	m.lines = append(m.lines, offs)
	m.leads = append(m.leads, nil)
	m.blanks = append(m.blanks, nil)
}

// AddLead registers a lead column and whether the lead is a space (blank). It
// is used by the parser to mark the starts of elements across multiple lines.
func (m *Map) AddLead(col int, blank bool) {
	l := len(m.lines)
	if l <= 0 {
		panic("cannot add lead, no lines yet")
	}
	ln := l - 1
	m.leads[ln] = append(m.leads[ln], col)
	m.blanks[ln] = append(m.blanks[ln], blank)
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
