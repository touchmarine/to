package template

import (
	"bufio"
	"errors"
	"io"
	"strings"
)

func ParseAttributes(s string) map[string]string {
	p := attributeParser{}
	reader := strings.NewReader(s)
	p.init(reader)
	p.parse()
	return p.attrs
}

// attributeParser holds the parser state.
//
// attributeParser parses HTML-like attributes. An attribute has the folowing
// form:
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
type attributeParser struct {
	reader *bufio.Reader

	ch    byte
	attrs map[string]string
}

func (p *attributeParser) init(r io.Reader) {
	p.reader = bufio.NewReader(r)
	p.attrs = map[string]string{}
}

func (p *attributeParser) next() bool {
	b, err := p.reader.ReadByte()
	if err != nil {
		return false
	}

	p.ch = b
	return true
}

func (p *attributeParser) peek() byte {
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

func (p attributeParser) isSpacing() bool {
	return p.ch == '\t' || p.ch == ' '
}

func (p *attributeParser) parse() {
	for p.next() {
		if p.isSpacing() || p.ch == '\n' {
			continue
		}

		name, value := p.parseAttribute()
		if name != "" {
			p.attrs[name] = value
		}
	}
}

func (p *attributeParser) parseAttribute() (string, string) {
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
