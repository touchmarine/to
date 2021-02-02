package token

//go:generate stringer -type=Token
type Token uint

// tokens
const (
	// special
	EOF      Token = iota
	LINEFEED       // newline
	INDENT         // tab or space
	COMMENT        // //-comment

	// block
	VLINE        // vertical line "|"
	GT           // greater-than sign ">"
	HYPEN        // hypen-minus "-"
	GRAVEACCENTS // "`"

	// inline
	LOWLINES // "_"
	GAP      // grave accent "`" + punctuation
	PAG      // punctuation + grave accent "`"
	LTP      // less-than sign "<" + puncutation
	PTL      // punctuation + less-than sign "<"
	DPUNCT   // double punctuation "**", "<>", ...
	TEXT     // text
)
