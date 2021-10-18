package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"os"

	"github.com/touchmarine/to/config"
	"github.com/touchmarine/to/node"
	"github.com/touchmarine/to/parser"
	"github.com/touchmarine/to/printer"
	totemplate "github.com/touchmarine/to/template"
	"github.com/touchmarine/to/transformer"
)

func main() {
	fmtCmd := flag.NewFlagSet("format", flag.ExitOnError)
	configPath := fmtCmd.String("config", "", "custom config")
	stringify := fmtCmd.Bool("stringify", false, "stringify")

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "no format")
		os.Exit(1)
	}

	fmtCmd.Parse(os.Args[2:])

	format := os.Args[1]

	var cfg *config.Config
	if *configPath != "" {
		f, err := os.Open(*configPath)
		if err != nil {
			log.Fatal(err)
		}

		if err := json.NewDecoder(f).Decode(&cfg); err != nil {
			log.Fatal(err)
		}
	} else {
		cfg = config.Default
	}

	root, err := parser.Parse(os.Stdin, cfg.ParserElements())
	if err != nil {
		parser.PrintError(os.Stderr, err)
		os.Exit(2)
	}

	root = transformer.Apply(root, cfg.DefaultTransformers())

	if *stringify {
		node.Fprint(os.Stdout, root)
	}

	if format == "fmt" {
		printer.Fprint(os.Stdout, cfg.PrinterElements(), root)
	} else {
		//aggregates := aggregator.Apply(root, cfg.DefaultAggregators())

		//data := map[string]interface{}{
		//	"aggregates": aggregates,
		//}

		tmpl := template.New(format)
		tmpl.Funcs(totemplate.Functions(tmpl))
		//tmpl.Funcs(totemplate.Functions)
		//tmpl.Funcs(totemplate.RenderFunctions(tmpl, data))
		template.Must(cfg.ParseTemplates(tmpl, format))
		if err := tmpl.ExecuteTemplate(os.Stdout, "root", root); err != nil {
			panic(err)
		}
	}
}
