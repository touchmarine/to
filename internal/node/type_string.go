// Code generated by "stringer -type=Type"; DO NOT EDIT.

package node

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[TypeLine-0]
	_ = x[TypeWalled-1]
	_ = x[TypeHanging-2]
	_ = x[TypeFenced-3]
	_ = x[TypeText-4]
}

const _Type_name = "TypeLineTypeWalledTypeHangingTypeFencedTypeText"

var _Type_index = [...]uint8{0, 8, 18, 29, 39, 47}

func (i Type) String() string {
	if i >= Type(len(_Type_index)-1) {
		return "Type(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Type_name[_Type_index[i]:_Type_index[i+1]]
}
