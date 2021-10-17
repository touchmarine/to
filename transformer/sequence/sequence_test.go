package sequence_test

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/touchmarine/to/node"
	"github.com/touchmarine/to/parser"
	"github.com/touchmarine/to/transformer"
	"github.com/touchmarine/to/transformer/sequence"
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

func runTest(t *testing.T, elements parser.ElementMap, testPath string) {
	bi, err := os.ReadFile(testPath + ".to")
	if err != nil {
		t.Fatal(err)
	}
	input := string(bi)

	root, err := parser.Parse(strings.NewReader(input), elements)
	if err != nil {
		t.Fatal(err)
	}

	root = transformer.Apply(root, []transformer.Transformer{sequence.Transformer{}})

	res, err := node.StringifyDetailed(root)
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
