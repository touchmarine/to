// refactortests converts old parser table tests to golden files.
//
// refactortests takes the old parser_test.go file, in which the tests were
// written using the old-style nodes, and writes test cases as input files:
// testdata/<testname>/<testcase>.input
//
// Output input files will then be read by the new parser tester and compared to
// golden files (expected result in a file). To create/update .golden files, use
// go test ./parser -update
//
// Each test directory must have an elements.json file which defines the parser
// elements. elements.json is added manually; to json-marshal element map use:
// https://play.golang.org/p/YPc2RnqxHNS
//
// Caveats (need to do manually):
// - tests that do not use a cases table, such as TestBOM, are not handled
// - expected errors are not handled
//
// NOTE: On publishing the module, go module zip creation failed as many ASCII
//       characters are not allowed in the path (e.g. '\', '*', '>').
//       https://github.com/golang/vgo/blob/master/vendor/cmd/go/internal/module/module.go
package main

import (
	"encoding/json"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
)

const trace = true

const (
	testfile  = "parser_test.bak"
	parserdir = "parser"
	testdata  = "testdata"
)

func main() {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filepath.Join(parserdir, testfile), nil, 0)
	if err != nil {
		panic(err)
	}

	// very long output, around ~70k lines
	//if trace {
	//	ast.Print(fset, f)
	//}

	// m holds tests cases bound to a test function.
	m := map[string][]string{}
	v := inspector{
		m: m,
	}
	ast.Inspect(f, v.inspectTestFunctions)

	if trace {
		b, err := json.MarshalIndent(m, "", "\t")
		if err != nil {
			log.Fatal(err)
		}

		log.Print(string(b))
	}

	for testName, cases := range m {
		for _, c := range cases {
			addTest(testName, c)
		}
	}
}

// clashes is used to check for duplicates.
// map["testName"]map["filename"]whateverBool
var clashes = map[string]map[string]bool{}

// addTest writes a test case as an input file.
func addTest(name, testcase string) {
	dir := filepath.Join(parserdir, testdata, name)
	_, err := os.Stat(dir)
	if errors.Is(err, fs.ErrNotExist) {
		err = os.Mkdir(dir, os.ModePerm)
		if err != nil {
			panic(err)
		}
	} else if err != nil {
		panic(err)
	}

	filename := filepath.Join(dir, normalize(testcase)+".input")
	if _, ok := clashes[name][filename]; ok {
		log.Fatalf("found duplicate; testname=%s case=%s filename=%s", name, testcase, filename)
	}

	clashes[name] = map[string]bool{}
	clashes[name][filename] = true

	if trace {
		log.Printf("write %s", filename)
	}
	//os.Remove(filename)
	if err := os.WriteFile(filename, []byte(testcase), 0644); err != nil {
		panic(err)
	}
}

// normalize escapes some control characters and /. It is used to normalize test
// filenames.
func normalize(s string) string {
	if s == "" {
		return "\"\""
	}

	var b strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == ' ' || s[i] == '\t' || s[i] == '\n' || s[i] == '/' {
			x := s[i]
			n := 1
			for ; i+n < len(s) && s[i+n] == x; n++ {
				// count sequence
			}

			if n > 1 {
				b.WriteString(strconv.FormatInt(int64(n), 10))
			}

			switch x {
			case ' ':
				b.WriteString("SP")
			case '\t':
				b.WriteString("TAB")
			case '\n':
				b.WriteString("NL")
			case '/':
				b.WriteString("SL")
			default:
				panic("unexpected repeated char")
			}

			i += n
		} else {
			b.WriteByte(s[i])
			i++
		}
	}

	return b.String()
}

type inspector struct {
	// map["name"]["cases"]
	m     map[string][]string
	cases []string // unquoted case ins
}

// inspectTestFunctions traverses Test functions and cases within.
func (v *inspector) inspectTestFunctions(n ast.Node) bool {
	fn, isFunc := n.(*ast.FuncDecl)
	if isFunc && strings.HasPrefix(fn.Name.Name, "Test") {
		name := strings.TrimPrefix(fn.Name.Name, "Test")

		ast.Inspect(fn, v.inspectCases)
		v.m[uncapitalize(name)] = v.cases
		v.cases = nil
	}

	return true // continue inspect
}

func (v *inspector) inspectCases(n ast.Node) bool {
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

								v.cases = append(v.cases, value)
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
	} else if isUpper(s) {
		return s
	}

	r := []rune(s)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}

func isUpper(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) && !unicode.IsUpper(r) {
			return false
		}
	}
	return true
}
