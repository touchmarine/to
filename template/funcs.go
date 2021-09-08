package template

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/touchmarine/to/aggregator/seqnum"
	"html/template"
	"io"
	"strings"
)

const trace = false

var Functions = template.FuncMap{
	"mapAny":                  mapAny,
	"map":                     mapString,
	"get":                     get,
	"set":                     set,
	"hasKey":                  hasKey,
	"head":                    head,
	"body":                    body,
	"groupBySequentialNumber": groupBySequentialNumber,
	"isSequentialNumberGroup": isSequentialNumberGroup,
	"parseAttrs":              parseAttrs,
	"htmlAttrs":               htmlAttrs,
	"trimSpacing":             trimSpacing,
}

func mapAny(v ...interface{}) (map[string]interface{}, error) {
	if len(v)%2 > 0 {
		return nil, fmt.Errorf("map got odd number of parameters")
	}

	m := make(map[string]interface{}, len(v)/2)

	for i := 0; i < len(v); i += 2 {
		key, ok := v[i].(string)
		if !ok {
			return nil, fmt.Errorf("map key not a string")
		}

		m[key] = v[i+1]
	}

	return m, nil
}

func mapString(v ...string) (map[string]string, error) {
	if len(v)%2 > 0 {
		return nil, fmt.Errorf("map got odd number of parameters")
	}

	m := make(map[string]string, len(v)/2)

	for i := 0; i < len(v); i += 2 {
		m[v[i]] = v[i+1]
	}

	return m, nil
}

func get(m map[string]string, key string) (string, error) {
	if m == nil {
		return "", fmt.Errorf("nil map")
	}

	if v, ok := m[key]; ok {
		return v, nil
	}
	return "", nil
}

func set(m map[string]string, key, value string) (map[string]string, error) {
	if m == nil {
		return nil, fmt.Errorf("nil map")
	}

	m[key] = value
	return m, nil
}

func hasKey(m map[string]string, key string) bool {
	_, ok := m[key]
	return ok
}

func head(lines []string) string {
	if len(lines) > 0 {
		return lines[0]
	}
	return ""
}

func body(lines []string) string {
	if len(lines) > 1 {
		return strings.Join(lines[1:], "\n")
	}
	return ""
}

type sequentialNumberNode interface {
	sequentialNumberNode()
}

type sequentialNumberGroup []sequentialNumberNode

func (sequentialNumberGroup) sequentialNumberNode() {}

type sequentialNumberParticle seqnum.Particle

func (sequentialNumberParticle) sequentialNumberNode() {}

func isSequentialNumberGroup(v interface{}) bool {
	_, ok := v.(sequentialNumberGroup)
	return ok
}

func groupBySequentialNumber(aggregate seqnum.Aggregate) sequentialNumberGroup {
	s := sequentialNumberGrouper{}
	return s.groupBySequentialNumber(aggregate)
}

// sequentialNumberGrouper groups seqnum.Aggregate by their sequence number.
type sequentialNumberGrouper struct {
	lowest int
	indent int
}

func (s *sequentialNumberGrouper) groupBySequentialNumber(aggregate seqnum.Aggregate) sequentialNumberGroup {
	if trace {
		defer s.trace("groupBySequentialNumber")()
	}

	var group sequentialNumberGroup

	var depth int
	if len(aggregate) > 0 {
		depth = len(aggregate[0].SequentialNumbers)
	}

	if trace {
		s.printf("depth=%d", depth)
	}

	if s.lowest == 0 || s.lowest > 0 && depth < s.lowest {
		if trace {
			s.printf("set lowest=%d", depth)
		}
		s.lowest = depth
	}

	for i := 0; i < len(aggregate); i++ {
		particle := aggregate[i]
		cur := len(particle.SequentialNumbers) // current depth

		if trace {
			s.printf("particle %s", particle.SequentialNumber)
		}

		if cur > depth {
			if trace {
				s.printf("%d > depth:", cur)
			}

			g := s.groupBySequentialNumber(aggregate[i:])
			i += len(g) - 1
			group = append(group, g)
		} else if cur == depth {
			if trace {
				s.printf("%d == depth, add to group", cur)
			}

			group = append(group, sequentialNumberParticle(particle))
		} else if cur < depth {
			if trace {
				s.printf("%d < depth:", cur)
			}

			if cur < s.lowest {
				g := s.groupBySequentialNumber(aggregate[i:])
				i += len(g) - 1
				group = append(sequentialNumberGroup{group}, g...)
			}
			break
		}
	}

	return group
}

