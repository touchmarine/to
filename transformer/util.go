package transformer

import (
	"github.com/touchmarine/to/node"
)

// isBlank determines whether ln is a blank line.
func isBlank(ln *node.Line) bool {
	return len(ln.InlineChildren()) == 0
}

// https://github.com/golang/go/wiki/SliceTricks
func cut(a []node.Node, i, j int) []node.Node {
	copy(a[i:], a[j:])
	for k, n := len(a)-j+i, len(a); k < n; k++ {
		a[k] = nil
	}
	a = a[:len(a)-j+i]
	return a
}
