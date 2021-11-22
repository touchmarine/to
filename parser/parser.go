// Package parser provides functions and types to parse Touch formatted text
// into node trees.
package parser

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/touchmarine/to/matcher"
	"github.com/touchmarine/to/node"
)

const trace = false

// Keys to values in node.Data.
const (
	KeyRank        = "rank"        // rank (level) in ranked hanging elements
	KeyOpeningText = "openingText" // opening text in fenced elements
)

// Elements is a map of element names to Elements.
type Elements map[string]Element

// Element tells the parser how to recognize an element.
type Element struct {
	Name      string    // element name
	Type      node.Type // node type
	Delimiter string    // delimiter character or a whole delimiter
	Matcher   string    // used to determine the contents of prefixed elements
}

// Parser parses Touch formatted text based on the values in this struct.
type Parser struct {
	Elements Elements    // element set
	Matchers matcher.Map // available matchers (by name)
	TabWidth int         // tab=<tabwidth> x spaces
}

// Parse parses Touch formatted text supplied by the given reader and returns
// the parsed node tree.
func (pp Parser) Parse(src io.Reader) (*node.Node, error) {
	var p parser
	p.registerElements(pp.Elements)
	p.registerMatchers(pp.Matchers)
	p.tabWidth = pp.TabWidth
	p.init(src)
	root := p.parse(nil)
	p.errors.Sort()
	return root, p.errors.Err()
}

// parser holds the parsing state.
type parser struct {
	errors         ErrorList
	src            []byte             // source
	blockMap       map[string]Element // registered elements by delimiter
	blockKeys      []string           // length-sorted (longest first) blockMap keys
	leaf           string             // leaf element name
	inlineMap      map[string]Element // registered inline elements by delimiter
	textElement    string             // text element name
	specialEscapes []string           // delimiters that do not start with a punctuation
	matchers       matcher.Map        // registered matchers by name
	tabWidth       int                // tab=tabWidth x spaces

	// parsing
	ch         rune // current character
	offset     int  // character offset
	rdOffset   int  // position after current character
	line       int  // current line
	lineOffset int  // current line offset

	blocks []rune // open blocks
	lead   []rune // blocks on current line
	blank  bool   // whether the lead is blank

	inlines []rune // open inlines

	// tracing
	indent int // trace indentation
}

func (p *parser) registerElements(elements Elements) {
	if p.blockMap == nil {
		p.blockMap = make(map[string]Element)
	}
	if p.inlineMap == nil {
		p.inlineMap = make(map[string]Element)
	}

	for _, e := range elements {
		if node.IsBlock(e.Type) {
			switch e.Type {
			case node.TypeLeaf:
				p.leaf = e.Name
			case node.TypeRankedHanging:
				p.blockMap[e.Delimiter+e.Delimiter] = e
			default:
				p.blockMap[e.Delimiter] = e
			}
		} else if node.IsInline(e.Type) {
			switch e.Type {
			case node.TypeText:
				p.textElement = e.Name
			default:
				r, _ := utf8.DecodeRuneInString(e.Delimiter)
				if !isPunct(r) {
					p.specialEscapes = append(p.specialEscapes, e.Delimiter)
				}

				runes := utf8.RuneCountInString(e.Delimiter)

				delimiter := ""
				if runes > 0 && e.Type == node.TypePrefixed {
					delimiter = e.Delimiter
				} else if runes == 1 {
					delimiter = e.Delimiter + e.Delimiter
				} else {
					panic(fmt.Sprintf(
						"parser: invalid inline delimiter %s (%s %s)",
						e.Delimiter, e.Name, e.Type,
					))
				}

				p.blockMap[delimiter] = e
				p.inlineMap[delimiter] = e
			}
		}
	}

	keys := make([]string, len(p.blockMap))
	i := 0
	for k, _ := range p.blockMap {
		keys[i] = k
		i++
	}

	// sort keys by length
	sort.Slice(keys, func(i, j int) bool {
		return len(keys[i]) > len(keys[j])
	})

	p.blockKeys = keys
}

func (p *parser) registerMatchers(m matcher.Map) {
	if p.matchers == nil {
		p.matchers = make(matcher.Map)
	}

	for k, v := range m {
		p.matchers[k] = v
	}
}

