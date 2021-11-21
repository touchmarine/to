package group_test

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/touchmarine/to/matcher"
	"github.com/touchmarine/to/node"
	"github.com/touchmarine/to/parser"
	"github.com/touchmarine/to/transformer"
	"github.com/touchmarine/to/transformer/group"
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

	var elements parser.Elements
	if err := json.NewDecoder(ef).Decode(&elements); err != nil {
		t.Fatal(err)
	}

	inputs, err := filepath.Glob(filepath.Join(dir, "*.to"))
	if err != nil {
		t.Fatal(err)
	}

	for _, in := range inputs {
		basePath := in[:len(in)-len(".to")]

		t.Run(basePath[len(testdata)+1:], func(t *testing.T) {
			runTest(t, elements, basePath)
		})
	}
}

func runTest(t *testing.T, elements parser.Elements, testPath string) {
	bi, err := os.ReadFile(testPath + ".to")
	if err != nil {
		t.Fatal(err)
	}
	input := string(bi)

	p := parser.Parser{
		Elements: elements,
		Matchers: matcher.Defaults(),
		TabWidth: 8,
	}
	root, err := p.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}

	root = transformer.Group{group.Transformer{group.Map{
		"GA": "A",
		"GB": "D",
	}}}.Transform(root)

	var b strings.Builder
	if err := node.Fprint(&b, root); err != nil {
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
		t.Errorf("\nfrom input:\n%s\ngot:\n%s\nwant:\n%s", input, res, golden)
	}

}
