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
	"github.com/touchmarine/to/stringifier"
	totemplate "github.com/touchmarine/to/template"
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
	nodes = transformer.Apply(nodes, conf.DefaultTransformers())

	if *stringify {
		stringifier.StringifyTo(os.Stdout, nodes...)
	}

	if format == "fmt" {
		printer.Fprint(os.Stdout, conf.PrinterElements(), nodes)
	} else {
		aggregates := aggregator.Apply(nodes, conf.DefaultAggregators())

		data := map[string]interface{}{
			"aggregates": aggregates,
		}

		tmpl := template.New(format)
		tmpl.Funcs(totemplate.Functions)
		tmpl.Funcs(totemplate.RenderFunctions(tmpl, data))
		template.Must(conf.ParseTemplates(tmpl, format))
		if err := tmpl.ExecuteTemplate(os.Stdout, "root", nodes); err != nil {
			panic(err)
		}
	}
}
