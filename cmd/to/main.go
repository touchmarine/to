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

	root, err := parser.Parse(os.Stdin, cfg.ParserElements())
	if err != nil {
		parser.PrintError(os.Stderr, err)
		os.Exit(2)
	}

	//root = transformer.Apply(root, cfg.DefaultTransformers())

	var transformers []transformer.Transformer

	paragraphGroups := cfg.GroupsByType("paragraph")
	paragraphMap := paragraph.Map{}
	for name, g := range paragraphGroups {
		var t node.Type
		if err := (&t).UnmarshalText([]byte(g.Option)); err != nil {
			log.Fatal(err)
		}
		paragraphMap[t] = name
	}
	transformers = append(transformers, paragraph.Transformer{paragraphMap})

	listGroups := cfg.GroupsByType("list")
	listMap := group.Map{}
	for name, g := range listGroups {
		listMap[g.Element] = group.Group{
			Name:    name,
			Element: g.Element,
		}
	}
	transformers = append(transformers, group.Transformer{listMap})

	stickyGroups := cfg.GroupsByType("sticky")
	//stickyMap = config.ToStickyMap(stickyGroups)
	stickyMap := sticky.Map{}
	for name, g := range stickyGroups {
		stickyMap[g.Element] = sticky.Sticky{
			Name:    name,
			Element: g.Element,
			Target:  g.Target,
			After:   g.Option == "after",
		}
	}
	transformers = append(transformers, sticky.Transformer{stickyMap})

	transformers = append(transformers, transformer.Func(sequentialnumber.Transform))

	transformer.Apply(root, transformers)

	if *stringify {
		node.Fprint(os.Stdout, root)
	}

	if format == "fmt" {
		printer.Fprint(os.Stdout, cfg.PrinterElements(), root)
	} else {
		aggregates := aggregator.Apply(root, cfg.DefaultAggregators())
		global := map[string]interface{}{
			"aggregates": aggregates,
		}

		tmpl := template.New(format)
		tmpl.Funcs(totemplate.Funcs(tmpl, global))
		//tmpl.Funcs(totemplate.Functions)
		//tmpl.Funcs(totemplate.RenderFunctions(tmpl, data))
		template.Must(cfg.ParseTemplates(tmpl, format))
		if err := tmpl.ExecuteTemplate(os.Stdout, "root", root); err != nil {
			panic(err)
		}
	}
}
