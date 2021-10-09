package parser

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/touchmarine/to/matcher"
	"github.com/touchmarine/to/node"
)

const trace = false

const tabWidth = 8

var tabSpaces []rune

func init() {
	tabSpaces = make([]rune, tabWidth)
	for i := 0; i < tabWidth; i++ {
		tabSpaces[i] = ' '
	}
}

type ElementMap map[string]Element

type Element struct {
	Name      string
	Type      node.Type
	Delimiter string
	Matcher   string
}

func Parse(src io.Reader, elementMap ElementMap) (*node.Node, []error) {
	var p parser
	p.matchers(matcher.Defaults())
	p.elements(elementMap)
	p.init(src)
	return p.parse(nil), p.errors
}

// parser holds the parsing state.
type parser struct {
	errors         []error
	src            []byte             // source
	blockMap       map[string]Element // registered elements by delimiter
	blockKeys      []string           // length-sorted (longest first) blockMap keys
	leaf           string             // leaf element name
	inlineMap      map[string]Element // registered inline elements by delimiter
	text           string             // text element name
	specialEscapes []string           // delimiters that do not start with a punctuation
	matcherMap     matcher.Map        // registered matchers by name

	// parsing
	ch       rune // current character
	offset   int  // character offset
	rdOffset int  // position after current character

	blocks []rune // open blocks
	lead   []rune // blocks on current line
	blank  bool   // whether the lead is blank

	inlines []rune // open inlines

	// tracing
	indent int // trace indentation
}

