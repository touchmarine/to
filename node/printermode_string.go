// Code generated by "stringer -type=PrinterMode"; DO NOT EDIT.

package node

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[PrintData-1]
	_ = x[PrintLocation-2]
}

const _PrinterMode_name = "PrintDataPrintLocation"

var _PrinterMode_index = [...]uint8{0, 9, 22}

func (i PrinterMode) String() string {
	i -= 1
	if i < 0 || i >= PrinterMode(len(_PrinterMode_index)-1) {
		return "PrinterMode(" + strconv.FormatInt(int64(i+1), 10) + ")"
	}
	return _PrinterMode_name[_PrinterMode_index[i]:_PrinterMode_index[i+1]]
}
