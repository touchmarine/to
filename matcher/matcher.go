// package matcher provides a matcher interface and related functions. Matchers
// are used by the parser to determine the contents of prefix elements.
package matcher

import (
	"github.com/touchmarine/to/matcher/url"
)

// Matcher recognizes patterns and returns the length of the match.
type Matcher interface {
	Match([]byte) int
}

// MatcherFunc is an adapter like http.HandlerFunc is for http.Handler. It
// allows the use of an ordinary function as a Matcher.
type MatcherFunc func([]byte) int

// Match calls m(p).
func (m MatcherFunc) Match(p []byte) int {
	return m(p)
}

// Map maps matcher names to matchers.
type Map map[string]Matcher

// Defaults retuns a Map of default matchers.
func Defaults() Map {
	return Map{
		"url": MatcherFunc(url.Match),
	}
}
