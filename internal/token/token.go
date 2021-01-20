package token

//go:generate stringer -type=Token
type Token uint

// tokens
const (
	EOF     Token = iota
	Indent        // larger indentation
	Dedent        // smaller indentation
	Newline       // \n
	Comment       // //

	BlockDelim  // |
	InlineDelim // __
	Sep         // <>
	Text        // text
)

// IsBlockDelim determines whether ch is a block delimiter.
func IsBlockDelim(ch byte) bool {
	for _, b := range blockDelims {
		if b == ch {
			return true
		}
	}
	return false
}

var blockDelims = [...]byte{
	// walled
	'|',
	'>',

	// offside
	'-',
}

// IsInlineDelim determines wheter cur and next can comprise an inline
// delimiter.
func IsInlineDelim(cur, next byte) bool {
	for i, b := range inlineDelims {
		if b == cur {
			// uniform delim: next == cur
			if i < 2 {
				return next == inlineDelims[i]
			}

			// escaped or composite delim: next == puncutation char
			return isPunct(next)
		}
	}
	return false
}

var inlineDelims = [...]byte{
	// uniform
	'_',
	'*',

	// escaped
	'`',

	// composite
	'<',
}

func isPunct(ch byte) bool {
	return ch >= 0x20 && ch <= 0x2F ||
		ch >= 0x3A && ch <= 0x40 ||
		ch >= 0x5B && ch <= 0x60 ||
		ch >= 0x7B && ch <= 0x7E
}
