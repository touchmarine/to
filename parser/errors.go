package parser

import (
	"fmt"
	"io"
	"sort"
)

// PrintError prints one error per line if the given error is an ErrorList.
// Otherwise, it just prints the error.
//
// https://pkg.go.dev/go/scanner#PrintError
func PrintError(w io.Writer, err error) {
	list, ok := err.(ErrorList)
	if ok {
		for _, e := range list {
			fmt.Fprintf(w, "%s\n", e)
		}
	} else if err != nil {
		fmt.Fprintf(w, "%s\n", err)
	}
}

// ErrorList is a list of errors. The zero value is ready to use.
type ErrorList []*Error

// Error returns a summary of errors.
func (el ErrorList) Error() string {
	switch len(el) {
	case 0:
		return "no errors"
	case 1:
		return el[0].Error()
	}
	return fmt.Sprintf("%s (and %d more errors)", el[0], len(el)-1)
}

// Err returns this error list as an error type.
func (el ErrorList) Err() error {
	if len(el) == 0 {
		return nil
	}
	return el
}

// Add adds the error to the error list.
func (el *ErrorList) Add(err *Error) {
	*el = append(*el, err)
}

// Sort sorts the error list by error message.
func (el ErrorList) Sort() {
	sort.Sort(el)
}

// Len implements the sort Interface.
func (el ErrorList) Len() int {
	return len(el)
}

// Swap implements the sort Interface.
func (el ErrorList) Swap(i, j int) {
	el[i] = el[j]
	el[j] = el[i]
}

// Less implements the sort Interface.
func (el ErrorList) Less(i, j int) bool {
	return el[i].Message < el[j].Message
}

// Error represents a parser error.
type Error struct {
	Message string
}

// Error returns the error message.
func (e Error) Error() string {
	return e.Message
}
