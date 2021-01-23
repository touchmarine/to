package token

//go:generate stringer -type=Token
type Token uint

// tokens
const (
	EOF      Token = iota
	LINEFEED       // \n
	COMMENT        // //
	INDENT         // larger indentation
	DEDENT         // smaller indentation

	BEGINBQUOTE // > at start of line or another block
	BEGINPARA   // | at start of line or another block
	ENDWALLED   // line starting with another or no block delimiter
	TEXT        // text
)
