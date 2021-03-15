package main

import (
	"fmt"
	"log"
	"os"
	"to/internal/node"
	"to/internal/parser"
	"to/internal/renderer"
	"flag"
)

var tmpl = flag.Bool("tmpl", false, "use HTML templates")

func main() {
	flag.Parse()

	nodes, err := parser.Parse(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	if *tmpl {
		renderer.HTML(os.Stdout, node.BlocksToNodes(nodes))
		return
	}

	html := renderer.Render(node.BlocksToNodes(nodes)...)
	fmt.Print(html)
}
