package node

import (
	"encoding/json"
	"strings"
)

type Block interface {
	block()
}

type Paragraph struct {
	Children []Block
}

func (p *Paragraph) block() {}

type Lines []string

func (p Lines) block() {}

func Pretty(a interface{}) string {
	b, err := json.MarshalIndent(a, "", "\t")
	if err != nil {
		panic("parser.Pretty: json.MarshalIndent() failed")
	}
	return string(b)
}

func PrettyIndented(a interface{}, indent int) string {
	b, err := json.MarshalIndent(a, strings.Repeat("\t", indent), "\t")
	if err != nil {
		panic("node.PrettyIndented: json.MarshalIndent() failed")
	}
	return string(b)
}
