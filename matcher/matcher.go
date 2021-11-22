// Package matcher provides a matcher interface which is used by the parser to
// determine the contents of prefix elements.
package matcher

import (
	"github.com/touchmarine/to/matcher/url"
)

// Matcher recognizes patterns in the given bytes and returns the length of the
// match.
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

// Map is a map of matcher names to Matchers.
type Map map[string]Matcher

// Defaults returns a Map of default matchers.
func Defaults() Map {
	return Map{
		"url": MatcherFunc(url.Match),
	}
}