func (p *parser) parse(reqdBlocks []rune) *node.Node {
	if trace {
		defer p.trace("parse")()
	}

	start := p.pos()
	container := &node.Node{
		Type: node.TypeContainer,
	}

	end := p.pos()
	for p.ch >= 0 {
		if isSpacing(p.ch) {
			p.parseSpacing()
			if t, a := p.continuesAmbiguous(reqdBlocks); t && !a {
				// continues non ambiguously
				end = p.pos()
			}
		} else if !p.continues(reqdBlocks) {
			break
		} else if p.ch == '\n' {
			p.next()
			p.parseLead()
			if t, a := p.continuesAmbiguous(reqdBlocks); t && !a {
				// continues non ambiguously
				end = p.pos()
			}
		} else {
			b := p.parseBlock()
			if b == nil {
				panic("parser: parseBlock() returned no block")
			}

			container.AppendChild(b)
			end = p.pos()
		}
	}

	container.Location = node.Location{
		Range: node.Range{
			Start: start,
			End:   end,
		},
	}
	return container
}

func (p *parser) parseBlock() *node.Node {
	if trace {
		defer p.trace("parseBlock")()
	}

	if !p.isEscape() {
		el, matchesBlock := p.matchBlock()
		if matchesBlock {
			switch el.Type {
			case node.TypeVerbatimLine:
				return p.parseVerbatimLine(el.Name, el.Delimiter)
			case node.TypeWalled:
				return p.parseWalled(el.Name)
			case node.TypeVerbatimWalled:
				return p.parseVerbatimWalled(el.Name)
			case node.TypeHanging:
				return p.parseHanging(el.Name, el.Delimiter)
			case node.TypeRankedHanging:
				return p.parseRankedHanging(el.Name, el.Delimiter)
			case node.TypeFenced:
				return p.parseFenced(el.Name)
			default:
				panic(fmt.Sprintf("parser.parseBlock: unexpected node type %s (%s)", el.Type, el.Name))
			}
		}
	}

	return p.parseLeaf(p.leaf)
}

// matchBlock determines which, if any, block is at the current offset.
//
// If multiple blocks match, matchBlock selects the one with the longest
// delimiter. If an inline element is more specific (longer delimiter) than all
// the matched blocks , matchBlock returns false.
func (p *parser) matchBlock() (Element, bool) {
	if trace {
		defer p.trace("matchBlock")()
	}

	// iterate through length-sorted (longest first) delimiters to prevent
	// clashes, e.g., "==" has precedence over "="
	for _, d := range p.blockKeys {
		if p.hasPrefix([]byte(d)) {
			e := p.blockMap[d]

			if node.IsBlock(e.Type) {
				if trace {
					p.printf("return true (%s)", e.Name)
				}

				return e, true
			} else if node.IsInline(e.Type) {
				if trace {
					p.print("return false, inline")
				}

				return Element{}, false
			}
		}
	}

	if trace {
		p.print("return false")
	}

	return Element{}, false
}

// hasPrefix determines whether b matches source from offset.
func (p *parser) hasPrefix(b []byte) bool {
	return bytes.HasPrefix(p.src[p.offset:], b)
}

func (p *parser) parseWalled(name string) *node.Node {
	if trace {
		defer p.tracef("parseWalled (%s)", name)()
	}

	start := p.pos()
	p.addLead(p.ch)
	defer p.open(p.ch)()

	p.next() // consume delimiter

	reqdBlocks := p.blocks
	children := p.parse(reqdBlocks)
	end := children.Location.Range.End

	n := &node.Node{
		Element: name,
		Type:    node.TypeWalled,
		Location: node.Location{
			Range: node.Range{
				Start: start,
				End:   end,
			},
		},
	}
	n.AppendChild(children)
	return n
}

func (p *parser) parseVerbatimWalled(name string) *node.Node {
	if trace {
		defer p.tracef("parseVerbatimWalled (%s)", name)()
	}

	start := p.pos()
	p.addLead(p.ch)
	defer p.open(p.ch)()

	p.next() // consume delimiter

	contentStart := p.pos()
	end := p.pos()
	var lines [][]byte
	for p.ch >= 0 && p.continues(p.blocks) {
		offs := p.offset
		for p.ch >= 0 && p.ch != '\n' {
			p.next()
		}
		lines = append(lines, p.src[offs:p.offset])
		end = p.pos()
		if p.ch == '\n' {
			p.next()
			p.parseLead()
		}
	}

	n := &node.Node{
		Element: name,
		Type:    node.TypeVerbatimWalled,
		Location: node.Location{
			Range: node.Range{
				Start: start,
				End:   end,
			},
		},
	}
	if content := bytes.Join(lines, []byte("\n")); len(content) > 0 {
		p.setTextContent(n, string(content), contentStart, end)
	}
	return n
}

