package main

import (
	"encoding/json"
	"flag"
	"html/template"
	"log"
	"os"
	"to/internal/config"
	"to/internal/node"
	"to/internal/parser"
	"to/internal/renderer"
	"to/internal/transformer"
)

var confPath = flag.String("conf", "", "custom config")

func main() {
	flag.Parse()

	var (
		conf   *config.Config
		blocks []node.Block
		perr   []error
	)

	if *confPath == "" {
		conf = config.Default

		blocks, perr = parser.Parse(os.Stdin)
		if perr != nil {
			log.Fatal(perr)
		}
	} else {
		f, err := os.Open(*confPath)
		if err != nil {
			log.Fatal(err)
		}

		if err := json.NewDecoder(f).Decode(&conf); err != nil {
			log.Fatal(err)
		}

		blocks, perr = parser.ParseCustom(os.Stdin, conf.Elements)
		if perr != nil {
			log.Fatal(perr)
		}
	}

	//rndr := renderer.New(conf)
	//rndr.Render(os.Stdout, "html", node.BlocksToNodes(blocks))

	nodes := node.BlocksToNodes(blocks)
	nodes = transformer.Group(nodes)

	tmpl := template.New("html")
	rndr := renderer.New(conf, tmpl)

	tmpl.Funcs(renderer.FuncMap)
	tmpl.Funcs(rndr.FuncMap())

	template.Must(conf.ParseTemplates(tmpl, "html"))
	rndr.Render(os.Stdout, nodes)
}