func (p *parser) elements(elementMap ElementMap) {
	if p.blockMap == nil {
		p.blockMap = make(map[string]Element)
	}
	if p.inlineMap == nil {
		p.inlineMap = make(map[string]Element)
	}

	for _, e := range elementMap {
		switch c := node.TypeCategory(e.Type); c {
		case node.CategoryBlock:
			switch e.Type {
			case node.TypeLeaf:
				p.leaf = e.Name
			case node.TypeRankedHanging:
				p.blockMap[e.Delimiter+e.Delimiter] = e
			default:
				p.blockMap[e.Delimiter] = e
			}

		case node.CategoryInline:
			switch e.Type {
			case node.TypeText:
				p.text = e.Name
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
					panic("parser: invalid inline delimiter " + e.Delimiter)
				}

				p.blockMap[delimiter] = e
				p.inlineMap[delimiter] = e
			}

		default:
			panic("parser: unexpected node category " + c.String())
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

func (p *parser) matchers(m matcher.Map) {
	if p.matcherMap == nil {
		p.matcherMap = make(matcher.Map)
	}

	for k, v := range m {
		p.matcherMap[k] = v
	}
}

func (p *parser) parse(reqdBlocks []rune) *node.Node {
	if trace {
		defer p.trace("parse")()
	}

	container := &node.Node{
		Type: node.TypeContainer,
	}

	for p.ch > 0 {
		if isSpacing(p.ch) {
			p.parseSpacing()
		} else if !p.continues(reqdBlocks) {
			break
		} else if p.ch == '\n' {
			p.next()
			p.parseLead()
		} else {
			b := p.parseBlock()
			if b == nil {
				panic("parser: parseBlock() returned no block")
			}

			container.AppendChild(b)
		}
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

			if node.TypeCategory(e.Type) == node.CategoryBlock {
				if trace {
					p.printf("return true (%s)", e.Name)
				}

				return e, true
			} else if node.TypeCategory(e.Type) == node.CategoryInline {
				if trace {
					p.print("return false, inline")
				}

				return Element{}, false
			} else {
				panic("unexpected element type category " + node.TypeCategory(e.Type).String())
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

	p.addLead(p.ch)
	defer p.open(p.ch)()

	p.next() // consume delimiter

	reqdBlocks := p.blocks
	children := p.parse(reqdBlocks)

	n := &node.Node{
		Element: name,
		Type:    node.TypeWalled,
	}
	n.AppendChild(children)
	return n
}

func (p *parser) parseVerbatimWalled(name string) *node.Node {
	if trace {
		defer p.tracef("parseVerbatimWalled (%s)", name)()
	}

	p.addLead(p.ch)
	defer p.open(p.ch)()

	p.next() // consume delimiter

	firstLine := p.consumeLine() // consume here so we can have a nicer loop

	lines := [][]byte{firstLine}

	for p.ch > 0 {
		if p.ch == '\n' {
			p.next()
			p.parseLead()
		}

		if !p.continues(p.blocks) {
			break
		}

		line := p.consumeLine()
		lines = append(lines, line)
	}

	return &node.Node{
		Element: name,
		Type:    node.TypeVerbatimWalled,
		Value:   strings.Join(btosSlice(lines), "\n"),
	}
}

// btosSlice converts [][]byte to []string.
func btosSlice(p [][]byte) []string {
	var lines []string
	for _, line := range p {
		lines = append(lines, string(line))
	}
	return lines
}

func (p *parser) parseHanging(name, delim string) *node.Node {
	if trace {
		defer p.tracef("parseHanging (%s, delim=%q)", name, delim)()
	}

	c := utf8.RuneCountInString(delim)
	p.addLead([]rune(strings.Repeat(" ", c))...)

	// consume delimiter
	for i := 0; i < c; i++ {
		p.next()
	}

	children := p.parseHanging0()
	n := &node.Node{
		Element: name,
		Type:    node.TypeHanging,
	}
	n.AppendChild(children)
	return n
}

func (p *parser) parseRankedHanging(name, delim string) *node.Node {
	if trace {
		defer p.tracef("parseRankedHanging (%s, delim=%q)", name, delim)()
	}

	var rank int

	// consume delimiter, count rank
	d := p.ch
	for p.ch == d {
		rank++

		p.next()
	}

	p.addLead([]rune(strings.Repeat(" ", rank))...)

	children := p.parseHanging0()
	n := &node.Node{
		Element: name,
		Type:    node.TypeRankedHanging,
		Data:    rank,
	}
	n.AppendChild(children)
	return n
}

func (p *parser) parseHanging0() *node.Node {
	if trace {
		defer p.trace("parseHanging0")()
	}

	newBlocks := diff(p.blocks, p.lead)
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
	if trace {
		defer p.trace("continues")()
		p.printDelims("reqd", blocks)
		p.printDelims("lead", p.lead)
	}

	if p.blank && onlySpacing(blocks) {
		if trace {
			p.print("return true (blank)")
		}
		return true
	}

	var i, j int
	for {
		if i > len(blocks)-1 {
			if trace {
				p.print("return true")
			}
			return true
		}

		if j > len(p.lead)-1 {
			if onlySpacing(blocks[i:]) && len(p.lead) > 0 &&
				(p.ch == 0 || p.ch == '\n' || p.ch == ' ' || p.ch == '\t') {
				if trace {
					p.print("return true")
				}
				return true
			}

			if trace {
				p.print("return false (not enough blocks)")
			}
			return false
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

			x := countSpacing(spacingSeq(blocks[n:i]))
			y := countSpacing(spacingSeq(p.lead[m:j]))

			if y < x {
				if trace {
					p.print("return false (lesser ident)")
				}
				return false
			}

			continue
		}

		if blocks[i] != p.lead[j] {
			if trace {
				p.printf("return false (%q != %q, i=%d, j=%d)", blocks[i], p.lead[j], i, j)
			}
			return false
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

	reqdBlocks := p.blocks

	openSpacing := diffSpacing(lastSpacingSeq(p.blocks), lastSpacingSeq(p.lead))
	defer p.open(openSpacing...)()

	delim := p.ch

	// consume delimiter
	p.next()

	escaped := p.ch == '\\'
	if escaped {
		p.next()
	}

	openingText := p.consumeLine()

	lines := [][]byte{openingText}
	afterNewline := false

	for p.ch > 0 && p.continues(reqdBlocks) {
		if !escaped && p.ch == delim || escaped && p.ch == '\\' && p.peek() == delim {
			// closing delimiter

			if escaped {
				p.next()
			}
			p.next()

			break
		}

		if p.ch == '\n' && !afterNewline {
			p.next()
			p.parseLead()

			afterNewline = true
		} else {
			// leading spacing that is part of the element
			spacing := diffSpacing(lastSpacingSeq(p.blocks), lastSpacingSeq(p.lead))
			l := p.consumeLine()

			line := string(spacing) + string(l)
			lines = append(lines, []byte(line))

			afterNewline = false
		}
	}

	if p.continues(reqdBlocks) {
		// closed by delimiter, not continues

		p.consumeLine()

		p.next()
		p.parseLead()
		p.parseSpacing()
	}

	return &node.Node{
		Element: name,
		Type:    node.TypeFenced,
		Value:   strings.Join(btosSlice(lines), "\n"),
	}
}

func (p *parser) consumeLine() []byte {
	if trace {
		defer p.trace("consumeLine")()
	}

	var b bytes.Buffer

	for p.ch > 0 && p.ch != '\n' {
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
func diffSpacing(a, b []rune) []rune {
	x := countSpacing(a)
	y := countSpacing(b)

	if y == x {
		return nil
	} else if y > x {
		var c []rune

		n := y - x
		for i := len(b) - 1; i >= 0; i-- {
			if n <= 0 {
				break
			}

			w := countSpacing([]rune{b[i]})
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

func countSpacing(s []rune) int {
	var i int
	for _, ch := range s {
		switch ch {
		case ' ':
			i++
		case '\t':
			i += tabWidth
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

	for i := 0; i < utf8.RuneCountInString(delim); i++ {
		p.next()

	}

	content := p.consumeLine()

	p.next()
	p.parseLead()
	p.parseSpacing()

	return &node.Node{
		Element: name,
		Type:    node.TypeVerbatimLine,
		Value:   string(content),
	}
}

func (p *parser) parseLeaf(name string) *node.Node {
	if trace {
		defer p.tracef("parseLeaf %s", name)()
	}

	children, _ := p.parseInlines()
	n := &node.Node{
		Element: name,
		Type:    node.TypeLeaf,
	}
	n.AppendChild(children)
	return n
}

func (p *parser) parseInlines() (*node.Node, bool) {
	if trace {
		defer p.trace("parseInlines")()
	}

	container := &node.Node{
		Type: node.TypeContainer,
	}

	for p.ch > 0 {
		if p.closingDelimiter() > 0 {
			break
		}

		inline, cont := p.parseInline()
		if inline != nil {
			container.AppendChild(inline)
		}

		if !cont {
			return container, false
		}
	}

	return container, true
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

	return p.parseText(p.text)
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

	n := &node.Node{
		Element: name,
		Type:    node.TypeUniform,
	}
	n.AppendChild(children)
	return n, cont
}

func (p *parser) parseEscaped(name string) (*node.Node, bool) {
	if trace {
		defer p.tracef("parseEscaped (%s)", name)()
	}

	delim := p.ch
	c := counterpart(delim)
	closing := []rune{c, c}

	// consume delimiter
	p.next()
	p.next()

	escaped := p.ch == '\\'
	if escaped {
		closing = append([]rune{'\\'}, closing...)
		p.next()
	}

	cont := true
	afterNewline := false

	var b bytes.Buffer
	for p.ch > 0 {
		if p.ch == '\n' {
			line := b.Bytes()
			trailingSpacing := len(line) - len(bytes.TrimRight(line, " \t"))

			if trailingSpacing > 0 {
				// remove trailing spacing
				b.Truncate(len(line) - trailingSpacing)
			}

			if afterNewline {
				cont = false
				break
			}

			p.next()
			p.parseLead()
			p.parseSpacing()

			afterNewline = true

			_, matchesBlock := p.matchBlock()
			if matchesBlock || !p.continues(p.blocks) {
				cont = false
				break
			}

			continue
		}

		var a []rune
		if escaped {
			a = append([]rune{p.ch}, []rune{p.peek(), p.peek2()}...)
		} else {
			a = []rune{p.ch, p.peek()}
		}

		if cmpRunes(a, closing) {
			// closing delimiter

			for i := 0; i < len(closing); i++ {
				p.next()

			}

			break
		}

		if afterNewline {
			b.WriteByte(' ') // newline separator

			afterNewline = false
		}

		b.WriteRune(p.ch)

		p.next()
	}

	txt := b.Bytes()

	if trace {
		defer p.printf("return %q", txt)
	}

	return &node.Node{
		Element: name,
		Type:    node.TypeEscaped,
		Value:   string(txt),
	}, cont
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

	// consume prefix
	for i := 0; i < len(prefix); i++ {
		p.next()
	}

	if matcher == "" {
		return &node.Node{
			Element: name,
			Type:    node.TypePrefixed,
		}, true
	}

	m, ok := p.matcherMap[matcher]
	if !ok {
		panic("parser: matcher " + matcher + " not found")
	}

	w := m.Match(p.src[p.offset:])
	end := p.offset + w

	// consume match
	var b bytes.Buffer
	for p.offset < end {
		b.WriteRune(p.ch)
		p.next()
	}

	return &node.Node{
		Element: name,
		Type:    node.TypePrefixed,
		Value:   string(b.Bytes()),
	}, true
}

func (p *parser) parseText(name string) (*node.Node, bool) {
	if trace {
		defer p.tracef("parseText (%s)", name)()
	}

	cont := true
	afterNewline := false

	var b bytes.Buffer
	for p.ch > 0 {
		if p.ch == '\n' {
			line := b.Bytes()
			trailingSpacing := len(line) - len(bytes.TrimRight(line, " \t"))

			if trailingSpacing > 0 {
				// remove trailing spacing
				b.Truncate(len(line) - trailingSpacing)
			}

			if afterNewline {
				cont = false
				break
			}

			p.next()
			p.parseLead()
			p.parseSpacing()

			afterNewline = true

			if !p.isEscape() {
				_, matchesBlock := p.matchBlock()
				if matchesBlock || !p.continues(p.blocks) {
					cont = false
					break
				}
			}
		} else {
			if afterNewline {
				b.WriteByte(' ') // newline separator

				afterNewline = false
			}

			if p.isEscape() {
				p.next()
			} else {
				if p.closingDelimiter() > 0 {
					break
				}

				if _, ok := p.matchInline(); ok {
					break
				}
			}

			b.WriteRune(p.ch)

			p.next()
		}
	}

	if b.Len() == 0 {
		if trace {
			p.print("return nil")
		}

		return nil, cont
	}

	txt := b.Bytes()

	if trace {
		defer p.printf("return %q", txt)
	}

	return &node.Node{
		Element: name,
		Type:    node.TypeText,
		Value:   string(txt),
	}, cont
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
		p.print("return 0 (not closing delim)")
	}

	return 0
}

func (p *parser) init(src io.Reader) {
	b, err := io.ReadAll(src)
	if err != nil {
		panic(fmt.Errorf("parser: ReadAll failed: %s", err))
	}

	p.src = b

	p.next()
	if p.ch == '\uFEFF' {
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

	a := expandTabs(p.blocks)

	var lead []rune
	var i int

	for p.ch > 0 && p.ch != '\n' && i < len(a) {
		if !isSpacing(p.ch) {
			p.blank = false
		}

		ch := a[i]

		if isSpacing(p.ch) && isSpacing(ch) {
			if countSpacing([]rune{p.ch}) <= countSpacing(spacingSeq(a[i:])) {
				i += countSpacing([]rune{p.ch})
			} else if countSpacing([]rune{p.ch}) > countSpacing(spacingSeq(a[i:])) {
				i += countSpacing([]rune{ch})
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
	ErrInvalidUTF8Encoding = errors.New("invalid UTF-8 encoding")
	ErrIllegalNULL         = errors.New("illegal character NULL")
	ErrIllegalBOM          = errors.New("illegal byte order mark")
)

// next reads the next character.
func (p *parser) next() {
	if trace {
		defer p.trace("next")()
	}

	if p.rdOffset < len(p.src) {
		p.offset = p.rdOffset

		r, w := utf8.DecodeRune(p.src[p.rdOffset:])

		switch r {
		case utf8.RuneError: // encoding error
			if w == 0 {
				panic("parser: cannot decode empty slice")
			} else if w == 1 {
				p.error(ErrInvalidUTF8Encoding)
				p.ch = utf8.RuneError
			} else {
				p.ch = r
			}
		case '\u0000': // NULL
			p.error(ErrIllegalNULL)
			p.ch = utf8.RuneError
		case '\uFEFF': // BOM
			if p.offset == 0 {
				// skip in p.init
				p.ch = r
			} else {
				p.error(ErrIllegalBOM)
				p.ch = utf8.RuneError
			}
		default:
			p.ch = r
		}

		p.rdOffset += w
	} else {
		p.ch = 0
		p.offset = len(p.src)
	}

	if trace {
		p.printf("p.ch=%q ", p.ch)
	}
}

func (p *parser) peek() rune {
	if p.rdOffset < len(p.src) {
		return validRune(utf8.DecodeRune(p.src[p.rdOffset:]))
	}
	return 0
}

func (p *parser) peek2() rune {
	if p.peek() > 0 {
		l := utf8.RuneLen(p.peek())

		if p.rdOffset+l < len(p.src) {
			return validRune(utf8.DecodeRune(p.src[p.rdOffset+l:]))
		}
	}

	return 0
}

func validRune(r rune, w int) rune {
	switch r {
	case utf8.RuneError:
		if w == 0 {
			panic("parser: cannot decode empty slice")
		} else {
			return utf8.RuneError
		}
	case '\u0000', '\uFEFF': // NULL, or BOM
		return utf8.RuneError
	}
	return r
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

func diff(old, new []rune) []rune {
	if len(old) == 0 {
		return new
	}

	n := expandTabs(new)

	a := trailingSpacing(old)
	if len(a) < len(old) {
		b := trailingSpacing(n)
		if len(a) > 0 && len(b) > 0 {
			x := countSpacing(a)
			y := countSpacing(b)
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

func expandTabs(a []rune) []rune {
	n := make([]rune, 0, len(a))
	for _, c := range a {
		if c == '\t' {
			n = append(n, tabSpaces...)
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

func (p *parser) error(err error) {
	p.errors = append(p.errors, err)
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
