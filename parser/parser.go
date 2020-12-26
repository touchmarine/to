package parser

import (
	"fmt"
	"strings"
	"to/node"
	"to/printer"
)

const trace = false

// tabWidth in spaces; used for determining list item identation
const tabWidth = 4

type headingTyp uint

// heading types
const (
	unnumberedHeading headingTyp = iota
	numberedHeading
)

//go:generate stringer -type=linkVariant
type linkVariant uint

// link variants
const (
	onePartLink linkVariant = iota
	twoPartLink
)

type Parser struct {
	// immutable state
	src string

	// scanning state
	ch       byte // current character
	offset   int  // character offset
	rdOffset int  // reading offset (position after current character)

	indent int // trace indentation level
}

func New(src string) *Parser {
	p := &Parser{src: src}
	// initialize ch, offset, and rdOffset
	p.next()
	return p
}

// next reads the next character into p.ch.
// p.ch < 0 means end-of-file.
func (p *Parser) next() {
	if p.rdOffset < len(p.src) {
		p.ch = p.src[p.rdOffset]
	} else {
		p.ch = 0 // eof
	}
	p.offset = p.rdOffset
	p.rdOffset += 1
}

// peek returns the byte following the most recently read character without
// advancing the parser. If the parser is at EOF, peek returns 0.
func (p *Parser) peek() byte {
	if p.rdOffset < len(p.src) {
		return p.src[p.rdOffset]
	}
	return 0
}

// reset returns the pointers to the offs position.
func (p *Parser) reset(offs int) {
	p.ch = 0 // gets overridden
	p.offset = offs - 1
	p.rdOffset = offs
	p.next() // set ch
}

func (p *Parser) ParseDocument() *node.Document {
	if trace {
		defer p.trace("ParseDocument")()
	}

	doc := &node.Document{}

	for p.ch != 0 {
		block := p.parseBlock()
		if block == nil {
			break
		}

		doc.Children = append(doc.Children, block)
		// pointers are advaced by p.parseBlock()
	}

	return doc
}

// eatIndent consumes consecutive tabs and spaces and counts them.
// skipBlankLines skips a blank line and resets count. It returns the identation
// of the first non-blank line if enabled.
// tab width = tabWidth spaces
func (p *Parser) eatIndent(skipBlankLines bool) int {
	var indent int
	for {
		if skipBlankLines && p.ch == '\n' {
			indent = 0
			p.next()
			continue
		}

		switch p.ch {
		case '\t':
			indent += tabWidth
		case ' ':
			indent++
		default:
			return indent
		}

		p.next()
	}
}

func (p *Parser) parseBlock() node.Node {
	if trace {
		defer p.trace("parseBlock")()
	}

	// count first non-blank line indent
	indent := p.eatIndent(true)

	switch p.ch {
	case 0: // EOF
		return nil
	case '=':
		return p.parseHeading(unnumberedHeading)
	default:
		// test if list; parseList returns nil if not a list without the
		// advancing pointers
		if list, _ := p.parseList(indent); list != nil {
			return list
		}

		switch {
		case p.ch == '#' && p.peek() == '#':
			return p.parseHeading(numberedHeading)
		case p.ch == '`' && p.peek() == '`':
			return p.parseCodeBlock()
		}

		return p.parseParagraph()
	}
}

func (p *Parser) parseHeading(typ headingTyp) *node.Heading {
	var isNumbered bool
	var delim byte

	// determine heading type we are parsing
	switch typ {
	case unnumberedHeading:
		isNumbered = false
		delim = '='
	case numberedHeading:
		isNumbered = true
		delim = '#'
	default:
		panic("unsupported heading type")
	}

	if trace {
		if isNumbered {
			defer p.trace("parseNumberedHeading")()
		} else {
			defer p.trace("parseHeading")()
		}
	}

	// count heading level by counting consecutive delimiters
	level := 0
	for p.ch == delim {
		level++
		p.next()
	}

	// skip whitespace
	for p.ch == '\t' || p.ch == ' ' {
		p.next()
	}

	h := &node.Heading{
		Level:      level,
		IsNumbered: isNumbered,
		Children:   p.parseInline(delimiters{}, 0),
	}
	// pointers are advanced by p.parseInline()

	return h
}

