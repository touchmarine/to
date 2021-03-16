package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"to/internal/config"
	"to/internal/node"
	"to/internal/parser"
	"to/internal/renderer"
)

var confPath = flag.String("conf", "", "custom config")

func main() {
	flag.Parse()

	var nodes []node.Block
	var perr []error

	if *confPath == "" {
		nodes, perr = parser.Parse(os.Stdin)
		if perr != nil {
			log.Fatal(perr)
		}
	} else {
		f, err := os.Open(*confPath)
		if err != nil {
			log.Fatal(err)
		}

		var conf config.Config
		if err := json.NewDecoder(f).Decode(&conf); err != nil {
			log.Fatal(err)
		}

		nodes, perr = parser.ParseCustom(os.Stdin, conf.Elements)
		if perr != nil {
			log.Fatal(perr)
		}
	}

	renderer.HTML(os.Stdout, node.BlocksToNodes(nodes))
}
