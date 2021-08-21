package matcher

import (
	"github.com/touchmarine/to/matcher/url"
)

// Matcher recognizes patterns.
//
// Match tests for pattern matches and returns the length of the match.
type Matcher interface {
	Match([]byte) int
}

// MatcherFunc is an adapter to allow the use of a function as a Matcher.
type MatcherFunc func([]byte) int

// Match calls m(p).
func (m MatcherFunc) Match(p []byte) int {
	return m(p)
}

// Map maps names to matchers.
type Map map[string]Matcher

// Defaults retuns a Map of default matchers.
func Defaults() Map {
	return Map{
		"url": MatcherFunc(url.Match),
	}
}
