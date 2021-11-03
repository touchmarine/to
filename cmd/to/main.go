package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
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

const version = "1.0.0-beta"

func usage() {
	fmt.Fprintln(os.Stderr, "usage: to [options] format")
	fmt.Fprintln(os.Stderr, "Run 'to -help' for details.")
	os.Exit(2)
}

func main() {
	var (
		conf        = flag.String("config", "", "base configuration file")
		stringify   = flag.Bool("stringify", false, "print nodes to stdout (debugging)")
		showHelp    = flag.Bool("help", false, "print help")
		showVersion = flag.Bool("version", false, "print version")
	)
	var overrides []string
	flag.Func("config-override", "configuration files that override the base file", func(s string) error {
		overrides = append(overrides, s)
		return nil
	})
	flag.Usage = usage
	flag.Parse()

	if *showHelp {
		fmt.Fprint(os.Stdout, `usage: to [options] format

Touch converts Touch formatted text to the given format. It reads the
text from standard input and writes the converted text to standard
output.

Options:
	-config          base configuration file
	-config-override configuration files that override the base file
	-stringify       print nodes to stdout (debugging)
	-help            print help
	-version         print version
`)
		return
	}
	if *showVersion {
		fmt.Fprintf(os.Stdout, "to %s\n", version)
		return
	}

	args := flag.Args()
	if len(args) < 1 {
		usage()
		return
	}

	var cfg *config.Config
	if *conf != "" {
		f, err := os.Open(*conf)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot open config file (%s): %v\n", *conf, err)
			os.Exit(2)
		}
		if err := json.NewDecoder(f).Decode(&cfg); err != nil {
			fmt.Fprintf(os.Stderr, "cannot decode json from config file (%s): %v\n", *conf, err)
			os.Exit(2)
		}
	} else {
		cfg = config.Default
	}

	for _, p := range overrides {
		var o *config.Config
		f, err := os.Open(p)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot open config file (%s): %v\n", *conf, err)
			os.Exit(2)
		}
		if err := json.NewDecoder(f).Decode(&o); err != nil {
			fmt.Fprintf(os.Stderr, "cannot decode json from config file (%s): %v\n", *conf, err)
			os.Exit(2)
		}
		config.ShallowMerge(cfg, o)
	}

	root, err := parser.Parse(os.Stdin, cfg.Elements.ParserElements())
	if err != nil {
		parser.PrintError(os.Stderr, err)
		os.Exit(1)
	}

	var transformers []transformer.Transformer
	paragraphs := paragraph.Map{}
	lists := group.Map{}
	stickies := sticky.Map{}
	for n, e := range cfg.Elements {
		var x node.Type
		if err := (&x).UnmarshalText([]byte(e.Type)); err == nil {
			// is a node element (can't be a group)
			continue
		}

		switch e.Type {
		case "paragraph":
			var t node.Type
			if err := (&t).UnmarshalText([]byte(e.Option)); err != nil {
				fmt.Fprintf(os.Stderr, "invalid paragraph option (%s)\n", e.Option)
				os.Exit(2)
			}
			paragraphs[t] = n
		case "list":
			lists[e.Element] = n
		case "sticky":
			stickies[e.Element] = sticky.Sticky{
				Name:   n,
				Target: e.Target,
				After:  e.Option == "after",
			}
		default:
			fmt.Fprintf(os.Stderr, "unsupported group type (%s)\n", e.Type)
			os.Exit(2)
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

	if format := args[0]; format == "fmt" {
		printer.Fprint(os.Stdout, cfg.Elements.PrinterElements(), root)
	} else {
		aggregators := aggregator.Aggregators{}
		for n, a := range cfg.Aggregates {
			switch a.Type {
			case "sequentialNumber":
				aggregators[n] = seqnumaggregator.Aggregator{a.Elements}
			default:
				fmt.Fprintf(os.Stderr, "unsupported aggregate type (%s)\n", a.Type)
				os.Exit(2)
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
			fmt.Fprintf(os.Stderr, "execute template failed ('root'): %v\n", err)
			os.Exit(1)
		}
	}
}