func (p *parser) parseHanging(name, delim string) *node.Node {
	if trace {
		defer p.tracef("parseHanging (%s, delim=%q)", name, delim)()
	}

	start := p.pos()
	c := utf8.RuneCountInString(delim)
	p.addLead([]rune(strings.Repeat(" ", c))...)

	for i := 0; i < c; i++ {
		// consume delimiter
		p.next()
	}

	children := p.parseHanging0()
	end := children.Location.Range.End
	n := &node.Node{
		Element: name,
		Type:    node.TypeHanging,
		Location: node.Location{
			Range: node.Range{
				Start: start,
				End:   end,
			},
		},
	}
	n.AppendChild(children)
	return n
}

func (p *parser) parseRankedHanging(name, delim string) *node.Node {
	if trace {
		defer p.tracef("parseRankedHanging (%s, delim=%q)", name, delim)()
	}

	start := p.pos()
	var rank int
	d := p.ch
	for p.ch == d {
		// count rank and consume delimiter
		rank++
		p.next()
	}
	p.addLead([]rune(strings.Repeat(" ", rank))...)

	children := p.parseHanging0()
	end := children.Location.Range.End
	n := &node.Node{
		Element: name,
		Type:    node.TypeRankedHanging,
		Data: node.Data{
			KeyRank: rank,
		},
		Location: node.Location{
			Range: node.Range{
				Start: start,
				End:   end,
			},
		},
	}
	n.AppendChild(children)
	return n
}

func (p *parser) parseHanging0() *node.Node {
	if trace {
		defer p.trace("parseHanging0")()
	}

	newBlocks := p.diff(p.blocks, p.lead)
	if trace {
		p.printDelims("reqd", p.blocks)
		p.printDelims("lead", p.lead)
		p.printDelims("diff", newBlocks)
	}
	defer p.open(newBlocks...)()

	reqdBlocks := p.blocks
	return p.parse(reqdBlocks)
}

func (p *parser) continues(blocks []rune) bool {
	t, _ := p.continuesAmbiguous(blocks)
	return t
}

func (p *parser) continuesAmbiguous(blocks []rune) (bool, bool) {
	if trace {
		defer p.trace("continues")()
		p.printDelims("reqd", blocks)
		p.printDelims("lead", p.lead)
	}

	if p.blank && onlySpacing(blocks) {
		if trace {
			p.print("return true (blank, ambiguous)")
		}
		return true, true
	}

	var i, j int
	for {
		if i > len(blocks)-1 {
			if trace {
				p.print("return true")
			}
			return true, false
		}

		if j > len(p.lead)-1 {
			if onlySpacing(blocks[i:]) && len(p.lead) > 0 &&
				(p.ch < 0 || p.ch == '\n' || p.ch == ' ' || p.ch == '\t') {
				if trace {
					p.print("return true (ambiguous)")
				}
				return true, true
			}

			if trace {
				p.print("return false (not enough blocks)")
			}
			return false, false
		}

		if blocks[i] == ' ' || blocks[i] == '\t' || p.lead[j] == ' ' || p.lead[j] == '\t' {
			n, m := i, j
			for i < len(blocks) {
				if blocks[i] == ' ' || blocks[i] == '\t' {
					i++
				} else {
					break
				}
			}

			for j < len(p.lead) {
				if p.lead[j] == ' ' || p.lead[j] == '\t' {
					j++
				} else {
					break
				}
			}

			x := p.countSpacing(spacingSeq(blocks[n:i]))
			y := p.countSpacing(spacingSeq(p.lead[m:j]))

			if y < x {
				if trace {
					p.print("return false (lesser ident)")
				}
				return false, false
			}

			continue
		}

		if blocks[i] != p.lead[j] {
			if trace {
				p.printf("return false (%q != %q, i=%d, j=%d)", blocks[i], p.lead[j], i, j)
			}
			return false, false
		}

		i++
		j++
	}
}

func onlySpacing(a []rune) bool {
	return len(spacingSeq(a)) == len(a)
}

func spacingSeq(a []rune) []rune {
	for i, v := range a {
		if isSpacing(v) {
			continue
		}
		return a[:i]
	}
	return a
}

func lastSpacingSeq(a []rune) []rune {
	for i := len(a) - 1; i >= 0; i-- {
		if !isSpacing(a[i]) {
			a = a[i+1:]
			break
		}
	}
	return a
}