func (p *Parser) parseCodeBlock() *node.CodeBlock {
	if trace {
		defer p.trace("parseCodeBlock")()
	}

	// count opening delimiters
	var openingDelims int
	for p.ch == '`' {
		openingDelims++
		p.next()
	}

	// parse metadata
	metadataOffs := p.offset
	for p.ch != '\n' && p.ch != 0 {
		p.next()
	}

	metadata := p.src[metadataOffs:p.offset]
	p.next() // eat EOL or EOF

	// parse body
	offs := p.offset
	var closingDelims int // we need this outside to offset closing delims
	for p.ch != 0 {
		// count consecutive backticks which may be closing delimiter
		for p.ch == '`' {
			closingDelims++
			p.next()
		}

		// test for possible closing delimiter
		// needs >= number of backticks as the opening delimiter
		if closingDelims >= openingDelims {
			break
		}
		closingDelims = 0 // reset counter if not closing delimiter

		p.next()
	}

	var body string
	if endOffs := p.offset - closingDelims; endOffs < len(p.src) {
		body = p.src[offs:endOffs]
	}

	// parse metadata language and filename
	var language string
	var filename string
	s := strings.SplitN(metadata, ",", 2)

	// strings.TrimSpace() removes Unicode whitespace so it currently does not
	// match our other whitespace removal...
	if len(s) >= 1 {
		language = strings.TrimSpace(s[0])
	}
	if len(s) >= 2 {
		filename = strings.TrimSpace(s[1])
	}

	cb := &node.CodeBlock{
		Language:    language,
		Filename:    filename,
		MetadataRaw: metadata,
		Body:        body,
	}

	if trace {
		p.print("return\n" + printer.Pretty(cb, p.indent+1))
	}

	return cb
}

// parseList returns a list and the ending indentation it consumed.
// It returns nil if no list present.
func (p *Parser) parseList(indent int) (*node.List, int) {
	// do not trace as it might not even be a list

	switch p.ch {
	case '-':
		return p.parseUnorderedList(indent)
	default:
		return nil, 0
	}
}

// parseUnoredredList parses until a line that is indented less than or equal
// to the opening line. List items on equal indentation are part of the list.
// It returns an unordered list and the ending indentation it consumed.
func (p *Parser) parseUnorderedList(indent int) (*node.List, int) {
	if trace {
		defer p.trace(fmt.Sprintf("parseUnorderedList(%d)", indent))()
	}

	var listItems []*node.ListItem

	var endIndent int
	for p.ch == '-' && p.ch != 0 {
		p.next() // eat opening '-'

		var li *node.ListItem
		li, endIndent = p.parseListItem(indent)
		listItems = append(listItems, li)

		// end list if indentation is less than starting indentation
		if endIndent < indent {
			break
		}
	}

	return &node.List{
		Type:      node.UnorderedList,
		ListItems: listItems,
	}, endIndent
}

// parseListItem parses until a line that is indented less than or equal to the
// opening line. It returns a list item and the ending indentation it consumed.
func (p *Parser) parseListItem(indent int) (*node.ListItem, int) {
	if trace {
		defer p.trace(fmt.Sprintf("parseListItem(%d)", indent))()
	}

	// we group consecutive lines into line groups
	// when a non-line node is encountered we flush current lines to children
	// and reset lines
	// grouped lines allow for easier processing
	var lines node.Lines

	// parse opening line
	line := p.parseLine()
	if line != nil {
		lines = append(lines, line)
	}
	p.next() // eat EOL

	var children []node.Node

	var endIndent int // indent after the list; we parse it to determine if still a list
	for p.ch != 0 {
		curIndent := p.eatIndent(false)

		// stop parsing if indentation not greater than starting
		if curIndent <= indent {
			// nested list could already consumed the indentation - which we set
			// below
			if curIndent > endIndent {
				endIndent = curIndent
			}
			break
		}

		// parse nested list if present; parseList returns nil if not a list
		if list, lEndIndent := p.parseList(curIndent); list != nil {
			// append lines group and reset lines if not empty
			if len(lines) > 0 {
				children = append(children, lines)
				lines = nil
			}

			children = append(children, list)
			endIndent = lEndIndent
			continue
		}

		// parse line
		line := p.parseLine()
		if line != nil {
			lines = append(lines, line)
		}
		p.next() // eat EOL

	}

	// append lines group if not empty
	if len(lines) > 0 {
		children = append(children, lines)
	}

	listItem := &node.ListItem{
		Children: children,
	}

	if trace {
		p.print("return\n" + printer.Pretty(listItem, p.indent))
	}

	return listItem, endIndent
}

// parseParagraph parses consecutive lines of inline text until another block,
// EOL, or EOF.
func (p *Parser) parseParagraph() *node.Paragraph {
	if trace {
		defer p.trace("parseParagraph")()
	}

	var lines node.Lines

	for {
		line := p.parseLine()
		if line == nil {
			break
		}

		lines = append(lines, line)
		p.next() // eat line EOL
	}

	return &node.Paragraph{
		Lines: lines,
	}
}