func (s *sequentialNumberGrouper) trace(msg string) func() {
	s.printf("%s (", msg)
	s.indent++

	return func() {
		s.indent--
		s.print(")")
	}
}

func (s *sequentialNumberGrouper) printf(format string, v ...interface{}) {
	s.print(fmt.Sprintf(format, v...))
}

func (s *sequentialNumberGrouper) print(msg string) {
	fmt.Println(strings.Repeat("\t", s.indent) + msg)
}

func parseAttrs(lines []string) map[string]string {
	reader := strings.NewReader(strings.Join(lines, " "))

	var p attrParser
	p.init(reader)
	p.parse()

	return p.attrs
}

// attrParser holds the parser state.
//
// attrParser parses HTML-like attributes. An attribute has the folowing form:
//
// <name> = <value>
//
// where <name> is a string of Unicode characters except spacing or newline and
// <value> is a string of Unicode characaters except spacing or a double quote
// '"' delimited string that can contain spacing and escape sequences or a
// single quote "'" delimited string that can contain raw content.
//
// Only the name is required. Spacing is a space or a tab.
//
// Attributes must be separated by spacing except after an attribute with quoted
// value—they can be placed one after another as such:
//
// a="b"c=d
type attrParser struct {
	reader *bufio.Reader

	ch byte

	attrs map[string]string
}

func (p *attrParser) init(r io.Reader) {
	p.reader = bufio.NewReader(r)
	p.attrs = make(map[string]string)
}

func (p *attrParser) next() bool {
	b, err := p.reader.ReadByte()
	if err != nil {
		return false
	}

	p.ch = b
	return true
}

func (p *attrParser) peek() byte {
	b, err := p.reader.Peek(1)
	if err != nil && !errors.Is(err, io.EOF) {
		panic(err)
	}

	if l := len(b); l == 0 {
		return 0
	} else if l == 1 {
		return b[0]
	} else {
		panic("template: unexpected byte length")
	}
}

func (p *attrParser) isSpacing() bool {
	return p.ch == '\t' || p.ch == ' '
}

func (p *attrParser) parse() {
	for p.next() {
		if p.isSpacing() || p.ch == '\n' {
			continue
		}

		name, value := p.parseAttr()
		if name != "" {
			p.attrs[name] = value
		}
	}
}

func (p *attrParser) parseAttr() (string, string) {
	var name, value strings.Builder

	var hitEquals bool // whether encountered name-value delimiter "="
	var quote byte     // which quote if any

	for {
		if hitEquals {
			// in value part of attribute

			if quote != 0 {
				// inside quotes

				if p.ch == quote {
					// closing quote
					break
				}

				if quote == '"' {
					// inside double quotes '"'

					if peek := p.peek(); p.ch == '\\' && (peek == '\\' || peek == '"') {
						// escape
						value.WriteByte(peek)

						if !p.next() { // skip escape backslash
							panic("template: no escape backslash")
						}
						if !p.next() { // skip escaped char
							panic("template: no escaped char")
						}

						continue
					}
				} else if quote == '\'' {
					// inside single quotes "'" (raw content)
				} else {
					panic("template: quote is neither '\"' or \"'\"")
				}

				value.WriteByte(p.ch)
			} else {
				if p.isSpacing() || p.ch == '\n' {
					break
				}

				value.WriteByte(p.ch)
			}
		} else if p.ch == '=' {
			// name-value delimiter

			hitEquals = true

			if peek := p.peek(); peek == '"' || peek == '\'' {
				// opening quote
				quote = peek

				if !p.next() {
					panic("template: no equals char")
				}
				if !p.next() {
					panic("template: no opening quote")
				}

				continue
			}
		} else {
			// in name part of attribute

			if p.isSpacing() || p.ch == '\n' {
				break
			}

			name.WriteByte(p.ch)
		}

		if !p.next() {
			break
		}
	}

	return name.String(), value.String()
}

// htmlAttrs returns a HTML-formatted string of attributes from the given attrs.
func htmlAttrs(attrs map[string]string) template.HTMLAttr {
	var b strings.Builder

	var i int
	for name, value := range attrs {
		if i > 0 {
			b.WriteString(" ")
		}

		b.WriteString(name)

		if value != "" {
			b.WriteString(`="` + value + `"`)
		}

		i++
	}

	return template.HTMLAttr(b.String())
}

func trimSpacing(s string) string {
	return strings.Trim(s, " \t")
}
