package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"os"

	"github.com/touchmarine/to/aggregator"
	"github.com/touchmarine/to/config"
	"github.com/touchmarine/to/node"
	"github.com/touchmarine/to/parser"
	"github.com/touchmarine/to/printer"
	totemplate "github.com/touchmarine/to/template"
	"github.com/touchmarine/to/transformer"
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

	var conf *config.Config
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

	root, err := parser.Parse(os.Stdin, conf.ParserElements())
	if err != nil {
		parser.PrintError(os.Stderr, err)
		os.Exit(2)
	}

	root = transformer.Apply(root, conf.DefaultTransformers())

	if *stringify {
		node.Fprint(os.Stdout, root)
	}

	if format == "fmt" {
		printer.Fprint(os.Stdout, conf.PrinterElements(), root)
	} else {
		aggregates := aggregator.Apply(root, conf.DefaultAggregators())

		data := map[string]interface{}{
			"aggregates": aggregates,
		}

		tmpl := template.New(format)
		tmpl.Funcs(totemplate.Functions)
		tmpl.Funcs(totemplate.RenderFunctions(tmpl, data))
		template.Must(conf.ParseTemplates(tmpl, format))
		if err := tmpl.ExecuteTemplate(os.Stdout, "root", root); err != nil {
			panic(err)
		}
	}
}