func (p *parser) parseFenced(name string) *node.Node {
	if trace {
		defer p.tracef("parseFenced (%s)", name)()
	}

	start := p.pos()
	reqdBlocks := p.blocks
	openSpacing := p.diffSpacing(lastSpacingSeq(p.blocks), lastSpacingSeq(p.lead))
	defer p.open(openSpacing...)()

	delim := p.ch
	// consume delimiter
	p.next()

	escaped := p.ch == '\\'
	if escaped {
		p.next()
	}

	// parse text on opening line
	o := p.offset
	for p.ch >= 0 && p.ch != '\n' {
		p.next()
	}
	openingText := string(p.src[o:p.offset])
	p.next()
	p.parseLead()

	var lines [][]byte
	textStart := p.pos()
	textEnd := p.pos()
	end := p.pos()
	for p.continues(reqdBlocks) {
		if !escaped && p.ch == delim || escaped && p.ch == '\\' && p.peek() == delim {
			// closing delimiter
			if escaped {
				p.next()
			}
			p.next()
			end = p.pos()

			if p.continues(reqdBlocks) {
				for p.ch >= 0 && p.ch != '\n' {
					// consume closing line
					p.next()
				}
				p.next()
				p.parseLead()
				p.parseSpacing()
			}
			break
		}

		offs := p.offset
		for p.ch >= 0 && p.ch != '\n' {
			p.next()
		}
		textEnd = p.pos()
		end = p.pos()

		if p.ch < 0 || p.ch == '\n' {
			// leading spacing that is part of the element
			spacing := p.diffSpacing(lastSpacingSeq(p.blocks), lastSpacingSeq(p.lead))
			l := p.src[offs:p.offset]
			line := append([]byte(string(spacing)), l...)
			lines = append(lines, line)
			end = p.pos()
			if p.ch < 0 {
				break
			}
			// p.ch='\n'
			p.next()
			p.parseLead()
		}

	}

	n := &node.Node{
		Element: name,
		Type:    node.TypeFenced,
		Data: node.Data{
			KeyOpeningText: openingText,
		},
		Location: node.Location{
			Range: node.Range{
				Start: start,
				End:   end,
			},
		},
	}
	if text := bytes.Join(lines, []byte("\n")); len(text) > 0 {
		p.setTextContent(n, string(text), textStart, textEnd)
	}
	return n
}

func (p *parser) consumeLine() []byte {
	if trace {
		defer p.trace("consumeLine")()
	}

	var b bytes.Buffer

	for p.ch >= 0 && p.ch != '\n' {
		b.WriteRune(p.ch)

		p.next()
	}

	if trace {
		p.printf("return %q", b.Bytes())
	}

	return b.Bytes()
}

// a=old spacing
// b=new spacing
func (p *parser) diffSpacing(a, b []rune) []rune {
	x := p.countSpacing(a)
	y := p.countSpacing(b)

	if y == x {
		return nil
	} else if y > x {
		var c []rune

		n := y - x
		for i := len(b) - 1; i >= 0; i-- {
			if n <= 0 {
				break
			}

			w := p.countSpacing([]rune{b[i]})
			if w > n {
				for j := 0; j < n; j++ {
					c = append(c, ' ')
				}
				break
			}
			c = append(c, b[i])
			n -= w
		}

		return c
	}

	return nil
}

func (p *parser) countSpacing(s []rune) int {
	var i int
	for _, ch := range s {
		switch ch {
		case ' ':
			i++
		case '\t':
			i += p.tabWidth
		default:
			panic(fmt.Sprintf("countSpacing: got %q, want ' ' or '\t'", ch))
		}
	}
	return i
}

func (p *parser) parseVerbatimLine(name, delim string) *node.Node {
	if trace {
		defer p.tracef("parseVerbatimLine (%s, delim=%q)", name, delim)()
	}

	start := p.pos()
	for i := 0; i < utf8.RuneCountInString(delim); i++ {
		// consume delimiter
		p.next()
	}
	contentStart := p.pos()
	offs := p.offset
	for p.ch >= 0 && p.ch != '\n' {
		p.next()
	}
	content := p.src[offs:p.offset]
	end := p.pos()

	// prepare next line
	p.next()
	p.parseLead()
	p.parseSpacing()

	n := &node.Node{
		Element: name,
		Type:    node.TypeVerbatimLine,
		Location: node.Location{
			Range: node.Range{
				Start: start,
				End:   end,
			},
		},
	}
	if len(content) > 0 {
		p.setTextContent(n, string(content), contentStart, end)
	}
	return n
}

func (p *parser) pos() node.Position {
	return node.Position{
		Offset: p.offset,
		Line:   p.line,
		Column: p.offset - p.lineOffset,
	}
}

func (p *parser) parseLeaf(name string) *node.Node {
	if trace {
		defer p.tracef("parseLeaf %s", name)()
	}

	start := p.pos()
	children, _ := p.parseInlines()
	end := children.Location.Range.End
	n := &node.Node{
		Element: name,
		Type:    node.TypeLeaf,
		Location: node.Location{
			Range: node.Range{
				Start: start,
				End:   end,
			},
		},
	}
	n.AppendChild(children)
	return n
}

