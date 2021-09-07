package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/touchmarine/to/aggregator"
	"github.com/touchmarine/to/config"
	"github.com/touchmarine/to/node"
	"github.com/touchmarine/to/parser"
	"github.com/touchmarine/to/printer"
	"github.com/touchmarine/to/renderer"
	"github.com/touchmarine/to/stringifier"
	"github.com/touchmarine/to/transformer"
	"html/template"
	"log"
	"os"
)

func main() {
	fmtCmd := flag.NewFlagSet("format", flag.ExitOnError)
	confPath := fmtCmd.String("conf", "", "custom config")
	stringify := fmtCmd.Bool("stringify", false, "stringify")

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "no format")
		os.Exit(1)
	}

	fmtCmd.Parse(os.Args[2:])

	format := os.Args[1]

	var (
		conf  *config.Config
		nodes []node.Node
	)

	if *confPath != "" {
		f, err := os.Open(*confPath)
		if err != nil {
			log.Fatal(err)
		}

		if err := json.NewDecoder(f).Decode(&conf); err != nil {
			log.Fatal(err)
		}
	} else {
		conf = config.Default
	}

	blocks, errs := parser.Parse(os.Stdin, conf.ParserElements())
	if errs != nil {
		log.Fatal(errs)
	}

	nodes = node.BlocksToNodes(blocks)
	nodes = transformer.Apply(nodes, transformer.Defaults(conf))

	if *stringify {
		stringifier.StringifyTo(os.Stdout, nodes...)
	}

	if format == "fmt" {
		printer.Fprint(os.Stdout, conf.PrinterElements(), nodes)
	} else {
		aggregates := aggregator.Aggregate(conf.Aggregates, nodes)

		data := map[string]interface{}{
			"Aggregates": aggregates,
		}

		tmpl := template.New(format)
		rndr := renderer.New(tmpl, data)

		tmpl.Funcs(renderer.FuncMap)
		tmpl.Funcs(rndr.FuncMap())

		template.Must(conf.ParseTemplates(tmpl, format))
		rndr.RenderWithCustomRoot(os.Stdout, nodes)
	}
}
