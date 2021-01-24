package scanner

import (
	"errors"
	"fmt"
	"to/internal/token"
	"unicode/utf8"
)

type Scanner interface {
	Scan() (token.Token, string)
}

// Mode controls which tokens to return.
type Mode uint

// Mode flags
const (
	ScanComments Mode = 1 << iota
)

// ErrorHandler is called with an error message and count if an error is
// encounterd.
type ErrorHandler func(err error, errCount uint)

// scanner holds the scanning state.
type scanner struct {
	src        string       // source
	mode       Mode         // which tokens to return
	errHandler ErrorHandler // error callback func

	ch       byte // current character
	offs     int  // current offset
	rdOffs   int  // read offset
	errCount uint // number of errors encountered
}

// New returns a new Scanner.
func New(src string, mode Mode, errHandler ErrorHandler) Scanner {
	s := &scanner{
		src:        src,
		mode:       mode,
		errHandler: errHandler,
	}
	s.next() // initialize pointers
	return s
}

// error increments s.errCount and calls s.errHandler() if present.
func (s *scanner) error(err error) {
	s.errCount++
	if s.errHandler != nil {
		s.errHandler(err, s.errCount)
	}
}

// illegal character errors
var (
	ErrIllegalNUL          = errors.New("illegal character NUL")
	ErrIllegalUTF8Encoding = errors.New("ilelgal UTF-8 encoding")
	ErrIllegalBOM          = errors.New("illegal byte order mark")
)

// next reads the next char into s.ch.
//
// s.ch == utf8.RuneSelf if a non-ASCII char is read. We skip non-ASCII chars
// as we never use them in our notations—they are always just content.
//
// NUL chars, BOMs in the middle, or invalid UTF-8 encoding call s.error().
// NULs are replaced with the replacement char. BOMs in the middle and invalid
// UTF-8 encoding chars are skipped.
//
// An encoding is invalid if it is incorrect UTF-8, encodes a rune that is out
// of range, or is not the shortest possible UTF-8 encoding for the value.
func (s *scanner) next() {
skip:
	// handle end of file
	if s.rdOffs >= len(s.src) {
		s.ch = 0
		s.offs = len(s.src)
		return
	}

	ch := s.src[s.rdOffs]

	// handle NUL
	// 1. replace NUL with the replacement char U+FFFD
	// 2. s.ch = utf8.RuneSelf—a one-byte char we never use
	if ch == 0 {
		s.error(fmt.Errorf("%w: %U", ErrIllegalNUL, ch))
		s.src = s.src[:s.rdOffs] + string(utf8.RuneError) + s.src[s.rdOffs+1:]
		s.offs = s.rdOffs
		s.rdOffs += utf8.RuneLen(utf8.RuneError)
		s.ch = utf8.RuneSelf
		return
	}

	// 1. error and skip if invalid UTF-8 encoding
	// 2. skip if BOM and error if in the middle
	// 3. s.rdOffs = after code point; s.ch = first byte of the code point
	if ch >= utf8.RuneSelf {
		r, w := utf8.DecodeRuneInString(s.src[s.rdOffs:])
		if r == utf8.RuneError && w == 1 {
			s.error(fmt.Errorf("%w: %U", ErrIllegalUTF8Encoding, r))
			s.rdOffs += w
			goto skip
		}

		const BOM = 0xFEFF
		if r == BOM {
			if s.rdOffs > 0 {
				s.error(fmt.Errorf("%w: %U", ErrIllegalBOM, r))
			}
			s.rdOffs += w
			goto skip
		}

		s.offs = s.rdOffs
		s.rdOffs += w
		s.ch = ch
		return
	}

	s.offs = s.rdOffs
	s.rdOffs++
	s.ch = ch
}

// peek returns the byte following the most recently read character without
// advancing the scanner. peek returns 0 if at EOF.
func (s *scanner) peek() byte {
	if s.rdOffs < len(s.src) {
		return s.src[s.rdOffs]
	}
	return 0
}

func (s *scanner) Scan() (token.Token, string) {
	var tok token.Token
	var lit string

skip:
	switch {
	case s.ch == 0:
		tok = token.EOF
	case s.ch == '\n':
		tok = token.LINEFEED
		lit = "\n"
	case s.ch == '\t' || s.ch == ' ':
		return token.INDENT, s.scanIndent()
	case s.ch == '/' && s.peek() == '/':
		comment := s.scanComment()
		if s.mode&ScanComments == 0 {
			goto skip
		}
		tok = token.COMMENT
		lit = comment
	case s.ch == '|':
		tok = token.VLINE
		lit = string(s.ch)
	case s.ch == '>':
		tok = token.GT
		lit = string(s.ch)
	default:
		return token.TEXT, s.scanText()
	}

	s.next()
	return tok, lit
}

// scanIndent scans spacing.
func (s *scanner) scanIndent() string {
	offs := s.offs
	for s.ch == '\t' || s.ch == ' ' {
		s.next()
	}
	return s.src[offs:s.offs]
}

// scanComment scans until line feed or EOF. Delimiter is included.
func (s *scanner) scanComment() string {
	// first '/' already consumed, onsume the second '/'
	offs := s.offs
	s.next()

	for s.ch != '\n' && s.ch > 0 {
		s.next()
	}

	return s.src[offs:s.offs]
}

// scanText scans until an inline element delimiter, line feed, or EOF.
func (s *scanner) scanText() string {
	offs := s.offs
	for s.ch != '\n' && s.ch > 0 {
		s.next()
	}
	return s.src[offs:s.offs]
}
