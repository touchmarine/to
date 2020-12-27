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
	log.SetFlags(0)
	flag.Parse()

	b, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	p := parser.New(string(b), nil)
	doc, errCount := p.ParseDocument()
	if errCount > 0 {
		log.Fatalf("ParseDocument encountered %d errors", errCount)
	}

	if *html {
		fmt.Print(renderer.HTML(doc, 0))
		return
	}

	if *pretty {
		fmt.Print(printer.Pretty(doc, 0))
		return
	}
}
