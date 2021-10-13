package parser

import (
	"fmt"
	"io"
	"sort"
)

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

type ErrorList []*Error

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

func (el *ErrorList) Add(err *Error) {
	*el = append(*el, err)
}

func (el ErrorList) Sort() {
	sort.Sort(el)
}

func (el ErrorList) Len() int {
	return len(el)
}

func (el ErrorList) Swap(i, j int) {
	el[i] = el[j]
	el[j] = el[i]
}

func (el ErrorList) Less(i, j int) bool {
	return el[i].Message < el[j].Message
}

type Error struct {
	Message string
}

func (e Error) Error() string {
	return e.Message
}
