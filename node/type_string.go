// Code generated by "stringer -type=Type -linecomment"; DO NOT EDIT.

package node

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[TypeError-0]
	_ = x[TypeContainer-1]
	_ = x[TypeWalled-2]
	_ = x[TypeVerbatimWalled-3]
	_ = x[TypeHanging-4]
	_ = x[TypeRankedHanging-5]
	_ = x[TypeFenced-6]
	_ = x[TypeVerbatimLine-7]
	_ = x[TypeLeaf-8]
	_ = x[TypeUniform-9]
	_ = x[TypeEscaped-10]
	_ = x[TypePrefixed-11]
	_ = x[TypeText-12]
}

const _Type_name = "ErrorContainerWalledVerbatimWalledHangingRankedHangingFencedVerbatimLineLeafUniformEscapedPrefixedText"

var _Type_index = [...]uint8{0, 5, 14, 20, 34, 41, 54, 60, 72, 76, 83, 90, 98, 102}

func (i Type) String() string {
	if i < 0 || i >= Type(len(_Type_index)-1) {
		return "Type(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Type_name[_Type_index[i]:_Type_index[i+1]]
}
