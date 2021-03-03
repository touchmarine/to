package main

import (
	"fmt"
	"log"
	"os"
	"to/internal/node"
	"to/internal/parser"
	"to/internal/renderer"
)

func main() {
	nodes, err := parser.Parse(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	html := renderer.Render(node.BlocksToNodes(nodes)...)
	fmt.Print(html)
}
