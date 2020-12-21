package printer_test

import (
	"testing"
	"to/parser"
	"to/printer"
)

func TestPretty(t *testing.T) {
	input := `
paragraph **strong** __emphasis__

= Heading
## Numbered

<Mammals></type/mammals>

` + "``" + `to
1. list item
` + "``" + `

- Tuesday:
	- milk
	- sugar
`

	want := `Document{
.   Children: [
.   .   Paragraph{
.   .   .   Lines: [
.   .   .   .   Line{
.   .   .   .   .   Children: [
.   .   .   .   .   .   "paragraph ",
.   .   .   .   .   .   Strong{
.   .   .   .   .   .   .   Children: [
.   .   .   .   .   .   .   .   "strong",
.   .   .   .   .   .   .   ],
.   .   .   .   .   .   },
.   .   .   .   .   .   " ",
.   .   .   .   .   .   Emphasis{
.   .   .   .   .   .   .   Children: [
.   .   .   .   .   .   .   .   "emphasis",
.   .   .   .   .   .   .   ],
.   .   .   .   .   .   },
.   .   .   .   .   ],
.   .   .   .   },
.   .   .   ],
.   .   },
.   .   Heading{
.   .   .   Children: [
.   .   .   .   "Heading",
.   .   .   ],
.   .   .   IsNumbered: false,
.   .   .   Level: 1,
.   .   },
.   .   Heading{
.   .   .   Children: [
.   .   .   .   "Numbered",
.   .   .   ],
.   .   .   IsNumbered: true,
.   .   .   Level: 2,
.   .   },
.   .   Paragraph{
.   .   .   Lines: [
.   .   .   .   Line{
.   .   .   .   .   Children: [
.   .   .   .   .   .   Link{
.   .   .   .   .   .   .   Children: [
.   .   .   .   .   .   .   .   "Mammals",
.   .   .   .   .   .   .   ],
.   .   .   .   .   .   .   Destination: "/type/mammals",
.   .   .   .   .   .   },
.   .   .   .   .   ],
.   .   .   .   },
.   .   .   ],
.   .   },
.   .   CodeBlock{
.   .   .   Body: "1. list item
",
.   .   .   Filename: "",
.   .   .   Language: "to",
.   .   .   MetadataRaw: "to",
.   .   },
.   .   List{
.   .   .   ListItems: [
.   .   .   .   ListItem{
.   .   .   .   .   Children: [
.   .   .   .   .   .   Lines[
.   .   .   .   .   .   .   Line{
.   .   .   .   .   .   .   .   Children: [
.   .   .   .   .   .   .   .   .   "Tuesday:",
.   .   .   .   .   .   .   .   ],
.   .   .   .   .   .   .   },
.   .   .   .   .   .   ],
.   .   .   .   .   .   List{
.   .   .   .   .   .   .   ListItems: [
.   .   .   .   .   .   .   .   ListItem{
.   .   .   .   .   .   .   .   .   Children: [
.   .   .   .   .   .   .   .   .   .   Lines[
.   .   .   .   .   .   .   .   .   .   .   Line{
.   .   .   .   .   .   .   .   .   .   .   .   Children: [
.   .   .   .   .   .   .   .   .   .   .   .   .   "milk",
.   .   .   .   .   .   .   .   .   .   .   .   ],
.   .   .   .   .   .   .   .   .   .   .   },
.   .   .   .   .   .   .   .   .   .   ],
.   .   .   .   .   .   .   .   .   ],
.   .   .   .   .   .   .   .   },
.   .   .   .   .   .   .   .   ListItem{
.   .   .   .   .   .   .   .   .   Children: [
.   .   .   .   .   .   .   .   .   .   Lines[
.   .   .   .   .   .   .   .   .   .   .   Line{
.   .   .   .   .   .   .   .   .   .   .   .   Children: [
.   .   .   .   .   .   .   .   .   .   .   .   .   "sugar",
.   .   .   .   .   .   .   .   .   .   .   .   ],
.   .   .   .   .   .   .   .   .   .   .   },
.   .   .   .   .   .   .   .   .   .   ],
.   .   .   .   .   .   .   .   .   ],
.   .   .   .   .   .   .   .   },
.   .   .   .   .   .   .   ],
.   .   .   .   .   .   .   Type: UnorderedList,
.   .   .   .   .   .   },
.   .   .   .   .   ],
.   .   .   .   },
.   .   .   ],
.   .   .   Type: UnorderedList,
.   .   },
.   ],
}`

	p := parser.New(input)
	doc := p.ParseDocument()

	if got := printer.Pretty(doc, 0); got != want {
		t.Errorf("\ngot:\n%s\nwant:\n%s", got, want)
	}
}
