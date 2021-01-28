package token

//go:generate stringer -type=Token
type Token uint

// tokens
const (
	EOF      Token = iota
	LINEFEED       // newline
	INDENT         // tab or space
	COMMENT        // //-comment

	VLINE        // vertical line "|"
	GT           // greater-than sign ">"
	HYPEN        // hypen-minus "-"
	GRAVEACCENTS // "`"
	TEXT         // text
)
