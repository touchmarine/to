package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"to/parser"
	"to/printer"
	"to/renderer"
)

var (
	html   = flag.Bool("html", false, "export HTML")
	pretty = flag.Bool("pretty", false, "pretty-print parse tree")
)

func main() {
	flag.Parse()

	b, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	p := parser.New(string(b))
	doc := p.ParseDocument()

	if *html {
		fmt.Print(renderer.HTML(doc, 0))
		return
	}

	if *pretty {
		fmt.Print(printer.Pretty(doc, 0))
		return
	}
}