func (p *parser) parseInlines() (*node.Node, bool) {
	if trace {
		defer p.trace("parseInlines")()
	}

	start := p.pos()
	container := &node.Node{
		Type: node.TypeContainer,
	}

	cont := true
	end := p.pos()
	for p.ch >= 0 {
		if p.closingDelimiter() >= 0 {
			break
		}

		var inline *node.Node
		inline, cont = p.parseInline()
		if inline != nil {
			container.AppendChild(inline)
		}

		if !cont {
			break
		}
	}

	if container.LastChild != nil {
		end = container.LastChild.Location.Range.End
	}
	container.Location = node.Location{
		Range: node.Range{
			Start: start,
			End:   end,
		},
	}
	return container, cont
}

func (p *parser) parseInline() (*node.Node, bool) {
	if trace {
		defer p.trace("parseInline")()
	}

	if !p.isEscape() {
		el, ok := p.matchInline()
		if ok {
			switch el.Type {
			case node.TypeUniform:
				return p.parseUniform(el.Name)
			case node.TypeEscaped:
				return p.parseEscaped(el.Name)
			case node.TypePrefixed:
				return p.parsePrefixed(el.Name, el.Delimiter, el.Matcher)
			default:
				panic(fmt.Sprintf("parser.parseInline: unexpected node type %s (%s)", el.Type, el.Name))
			}
		}
	}

	return p.parseText(p.textElement)
}

func (p *parser) isEscape() bool {
	if trace {
		defer p.trace("isEscape")()
	}

	if p.ch == '\\' {
		if isPunct(p.peek()) {
			if trace {
				p.print("return true")
			}

			return true
		} else {
			for _, escape := range p.specialEscapes {
				if p.hasPrefix([]byte("\\" + escape)) {
					if trace {
						p.print("return true")
					}

					return true
				}
			}
		}
	}

	if trace {
		p.print("return false")
	}

	return false
}

// isPunct determines whether ch is an ASCII punctuation character.
func isPunct(ch rune) bool {
	return ch >= 0x21 && ch <= 0x2F ||
		ch >= 0x3A && ch <= 0x40 ||
		ch >= 0x5B && ch <= 0x60 ||
		ch >= 0x7B && ch <= 0x7E
}

func (p *parser) parseUniform(name string) (*node.Node, bool) {
	if trace {
		defer p.tracef("parseUniform (%s)", name)()
		p.printDelims("inlines", p.inlines)
	}

	start := p.pos()
	delim := p.ch
	// consume delimiter
	p.next()
	p.next()

	defer p.openInline(delim)()

	children, cont := p.parseInlines()

	if p.closingDelimiter() == counterpart(delim) {
		// consume closing delimiter
		p.next()
		p.next()
	}

	end := p.pos()
	n := &node.Node{
		Element: name,
		Type:    node.TypeUniform,
		Location: node.Location{
			Range: node.Range{
				Start: start,
				End:   end,
			},
		},
	}
	n.AppendChild(children)
	return n, cont
}

