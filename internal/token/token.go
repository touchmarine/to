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

	Pipeline    // |
	GreaterThan // >
	Text        // text
)
