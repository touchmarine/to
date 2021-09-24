package paragraph_test

import (
	"encoding/json"
	"errors"
	"flag"
	"github.com/touchmarine/to/node"
	"github.com/touchmarine/to/parser"
	"github.com/touchmarine/to/transformer"
	"github.com/touchmarine/to/transformer/paragraph"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testdata = "testdata"

// use go test -update to create/update the golden files
var update = flag.Bool("update", false, "update golden files")

func TestGolden(t *testing.T) {
	testDir(t, testdata)
}

func testDir(t *testing.T, dir string) {
	ef, err := os.Open(filepath.Join(dir, "elements.json"))
	if err != nil {
		t.Fatal(err)
	}
	defer ef.Close()

	var elements parser.ElementMap
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

func runTest(t *testing.T, elements parser.ElementMap, testPath string) {
	bi, err := os.ReadFile(testPath + ".input")
	if err != nil {
		t.Fatal(err)
	}
	input := string(bi)

	root, errs := parser.Parse(strings.NewReader(input), elements)
	testErrors(t, testPath, errs)

	root = transformer.Apply(root, []transformer.Transformer{paragraph.NewTransformer("GP")})

	res, err := node.Stringify(root)
	if err != nil {
		t.Fatal(err)
	}

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
		t.Errorf("\nfrom input:\n%s\ngot:\n%s\nwant:\n%s", input, res, golden)
	}

}

func testErrors(t *testing.T, testPath string, errs []error) {
	errorPath := testPath + ".error"
	if _, err := os.Stat(errorPath); err == nil {
		// expected errors
		b, err := os.ReadFile(errorPath)
		if err != nil {
			t.Fatal(err)
		}

		expected := strings.Split(string(b), "\n")
		left := expected
		for _, e := range errs {
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
	} else if errors.Is(err, fs.ErrNotExist) {
		// no expected errors
		for _, e := range errs {
			t.Errorf("got error %q", e)
		}
	} else {
		t.Fatal(err)
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
