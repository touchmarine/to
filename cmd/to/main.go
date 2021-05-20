package main

import (
	"encoding/json"
	"flag"
	"github.com/touchmarine/to/internal/aggregator"
	"github.com/touchmarine/to/internal/config"
	"github.com/touchmarine/to/internal/node"
	"github.com/touchmarine/to/internal/parser"
	"github.com/touchmarine/to/internal/renderer"
	"github.com/touchmarine/to/internal/stringifier"
	"github.com/touchmarine/to/internal/transformer"
	"html/template"
	"log"
	"os"
)

var (
	confPath  = flag.String("conf", "", "custom config")
	stringify = flag.Bool("stringify", false, "stringify")
)

func main() {
	flag.Parse()

	var (
		conf  *config.Config
		nodes []node.Node
	)

	if *confPath == "" {
		conf = config.Default

		blocks, perr := parser.Parse(os.Stdin)
		if perr != nil {
			log.Fatal(perr)
		}

		nodes = node.BlocksToNodes(blocks)
	} else {
		f, err := os.Open(*confPath)
		if err != nil {
			log.Fatal(err)
		}

		if err := json.NewDecoder(f).Decode(&conf); err != nil {
			log.Fatal(err)
		}

		blocks, perr := parser.ParseCustom(os.Stdin, conf.Elements)
		if perr != nil {
			log.Fatal(perr)
		}

		nodes = node.BlocksToNodes(blocks)
	}

	nodes = transformer.Paragraph(nodes)
	nodes = transformer.Group(conf.Groups, nodes)
	nodes = transformer.Sequence(conf.Elements, nodes)

	if *stringify {
		stringifier.StringifyTo(os.Stdout, nodes...)
	} else {
		aggregates := aggregator.Aggregate(config.Default.Aggregates, nodes)

		data := map[string]interface{}{
			"Aggregates": aggregates,
		}

		tmpl := template.New("html")
		rndr := renderer.New(tmpl, data)

		tmpl.Funcs(renderer.FuncMap)
		tmpl.Funcs(rndr.FuncMap())

		template.Must(conf.ParseTemplates(tmpl, "html"))
		rndr.Render(os.Stdout, nodes)
	}
}