func (p *parser) parseEscaped(name string) (*node.Node, bool) {
	if trace {
		defer p.tracef("parseEscaped (%s)", name)()
	}

	start := p.pos()
	delim := p.ch
	c := string(counterpart(delim))
	closing := c + c

	// consume delimiter
	p.next()
	p.next()

	isEscaped := p.ch == '\\'
	if isEscaped {
		closing = `\` + closing
		p.next()
	}

	cont := true
	end := p.pos()
	textStart := p.pos()
	textEnd := p.pos()
	var lines [][]byte
	offs := p.offset
	for {
		if p.ch < 0 || p.ch == '\n' || p.isEscapedClosingDelimiter(closing, isEscaped) {
			line := p.src[offs:p.offset]
			lines = append(lines, line)

			if p.ch < 0 {
				break
			} else if p.ch == '\n' {
				p.next()
				p.parseLead()
				p.parseSpacing()

				if _, matchesBlock := p.matchBlock(); p.ch == '\n' || matchesBlock || !p.continues(p.blocks) {
					cont = false
					break
				}

				offs = p.offset
			} else if p.isEscapedClosingDelimiter(closing, isEscaped) {
				textEnd = p.pos()
				for i := 0; i < utf8.RuneCountInString(closing); i++ {
					p.next()
				}
				end = p.pos()
				break
			} else {
				panic("unexpected case")
			}
		} else {
			p.next()
			end = p.pos()
			textEnd = p.pos()
		}
	}

	n := &node.Node{
		Element: name,
		Type:    node.TypeEscaped,
		Location: node.Location{
			Range: node.Range{
				Start: start,
				End:   end,
			},
		},
	}
	txt := bytes.Join(lines, []byte("\n"))
	if trace {
		defer p.printf("return %q", txt)
	}
	if len(txt) > 0 {
		p.setTextContent(n, string(txt), textStart, textEnd)
	}
	return n, cont
}

func (p *parser) isEscapedClosingDelimiter(closing string, escaped bool) bool {
	var x []rune
	if escaped {
		x = append([]rune{p.ch}, []rune{p.peek(), p.peek2()}...)
	} else {
		x = []rune{p.ch, p.peek()}
	}
	return cmpRunes(x, []rune(closing))
}

// cmpRunes determines whether a and b have the same values.
func cmpRunes(a, b []rune) bool {
	if len(a) != len(b) {
		return false
	}

	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func counterpart(ch rune) rune {
	c, ok := leftRightChars[ch]
	if ok {
		return c
	}
	return ch
}

var leftRightChars = map[rune]rune{
	'{': '}',
	'}': '{',
	'[': ']',
	']': '[',
	'(': ')',
	')': '(',
	'<': '>',
	'>': '<',
}

func (p *parser) parsePrefixed(name, prefix string, matcher string) (*node.Node, bool) {
	if trace {
		defer p.tracef("parsePrefixed (%s, prefix=%q, matcher=%q)", name, prefix, matcher)()
	}

	start := p.pos()
	// consume prefix
	for i := 0; i < utf8.RuneCountInString(prefix); i++ {
		p.next()
	}
	textStart := p.pos()

	if matcher == "" {
		end := p.pos()
		return &node.Node{
			Element: name,
			Type:    node.TypePrefixed,
			Location: node.Location{
				Range: node.Range{
					Start: start,
					End:   end,
				},
			},
		}, true
	}

	m, ok := p.matchers[matcher]
	if !ok {
		panic("parser: matcher " + matcher + " not found")
	}

	w := m.Match(p.src[p.offset:])
	offs := p.offset
	offsetEnd := p.offset + w
	for p.offset < offsetEnd {
		// consume match
		p.next()
	}
	content := p.src[offs:offsetEnd]

	end := p.pos()
	n := &node.Node{
		Element: name,
		Type:    node.TypePrefixed,
		Location: node.Location{
			Range: node.Range{
				Start: start,
				End:   end,
			},
		},
	}
	if len(content) > 0 {
		p.setTextContent(n, string(content), textStart, end)
	}
	return n, true
}

func (p parser) setTextContent(n *node.Node, text string, start, end node.Position) {
	n.AppendChild(&node.Node{
		Element: p.textElement,
		Type:    node.TypeText,
		Value:   text,
		Location: node.Location{
			Range: node.Range{
				Start: start,
				End:   end,
			},
		},
	})
}

func (p *parser) parseText(name string) (*node.Node, bool) {
	if trace {
		defer p.tracef("parseText (%s)", name)()
	}

	start := p.pos()
	cont := true
	end := p.pos()
	var lines [][]byte
	var escapes []int
	offs := p.offset
	for {
		isEscape := p.isEscape()
		_, matchesInline := p.matchInline()
		if p.ch < 0 || p.ch == '\n' || !isEscape && (p.closingDelimiter() >= 0 || matchesInline) {
			line := p.src[offs:p.offset]
			for i := len(escapes) - 1; i >= 0; i-- { // reverse so we don't have to account for removed chars
				x := escapes[i] - offs                 // escape char position in slice
				line = append(line[:x], line[x+1:]...) // remove escape char
			}
			escapes = nil
			line = bytes.TrimRight(line, " \t")
			lines = append(lines, line)

			if p.ch < 0 {
				break
			} else if p.ch == '\n' {
				p.next()
				p.parseLead()
				p.parseSpacing()

				if _, matchesBlock := p.matchBlock(); p.ch < 0 || p.ch == '\n' || matchesBlock || !p.continues(p.blocks) {
					cont = false
					break
				}

				offs = p.offset
			} else if p.closingDelimiter() >= 0 || matchesInline {
				end = p.pos()
				break
			} else {
				panic("unexpected case")
			}
		} else {
			if isEscape {
				escapes = append(escapes, p.offset)
				p.next()
			}
			p.next()
			end = p.pos()
		}
	}

	txt := bytes.Join(lines, []byte("\n"))
	if len(txt) == 0 {
		if trace {
			p.print("return nil")
		}
		return nil, cont
	}
	if trace {
		defer p.printf("return %q", txt)
	}

	n := &node.Node{
		Element: name,
		Type:    node.TypeText,
		Value:   string(txt),
		Location: node.Location{
			Range: node.Range{
				Start: start,
				End:   end,
			},
		},
	}
	return n, cont
}

func (p *parser) matchInline() (Element, bool) {
	if trace {
		defer p.trace("matchInline")()
	}

	for d, e := range p.inlineMap {
		if p.hasPrefix([]byte(d)) {
			if trace {
				p.printf("return true (%s)", e.Name)
			}

			return e, true
		}
	}

	if trace {
		p.print("return false")
	}

	return Element{}, false
}

// closingDelimiter returns the closing delimiter if found, otherwise 0.
func (p *parser) closingDelimiter() rune {
	if trace {
		defer p.trace("closingDelimiter")()
	}

	for i := len(p.inlines) - 1; i >= 0; i-- {
		delim := p.inlines[i]
		c := counterpart(delim)

		if p.ch == c && p.peek() == c {
			if trace {
				p.printf("return %q (is closing delim)", c)
			}

			return c
		}
	}

	if trace {
		p.print("return -1 (not closing delim)")
	}

	return -1
}

func (p *parser) init(src io.Reader) {
	b, err := io.ReadAll(src)
	if err != nil {
		panic(fmt.Errorf("parser: ReadAll failed: %s", err))
	}

	p.src = b

	p.next()
	if p.ch == bom {
		// skip BOM at file beginning
		p.next()
	}
}

// parseLead parses block delimiters at the start of the line.
//
// Call at the start of the line as it consumes required block delimiters.
func (p *parser) parseLead() {
	if trace {
		defer p.trace("parseLead")()
		p.printDelims("reqd", p.blocks)
	}

	p.lead = nil
	p.blank = true

	a := p.expandTabs(p.blocks)

	var lead []rune
	var i int

	for p.ch >= 0 && p.ch != '\n' && i < len(a) {
		if !isSpacing(p.ch) {
			p.blank = false
		}

		ch := a[i]

		if isSpacing(p.ch) && isSpacing(ch) {
			if p.countSpacing([]rune{p.ch}) <= p.countSpacing(spacingSeq(a[i:])) {
				i += p.countSpacing([]rune{p.ch})
			} else if p.countSpacing([]rune{p.ch}) > p.countSpacing(spacingSeq(a[i:])) {
				i += p.countSpacing([]rune{ch})
			} else {
				break
			}
		} else if isSpacing(p.ch) {
			p.parseSpacing() // calls p.next()
			continue
		} else if p.ch == ch {
			i++
		} else {
			break
		}

		lead = append(lead, p.ch)

		p.next()
	}

	p.addLead(lead...)

	if trace {
		p.printDelims("new", lead)
		p.printDelims("lead", p.lead)
	}
}

func stripSpacing(a []rune) []rune {
	var b []rune
	for _, c := range a {
		if c == ' ' || c == '\t' {
			continue
		}
		b = append(b, c)
	}
	return b
}

// parseSpacing is like parseLead but only parses spacing and can be used in the
// middle of the line.
func (p *parser) parseSpacing() {
	if trace {
		defer p.trace("parseSpacing")()
	}

	var lead []rune
	for isSpacing(p.ch) {
		lead = append(lead, p.ch)

		p.next()
	}

	p.addLead(lead...)

	if trace {
		p.printDelims("new", lead)
		p.printDelims("lead", p.lead)
	}
}

func isSpacing(r rune) bool {
	return r == ' ' || r == '\t'
}

// Encoding errors
var (
	ErrInvalidUTF8Encoding = &Error{"invalid UTF-8 encoding"}
	ErrIllegalNULL         = &Error{"illegal character NULL"}
	ErrIllegalBOM          = &Error{"illegal byte order mark"}
)

const (
	bom = 0xFEFF // byte order mark (permitted as first character)
	eof = -1     // end of file
)

// next reads the next character.
// https://cs.opensource.google/go/go/+/refs/tags/go1.17.3:src/go/scanner/scanner.go;l=61
func (p *parser) next() {
	if trace {
		defer p.trace("next")()
	}

	if p.rdOffset < len(p.src) {
		p.offset = p.rdOffset
		if p.ch == '\n' {
			p.line++
			p.lineOffset = p.offset
		}

		r, w := rune(p.src[p.rdOffset]), 1
		switch {
		case r == 0:
			p.error(ErrIllegalNULL)
		case r >= utf8.RuneSelf:
			// not ASCII
			r, w = utf8.DecodeRune(p.src[p.rdOffset:])
			if r == utf8.RuneError && w == 1 {
				p.error(ErrInvalidUTF8Encoding)
			} else if r == bom && p.offset > 0 {
				// BOM at offset 0 is skipped at init
				p.error(ErrIllegalBOM)
			}
		}
		p.rdOffset += w
		p.ch = r
	} else {
		p.offset = len(p.src)
		if p.ch == '\n' {
			p.line++
			p.lineOffset = p.offset
		}
		p.ch = eof
	}

	if trace {
		p.printf("p.ch=%q ", p.ch)
	}
}

func (p *parser) peek() rune {
	if p.rdOffset < len(p.src) {
		r := rune(p.src[p.rdOffset])
		if r >= utf8.RuneSelf {
			// not ASCII
			r, _ = utf8.DecodeRune(p.src[p.rdOffset:])
		}
		return r
	}
	return eof
}

func (p *parser) peek2() rune {
	if p.rdOffset < len(p.src) {
		r, w := rune(p.src[p.rdOffset]), 1
		if r >= utf8.RuneSelf {
			// not ASCII
			r, w = utf8.DecodeRune(p.src[p.rdOffset:])
		}
		if o := p.rdOffset + w; o < len(p.src) {
			rr, _ := rune(p.src[o]), 1
			if rr >= utf8.RuneSelf {
				// not ASCII
				rr, _ = utf8.DecodeRune(p.src[o:])
			}
			return rr
		}
	}
	return eof
}

func (p *parser) open(blocks ...rune) func() {
	size := len(p.blocks)
	p.blocks = append(p.blocks, blocks...)

	return func() {
		p.blocks = p.blocks[:size]
	}
}

func (p *parser) addLead(blocks ...rune) {
	p.lead = append(p.lead, blocks...)
}

func (p *parser) diff(old, new []rune) []rune {
	if len(old) == 0 {
		return new
	}

	n := p.expandTabs(new)

	a := trailingSpacing(old)
	if len(a) < len(old) {
		b := trailingSpacing(n)
		if len(a) > 0 && len(b) > 0 {
			x := p.countSpacing(a)
			y := p.countSpacing(b)
			if y > x {
				// different trailing spacing
				z := y - x
				return n[len(n)-z:]
			}
		}
	}

	var i int
	for i = len(n) - 1; i >= 0; i-- {
		if i < len(old) {
			break
		}

		c := n[i]
		if c != ' ' && c != '\t' && c == old[len(old)-1] {
			break
		}
	}

	return n[i+1:]
}

func trailingSpacing(a []rune) []rune {
	var i int
	for i = len(a) - 1; i >= 0; i-- {
		c := a[i]
		if c != ' ' && c != '\t' {
			break
		}
	}
	return a[i+1:]
}

func (p *parser) expandTabs(a []rune) []rune {
	n := make([]rune, 0, len(a))
	for _, c := range a {
		if c == '\t' {
			for i := 0; i < p.tabWidth; i++ {
				n = append(n, ' ')
			}
		} else {
			n = append(n, c)
		}
	}
	return n
}

func (p *parser) openInline(delim rune) func() {
	size := len(p.inlines)
	p.inlines = append(p.inlines, delim)

	return func() {
		p.inlines = p.inlines[:size]
	}
}

func (p *parser) error(err *Error) {
	p.errors.Add(err)
}

func (p *parser) printDelims(name string, blocks []rune) {
	p.print(name + "=" + fmtBlocks(blocks))
}

func fmtBlocks(blocks []rune) string {
	var b strings.Builder
	b.WriteString("[")

	for i := 0; i < len(blocks); i++ {
		if i > 0 {
			b.WriteString(", ")
		}

		j := 1
		for i+j < len(blocks) && blocks[i+j-1] == blocks[i+j] {
			j++
		}

		if j > 2 {
			b.WriteString(fmt.Sprintf("%dx", j))
			i += j - 1
		}

		b.WriteString(fmt.Sprintf("%q", blocks[i]))
	}

	b.WriteString("]")
	return b.String()
}

func (p *parser) tracef(format string, v ...interface{}) func() {
	return p.trace(fmt.Sprintf(format, v...))
}

func (p *parser) trace(msg string) func() {
	p.printf("%q -> %s (", p.ch, msg)
	p.indent++

	return func() {
		p.indent--
		p.print(")")
	}
}

func (p *parser) printf(format string, v ...interface{}) {
	p.print(fmt.Sprintf(format, v...))
}

func (p *parser) print(msg string) {
	fmt.Println(strings.Repeat("\t", p.indent) + msg)
}
