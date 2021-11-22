package node

import (
	"fmt"
	"strings"
)

//go:generate stringer -type=Type -linecomment
type Type int

const (
	// special
	TypeError     Type = iota // Error
	TypeContainer             // Container

	// blocks
	TypeWalled         // Walled
	TypeVerbatimWalled // VerbatimWalled
	TypeHanging        // Hanging
	TypeRankedHanging  // RankedHanging
	TypeFenced         // Fenced
	TypeVerbatimLine   // VerbatimLine
	TypeLeaf           // Leaf

	// inlines
	TypeUniform  // Uniform
	TypeEscaped  // Escaped
	TypePrefixed // Prefixed
	TypeText     // Text
)

// UnmarshalText implements the encoding.TextUnmarshaler interface. It is
// case-insensitive and supports only valid Types (all except TypeError and
// TypeContainer).
func (t *Type) UnmarshalText(text []byte) error {
	s := strings.ToLower(string(text))
	if tt, ok := validTypes[s]; ok {
		*t = tt
	} else {
		return fmt.Errorf("unexpected Type value: %q", text)
	}
	return nil
}

// validTypes maps Types to theirs lower-case strings representations. Valid
// types are all but the special types (TypeError and TypeContainer).
var validTypes = map[string]Type{
	strings.ToLower(TypeWalled.String()):         TypeWalled,
	strings.ToLower(TypeVerbatimWalled.String()): TypeVerbatimWalled,
	strings.ToLower(TypeHanging.String()):        TypeHanging,
	strings.ToLower(TypeRankedHanging.String()):  TypeRankedHanging,
	strings.ToLower(TypeFenced.String()):         TypeFenced,
	strings.ToLower(TypeVerbatimLine.String()):   TypeVerbatimLine,
	strings.ToLower(TypeLeaf.String()):           TypeLeaf,

	strings.ToLower(TypeUniform.String()):  TypeUniform,
	strings.ToLower(TypeEscaped.String()):  TypeEscaped,
	strings.ToLower(TypePrefixed.String()): TypePrefixed,
	strings.ToLower(TypeText.String()):     TypeText,
}

// IsBlock reports whether the given type is a member of the block type set.
//
// Note that a non-block type is not necessarily an inline type:
// 	!IsBlock() != IsInline()
func IsBlock(t Type) bool {
	return t >= TypeWalled && t <= TypeLeaf
}

// IsInline reports whether the given type is a member of the inline type set.
//
// Note that a non-inline type is not necessarily a block type:
// 	!IsInline() != IsBlock()
func IsInline(t Type) bool {
	return t >= TypeUniform
}

// HasDelimiter reports whether the given type has a delimiter.
func HasDelimiter(t Type) bool {
	return t == TypeWalled || t == TypeVerbatimWalled || t == TypeHanging ||
		t == TypeRankedHanging || t == TypeFenced || t == TypeVerbatimLine ||
		t == TypeUniform || t == TypeEscaped || t == TypePrefixed
}
