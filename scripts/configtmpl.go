// configtmpl takes a config JSON template as stdin and outputs JSON. It
// provides raw multiline strings delimited by triple single quotes "'".
// Immediate newline after the delimiter is discarded if present.
//
// Usage:
//
// go run scripts/configtmpl < config/to.jsontmpl > config/to.json
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
)

func main() {
	var p parser
	p.init(os.Stdin, os.Stdout)
	p.parse()
}

type parser struct {
	reader *bufio.Reader
	w      io.Writer

	ch byte
}

func (p *parser) init(r io.Reader, w io.Writer) {
	p.reader = bufio.NewReader(r)
	p.w = w
}

func (p *parser) next() bool {
	b, err := p.reader.ReadByte()
	if err != nil {
		p.ch = 0
		return false
	}

	p.ch = b
	return true
}

// isRawDelim determines whether peek characters are three consecutive single
// quotes "'''".
func (p *parser) isRawDelim() bool {
	if p.ch != '\'' {
		return false
	}

	b, err := p.reader.Peek(2)
	if err != nil && !errors.Is(err, io.EOF) {
		panic(err)
	}

	return bytes.Compare(b, []byte("''")) == 0
}

func (p *parser) parse() {
	p.next()

	for p.ch != 0 {
		if p.isRawDelim() {
			p.parseRaw()
			continue
		}

		p.w.Write([]byte{p.ch})

		p.next()
	}
}

// parseRaw parses a raw string like a TOML multi-line literal string. It is
// delimited by three single quotes "'". Immediate newline after the opening
// delimiter is discarded if present.
func (p *parser) parseRaw() {
	// consume opening delimiter
	if !p.consume(3) {
		return
	}

	if p.ch == '\n' {
		// immediate newline
		if !p.next() {
			return
		}
	}

	var b bytes.Buffer

	for p.ch != 0 {
		if p.isRawDelim() {
			// closing delimiter
			p.consume(3)
			break
		}

		b.WriteByte(p.ch)

		p.next()
	}

	if b.Len() > 0 {
		var j bytes.Buffer

		e := json.NewEncoder(&j)
		e.SetEscapeHTML(false)
		if err := e.Encode(b.String()); err != nil {
			panic(err)
		}

		p.w.Write(j.Bytes()[:j.Len()-1]) // remove trailing newline added by encoder
	}

	return
}

// consume consumes n characters and returns whether it succeeded or EOF was
// encountered.
func (p *parser) consume(n int) bool {
	for i := 0; i < n; i++ {
		if !p.next() {
			return false
		}
	}
	return true
}
