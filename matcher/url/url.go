// Package url provides an URL matcher.
package url

import (
	"unicode"
	"unicode/utf8"
)

// Match matches an URL after the scheme. It matches a valid domain and if it is
// present a relative reference (path + query + fragment). Trailing puncutation
// and trailing unmatched parentheses are not part of the URL.
//
// A valid domain consists of segments of alphanumeric characters, underscores
// "_", and hypens "-" separated by periods ".". No underscores may be present
// in the last two segments of the domain.
//
// https://github.github.com/gfm/#valid-domain
func Match(p []byte) int {
	end := 0

	domainLen := validDomain(p)
	if domainLen == 0 {
		// invalid domain
		return 0
	}

	end += domainLen

	refLen := naiveRelativeRef(p[domainLen:])
	end += refLen

	// remove trailing punct and unmatched parens from url
	end = preciseRelativeRefEnd(p[:end])

	return end
}

// validDomain returns the length of a valid domain.
func validDomain(p []byte) int {
	u1, u2 := false, false // underscores in last two segments

	i := 0
	for ; i < len(p); i++ {
		ch := p[i]

		if ch == '_' {
			u2 = true
		} else if ch == '.' {
			u1 = u2
			u2 = false
		} else if ch == '-' {
		} else if r := validRune(utf8.DecodeRune(p[i:])); ch == 0 || unicode.IsSpace(r) || unicode.IsPunct(r) {
			break
		}
	}

	if u1 || u2 {
		// underscore present in last two segments
		return 0
	}

	return i
}

// naiveRelativeRef naively determines the length of a relative reference (path
// + query + fragment).
func naiveRelativeRef(p []byte) int {
	i := 0
	for ; i < len(p); i++ {
		ch := p[i]

		if ch == 0 || ch == '\n' || isSpacing(ch) {
			break
		}
	}
	return i
}

// preciseRelativeRefEnd takes an entire naive url and removes trailing
// punctuation and unmatched trailing parentheses.
func preciseRelativeRefEnd(p []byte) int {
	end := len(p)

	for end > 0 {
		if isTrailingPunct(p[end-1]) {
			end--
		} else if p[end-1] == ')' {
			if matchingParens(p[:end]) {
				break
			} else {
				// unmatched ")", remove it
				end--
			}
		} else {
			break
		}
	}

	return end
}

func isTrailingPunct(ch byte) bool {
	return ch == '?' || ch == '!' || ch == '.' || ch == ',' || ch == ':' ||
		ch == '*' || ch == '_' || ch == '~' || ch == '\'' || ch == '"'
}

// matchingParens determines whether number of "(" and ")" match.
func matchingParens(p []byte) bool {
	opening, closing := 0, 0

	for i := 0; i < len(p); i++ {
		ch := p[i]

		if ch == '(' {
			opening++
		} else if ch == ')' {
			closing++
		}
	}

	return closing <= opening
}

// same as in parser/parser.go, cycle import not allowed
func isSpacing(ch byte) bool {
	return ch == ' ' || ch == '\t'
}

func validRune(r rune, w int) rune {
	switch r {
	case utf8.RuneError:
		if w == 0 {
			panic("parser: cannot decode empty slice")
		} else if w == 1 {
			return utf8.RuneError
		} else {
			panic("parser: utf8 lib error")
		}
	case '\u0000', '\uFEFF': // NULL, or BOM
		return utf8.RuneError
	}
	return r
}
