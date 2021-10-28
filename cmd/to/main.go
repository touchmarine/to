package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"os"

	"github.com/touchmarine/to/aggregator"
	seqnumaggregator "github.com/touchmarine/to/aggregator/sequentialnumber"
	"github.com/touchmarine/to/config"
	"github.com/touchmarine/to/node"
	"github.com/touchmarine/to/parser"
	"github.com/touchmarine/to/printer"
	totemplate "github.com/touchmarine/to/template"
	"github.com/touchmarine/to/transformer"
	"github.com/touchmarine/to/transformer/group"
	"github.com/touchmarine/to/transformer/paragraph"
	"github.com/touchmarine/to/transformer/sequentialnumber"
	"github.com/touchmarine/to/transformer/sticky"
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

	root, err := parser.Parse(os.Stdin, cfg.Elements.ParserElements())
	if err != nil {
		parser.PrintError(os.Stderr, err)
		os.Exit(2)
	}

	var transformers []transformer.Transformer
	paragraphs := paragraph.Map{}
	lists := group.Map{}
	stickies := sticky.Map{}
	for n, g := range cfg.Groups {
		switch g.Type {
		case "paragraph":
			var t node.Type
			if err := (&t).UnmarshalText([]byte(g.Option)); err != nil {
				log.Fatal(err)
			}
			paragraphs[t] = n
		case "list":
			lists[g.Element] = n
		case "sticky":
			stickies[g.Element] = sticky.Sticky{
				Name:   n,
				Target: g.Target,
				After:  g.Option == "after",
			}
		default:
			fmt.Fprintf(os.Stderr, "unexpected group type %s\n", g.Type)
			os.Exit(1)
		}
	}
	transformers = append(transformers, paragraph.Transformer{paragraphs})
	transformers = append(transformers, group.Transformer{lists})
	transformers = append(transformers, sticky.Transformer{stickies})
	transformers = append(transformers, transformer.Func(sequentialnumber.Transform))
	transformer.Apply(root, transformers)

	if *stringify {
		node.Fprint(os.Stdout, root)
	}

	if format == "fmt" {
		printer.Fprint(os.Stdout, cfg.Elements.PrinterElements(), root)
	} else {
		aggregators := aggregator.Aggregators{}
		for n, a := range cfg.Aggregates {
			switch a.Type {
			case "sequentialNumber":
				aggregators[n] = seqnumaggregator.Aggregator{a.Elements}
			default:
				fmt.Fprintf(os.Stderr, "unexpected aggregate type %s\n", a.Type)
				os.Exit(1)
			}
		}
		aggregates := aggregator.Apply(root, aggregators)

		tmpl := template.New(format)
		global := map[string]interface{}{
			"aggregates": aggregates,
		}
		tmpl.Funcs(totemplate.Funcs(tmpl, global))
		template.Must(cfg.ParseTemplates(tmpl, format))
		if err := tmpl.ExecuteTemplate(os.Stdout, "root", root); err != nil {
			panic(err)
		}
	}
}
