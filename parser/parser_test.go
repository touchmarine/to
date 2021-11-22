package parser_test

import (
	"encoding/json"
	"errors"
	"flag"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/touchmarine/to/matcher"
	"github.com/touchmarine/to/node"
	"github.com/touchmarine/to/parser"
)

const testdata = "testdata"

// use go test ./parser -update to create/update the golden files
var update = flag.Bool("update", false, "update golden files")

func TestGolden(t *testing.T) {
	f, err := os.Open(testdata)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	entries, err := f.ReadDir(0)
	if err != nil {
		t.Fatal(err)
	}

	for _, e := range entries {
		if e.IsDir() {
			testDir(t, filepath.Join(testdata, e.Name()))
		}
	}
}

func testDir(t *testing.T, dir string) {
	ef, err := os.Open(filepath.Join(dir, "elements.json"))
	if err != nil {
		t.Fatal(err)
	}
	defer ef.Close()

	var elements parser.Elements
	if err := json.NewDecoder(ef).Decode(&elements); err != nil {
		t.Fatal(err)
	}

	inputs, err := filepath.Glob(filepath.Join(dir, "*.input"))
	if err != nil {
		t.Fatal(err)
	}

	for _, in := range inputs {
		basePath := in[:len(in)-len(".input")]

		t.Run(basePath[len(testdata)+1:], func(t *testing.T) {
			runTest(t, elements, basePath)
		})
	}
}

func runTest(t *testing.T, elements parser.Elements, testPath string) {
	bi, err := os.ReadFile(testPath + ".input")
	if err != nil {
		t.Fatal(err)
	}
	// \n is always added, don't know by what, but it isn't nice for testing
	// positions
	input := strings.TrimSuffix(string(bi), "\n")

	fs := flag.NewFlagSet("", flag.ContinueOnError)
	var printModes []string
	fs.Func("print-mode", "enable print tree mode (options: PrintData, PrintLocation)", func(s string) error {
		printModes = append(printModes, s)
		return nil
	})

	var src string
	const prefix = "//to:"
	if strings.HasPrefix(input, prefix) {
		var end int
		if i := strings.Index(input, "\n"); i > 0 {
			end = i
		} else {
			end = len(input)
		}
		if err := fs.Parse(strings.Split(input[len(prefix):end], " ")); err != nil {
			t.Fatal(err)
		}
		src = input[end+1:]
	} else {
		src = input
	}

	p := parser.Parser{
		Elements: elements,
		Matchers: matcher.Defaults(),
		TabWidth: 8,
	}
	nodes, err := p.Parse(strings.NewReader(src))
	testError(t, testPath, err)

	var m node.PrinterMode
	for _, s := range printModes {
		var mm node.PrinterMode
		if err := (&mm).UnmarshalText([]byte(s)); err != nil {
			t.Fatal(err)
		}
		m = m | mm // set flag
	}
	if len(printModes) == 0 && m&node.PrintData == 0 {
		// set as default because before print modes were added,
		// Data was always printed
		m = m | node.PrintData
	}
	var b strings.Builder
	if err := (node.Printer{m}).Fprint(&b, nodes); err != nil {
		t.Fatal(err)
	}
	res := b.String()

	goldenPath := testPath + ".golden"
	if *update {
		if err := os.WriteFile(goldenPath, []byte(res), 0644); err != nil {
			t.Fatal(err)
		}
	}

	bg, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatal(err)
	}
	golden := string(bg)

	if res != golden {
		t.Errorf("\nfrom input:\n%s\ngot:\n%s\nwant:\n%s", src, res, golden)
	}

}

func testError(t *testing.T, testPath string, err error) {
	errorPath := testPath + ".error"
	if _, statErr := os.Stat(errorPath); statErr == nil {
		// expected errors
		b, fileErr := os.ReadFile(errorPath)
		if fileErr != nil {
			t.Fatal(fileErr)
		}

		list, ok := err.(parser.ErrorList)
		if !ok {
			t.Fatalf("err not ErrorList (%T)", err)
		}

		expected := strings.Split(string(b), "\n")
		left := expected
		for _, e := range list {
			if contains(left, e.Error()) {
				left = left[1:]
			} else {
				t.Errorf("got error %q", e)
			}
		}

		// wanted errors that were not reported
		for _, e := range left {
			if e == "" {
				continue
			}

			t.Errorf("want error %q", e)
		}
	} else if errors.Is(statErr, fs.ErrNotExist) {
		// no expected errors
		if err != nil {
			t.Fatal(err)
		}
	} else {
		t.Fatal(statErr)
	}
}

func contains(p []string, s string) bool {
	for _, x := range p {
		if x == s {
			return true
		}
	}
	return false
}
