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
	UNDERSCORES // "_"
	GAPUNCT     // grave accent "`" + punctuation
	PUNCTGA     // punctuation + grave accent "`"
	TEXT        // text
)