// parseLine parses a line of inline elements. It returns nil if line has no
// inline elements, i.e., if line starts a block or is blank.
func (p *Parser) parseLine() *node.Line {
	if trace {
		defer p.trace("parseLine")()
	}

	var children []node.Inline

	for p.ch != '\n' && p.ch != 0 {
		// skip leading whitespace
		for p.ch == '\t' || p.ch == ' ' {
			p.next()
		}

		// end if single delim block
		if contains(blockDelims.single, p.ch) {
			break
		}

		// end if double or more delims block
		if p.ch == p.peek() && contains(blockDelims.double, p.ch) &&
			contains(blockDelims.double, p.peek()) {
			break
		}

		if trace {
			p.print(fmt.Sprintf("p.ch=%s, p.peek()=%s", char(p.ch), char(p.peek())))
		}

		children = append(children, p.parseInline(delimiters{}, 0)...)
	}

	if len(children) == 0 {
		return nil
	}

	return &node.Line{
		Children: children,
	}
}

type delimiters struct {
	single []byte // <https://koala.test>
	double []byte // **strong**
}

var inlineDelims = delimiters{
	single: []byte{'<'},
	double: []byte{'_', '*'},
}

var blockDelims = delimiters{
	single: []byte{'=', '-'},
	double: []byte{'#', '`'},
}

// parseInline parses until one of the provided delims, EOL, or EOF.
// exclude excludes the provided character as a delimiter if not 0.
func (p *Parser) parseInline(delims delimiters, exclude byte) []node.Inline {
	if trace {
		defer p.trace("parseInline")()
		p.print(fmt.Sprintf(
			"single delims=%s, double delims=%s, exclude=%s",
			delims.single,
			delims.double,
			string(exclude),
		))
	}

	var inlines []node.Inline
	for p.ch != '\n' && p.ch != 0 {
		// end if delimiter but not if excluded, non-zero delimiter
		if exclude == 0 || exclude != 0 && p.ch != exclude {
			if contains(delims.single, p.ch) {
				break
			}

			if p.ch == p.peek() && contains(delims.double, p.ch) &&
				contains(delims.double, p.peek()) {
				break
			}
		}

		if trace {
			p.print(fmt.Sprintf("p.ch=%s, p.peek()=%s", char(p.ch), char(p.peek())))
		}

		switch {
		case exclude != 0 && p.ch == exclude:
			inlines = append(inlines, p.parseText(delims, exclude))
		case p.ch == '_' && p.peek() == '_':
			inlines = append(inlines, p.parseEmphasis(delims))
		case p.ch == '*' && p.peek() == '*':
			inlines = append(inlines, p.parseStrong(delims))
		case p.ch == '<':
			inlines = append(inlines, p.parseLink(delims))
		default:
			inlines = append(inlines, p.parseText(delims, exclude))
		}

		// pointers are advanced by parslets
	}

	if trace {
		p.print("return " + printer.Pretty(node.InlinesToNodes(inlines), p.indent))
	}

	return inlines
}

// parseEmphasis parses until a '__' and consumes the closing delimiter.
func (p *Parser) parseEmphasis(delims delimiters) *node.Emphasis {
	if trace {
		defer p.trace("parseEmphasis")()
	}

	// eat opening '__'
	p.next()
	p.next()

	// no possible duplicates because p.parseInline() returns on delim match
	delims.double = append(delims.double, '_')

	em := &node.Emphasis{
		Children: p.parseInline(delims, 0),
	}

	// eat closing '__' if it is the closing delimiter
	if p.ch == '_' && p.peek() == '_' {
		p.next()
		p.next()
	}

	if trace {
		p.print("return\n" + printer.Pretty(em, p.indent+1))
	}

	return em
}

// parseStrong parses until a '**' and consumes it.
func (p *Parser) parseStrong(delims delimiters) *node.Strong {
	if trace {
		defer p.trace("parseStrong")()
	}

	// eat opening '**'
	p.next()
	p.next()

	// no possible duplicates because p.parseInline() returns on delim match
	delims.double = append(delims.double, '*')

	strong := &node.Strong{
		Children: p.parseInline(delims, 0),
	}

	// eat closing '**' if it is the closing delimiter
	if p.ch == '*' && p.peek() == '*' {
		p.next()
		p.next()
	}

	if trace {
		p.print("return\n" + printer.Pretty(strong, p.indent+1))
	}

	return strong
}

