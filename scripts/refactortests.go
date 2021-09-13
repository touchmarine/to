package main

import (
	"encoding/json"
	"github.com/touchmarine/to/node"
	toparser "github.com/touchmarine/to/parser"
	"github.com/touchmarine/to/stringifier"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strconv"
	"strings"
	"unicode"
)

const debug = true

func main() {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "parser/parser_test.go", nil, 0)
	if err != nil {
		panic(err)
	}

	// very long output, around ~70k lines
	//if debug {
	//	ast.Print(fset, f)
	//}

	m := map[string][]string{}
	v := inspector{
		m: m,
	}
	ast.Inspect(f, v.inspectTestFunctions)

	if debug {
		b, err := json.MarshalIndent(m, "", "\t")
		if err != nil {
			log.Fatal(err)
		}

		log.Print(string(b))
	}

	n, errs := toparser.Parse(strings.NewReader("a\nb"), toparser.ElementMap{
		"T": {
			Name: "T",
			Type: node.TypeLeaf,
		},
		"MT": {
			Name: "MT",
			Type: node.TypeText,
		},
	})
	for _, err := range errs {
		log.Print(err)
	}

	stringifier.StringifyTo(os.Stdout, n)
}

type inspector struct {
	// map["name"]["inputs"]
	m      map[string][]string
	inputs []string // unquoted case ins
}

func (v *inspector) inspectTestFunctions(n ast.Node) bool {
	fn, isFunc := n.(*ast.FuncDecl)
	if isFunc && strings.HasPrefix(fn.Name.Name, "Test") {
		name := strings.TrimPrefix(fn.Name.Name, "Test")

		ast.Inspect(fn, v.inspectInputs)
		v.m[uncapitalize(name)] = v.inputs
		v.inputs = nil
	}

	return true // continue inspect
}

func (v *inspector) inspectInputs(n ast.Node) bool {
	assign, isAssign := n.(*ast.AssignStmt)
	if isAssign {
		ident, isIdent := assign.Lhs[0].(*ast.Ident)
		if isIdent && ident.Name == "cases" {
			comp, isComp := assign.Rhs[0].(*ast.CompositeLit)
			if isComp {
				for _, elt := range comp.Elts {
					comp1, isComp1 := elt.(*ast.CompositeLit)
					if isComp1 {
						for _, elt1 := range comp1.Elts {
							basic, isBasic := elt1.(*ast.BasicLit)
							if isBasic {
								value, err := strconv.Unquote(basic.Value)
								if err != nil {
									panic(err)
								}

								v.inputs = append(v.inputs, value)
							}
						}
					}
				}
			}
		}
	}

	return true // continue inspect
}

func uncapitalize(s string) string {
	if s == "" {
		return ""
	}

	r := []rune(s)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}