// parseLink parses link.
//
// Link can consist from one or two parts:
// <link-destination> | <link-text><link-destination>
//
// Link destination is plain text and is also used as link text if link text is
// no present.
// Link text is inline content.
func (p *Parser) parseLink(delims delimiters) *node.Link {
	if trace {
		defer p.trace("parseLink")()
	}

	p.next() // eat opening '<'

	var children []node.Inline
	linkVariant := p.linkVariant()

	// parse link text if a two part link
	// <link-text><link-destination>
	if linkVariant == twoPartLink {
		delims.single = append(delims.single, '>')
		children = p.parseInline(delims, '<')

		p.next() // eat closing '>' of link text
		p.next() // eat opening '<' of link destination
	}

	// we may have multiple offsets as we need to leave out '\' in escapes
	var offsets [][2]int

	// parse link destination (also link text if no link text present)
	offs := p.offset
	for p.ch != '\n' && p.ch != 0 {
		// escape sequences
		if p.ch == '\\' && (p.peek() == '\\' || p.peek() == '>') {
			if offs != p.offset {
				offsets = append(offsets, [2]int{offs, p.offset})
			}

			p.next()        // eat '\'
			offs = p.offset // set new offset so we leave out '\'
			p.next()        // eat escaped char
			continue
		}

		if p.ch == '>' {
			break
		}

		p.next()
	}

	// add last offset
	offsets = append(offsets, [2]int{offs, p.offset})

	var text string
	for _, o := range offsets {
		text += p.src[o[0]:o[1]]
	}

	p.next() // eat closing '>'

	// use link destination as link text if one part link
	if linkVariant == onePartLink {
		children = append(children, &node.Text{
			Value: text,
		})
	}

	link := &node.Link{
		Destination: text,
		Children:    children,
	}

	if trace {
		p.print("return\n" + printer.Pretty(link, p.indent+1))
	}

	return link
}

// linkVariant determines whether link consists of one or two consecutive parts:
// <link-destination> | <link-text><link-destination>
func (p *Parser) linkVariant() linkVariant {
	if trace {
		defer p.trace("linkVariant")()
	}

	// opening '<' already consumed

	// reset pointers to where they were before calling linkVariant
	defer p.reset(p.offset)

	for p.ch != '\n' && p.ch != 0 {
		// escape sequences
		if p.ch == '\\' && (p.peek() == '\\' || p.peek() == '>') {
			// eat escape sequence
			p.next()
			p.next()
			continue
		}

		if p.ch == '>' {
			break
		}

		p.next()
	}

	var lv linkVariant
	if p.ch == '>' && p.peek() == '<' {
		lv = twoPartLink
	} else {
		lv = onePartLink
	}

	if trace {
		p.print("return " + lv.String())
	}

	return lv
}

// parseText parses until an inline delimiter, extra delimiter, EOL, or EOF.
// exclude excludes the provided character as a delimiter if not 0.
func (p *Parser) parseText(extraDelims delimiters, exclude byte) *node.Text {
	if trace {
		defer p.trace("parseText")()
	}

	// add extra delims
	inDelims := inlineDelims
	inDelims.single = append(inDelims.single, extraDelims.single...)
	inDelims.double = append(inDelims.double, extraDelims.double...)

	// we may have multiple offsets as we need to leave out '\' in escapes
	var offsets [][2]int

	offs := p.offset
	for p.ch != '\n' && p.ch != 0 {
		// escape sequences; do not escape excluded delim if present
		if p.ch == '\\' && isEscapeChar(p.peek(), exclude) {
			// add offset if consumed any chars before
			if offs != p.offset {
				offsets = append(offsets, [2]int{offs, p.offset})
			}

			p.next()        // eat '\'
			offs = p.offset // set new offset so we leave out '\'
			p.next()        // eat escaped char
			continue
		}

		// consume excluded delim as text if present
		if exclude != 0 && p.ch == exclude {
			p.next()
			continue
		}

		if contains(inDelims.single, p.ch) {
			break
		}

		if p.ch == p.peek() && contains(inDelims.double, p.ch) &&
			contains(inDelims.double, p.peek()) {
			break
		}

		p.next()
	}

	// add last offset
	offsets = append(offsets, [2]int{offs, p.offset})

	text := &node.Text{}
	for _, o := range offsets {
		text.Value += p.src[o[0]:o[1]]
	}

	if trace {
		p.print("return " + printer.Pretty(text, 0))
	}

	return text
}

func isEscapeChar(ch, exclude byte) bool {
	// do not treat excluded delim as escape if present
	if exclude != 0 && ch == exclude {
		return false
	}

	// always escape '\' and '>' for consistency
	return ch == '\\' || ch == '>' ||
		contains(inlineDelims.single, ch) ||
		contains(inlineDelims.double, ch) ||
		contains(blockDelims.single, ch) ||
		contains(blockDelims.double, ch)
}

func (p *Parser) trace(msg string) func() {
	p.print(msg + " (")
	p.indent++

	return func() {
		p.indent--
		p.print(")")
	}
}

func (p *Parser) print(msg string) {
	fmt.Println(strings.Repeat(".   ", p.indent) + msg)
}

// char returns a string representation of a character.
func char(ch byte) string {
	s := string(ch)

	switch ch {
	case 0:
		s = "EOF"
	case '\t':
		s = "\\t"
	case '\n':
		s = "\\n"
	}

	return "'" + s + "'"
}

// contains determines whether needle is in haystack.
func contains(haystack []byte, needle byte) bool {
	for _, v := range haystack {
		if v == needle {
			return true
		}
	}
	return false
}
