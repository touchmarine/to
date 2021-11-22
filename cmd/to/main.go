package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"os"
	"strings"

	"github.com/touchmarine/to/aggregator"
	seqnumaggregator "github.com/touchmarine/to/aggregator/sequentialnumber"
	"github.com/touchmarine/to/config"
	"github.com/touchmarine/to/matcher"
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

func main() {
	rootFlags := flag.NewFlagSet("to", flag.ContinueOnError)
	rootFlags.Usage = usage
	if err := rootFlags.Parse(os.Args[1:]); err != nil {
		// By default, Parse prints usage and returns flag.ErrHelp on
		// -h/-help. However, we don't want this behaviour as Parse
		// already prints usage (in which we show how to get help) on
		// any error.
		os.Exit(2)
		return
	}
	// get non-flag arguments; it is considered a flag only if it's before
	// any non-flag arguments, e.g. `-b` in `a -b` is not a flag
	args := rootFlags.Args()
	if len(args) == 0 {
		// no command
		usage()
		os.Exit(2)
		return
	}
	cmd, args := args[0], args[1:]

	switch cmd {
	case "build", "fmt", "tree":
		var (
			configs  []string
			tabWidth int
		)
		registerWorkFlags := func(fs *flag.FlagSet) {
			fs.Func("config", "config files (shallow merged into the core config)", func(c string) error {
				configs = append(configs, c)
				return nil
			})
			fs.IntVar(&tabWidth, "tabwidth", 0, "tab=tabwidth x spaces") // default set in parse()
		}

		switch cmd {
		case "build":
			if len(args) < 1 {
				fmt.Fprintln(os.Stderr, strings.TrimSpace(`
to build: missing <format>

usage:   to build <format> [options] stdin
example: to build html < file.to
Run 'to help build' for details.
`))
				os.Exit(2)
				return
			}
			format, args := args[0], args[1:]

			fs := flag.NewFlagSet("to build", flag.ContinueOnError)
			fs.Usage = func() {
				fmt.Fprintln(os.Stderr, strings.TrimSpace(`
usage: to build <format> [options] stdin
Run 'to help build' for details.
`))
			}
			registerWorkFlags(fs)
			if err := fs.Parse(args); err != nil {
				os.Exit(2)
				return
			}
			args = fs.Args()
			if len(args) > 0 {
				fmt.Fprintf(os.Stderr, strings.TrimSpace(`
to build %s: unexpected arguments: %s
Run 'to help build' for details.
`)+"\n", format, strings.Join(args, " "))
				os.Exit(2)
				return
			}

			if isEmptyStdin() {
				fmt.Fprintf(os.Stderr, strings.TrimSpace(`
to build: empty stdin

usage:   to build <format> [options] stdin
example: to build html < file.to
Run 'to help build' for details.
`)+"\n")
				os.Exit(2)
				return
			}

			cfg := config.Default
			shallowMergeConfigs(cfg, configs)
			root := parse(os.Stdin, cfg.Elements.ParserElements(), tabWidth)
			root = transformers(cfg.Elements).Transform(root)

			build(cfg, root, format) // exits on error
			return
		case "fmt":
			fs := flag.NewFlagSet("to fmt", flag.ContinueOnError)
			fs.Usage = func() {
				fmt.Fprintln(os.Stderr, strings.TrimSpace(`
usage: to fmt [options] stdin
Run 'to help fmt' for details.
`))
			}
			lineLength := fs.Int("linelength", 0, "prose line length (hard-wrap)")
			registerWorkFlags(fs)
			if err := fs.Parse(args); err != nil {
				os.Exit(2)
				return
			}
			args := fs.Args()
			if len(args) > 0 {
				fmt.Fprintf(os.Stderr, strings.TrimSpace(`
to fmt: unexpected arguments: %s
Run 'to help fmt' for details.
`)+"\n", strings.Join(args, " "))
				os.Exit(2)
				return
			}

			if isEmptyStdin() {
				fmt.Fprintf(os.Stderr, strings.TrimSpace(`
to fmt: empty stdin

usage:   to fmt [options] stdin
example: to fmt < file.to
Run 'to help fmt' for details.
`)+"\n")
				os.Exit(2)
				return
			}

			cfg := config.Default
			shallowMergeConfigs(cfg, configs)
			root := parse(os.Stdin, cfg.Elements.ParserElements(), tabWidth)
			root = transformers(cfg.Elements).Transform(root)

			format(cfg.Elements.ParserElements(), *lineLength, root) // exits on error
			return
		case "tree":
			var modes []string
			fs := flag.NewFlagSet("to tree", flag.ContinueOnError)
			fs.Usage = func() {
				fmt.Fprintln(os.Stderr, strings.TrimSpace(`
usage: to tree [options] stdin
Run 'to help tree' for details.
`))
			}
			fs.Func("mode", "set print mode (modes: printdata, printlocation)", func(s string) error {
				modes = append(modes, s)
				return nil
			})
			registerWorkFlags(fs)
			if err := fs.Parse(args); err != nil {
				os.Exit(2)
				return
			}
			args := fs.Args()
			if len(args) > 0 {
				fmt.Fprintf(os.Stderr, strings.TrimSpace(`
to tree: unexpected arguments: %s
Run 'to help tree' for details.
`)+"\n", strings.Join(args, " "))
				os.Exit(2)
				return
			}

			if isEmptyStdin() {
				fmt.Fprintf(os.Stderr, strings.TrimSpace(`
to tree: empty stdin

usage:   to tree [options] stdin
example: to tree < file.to
Run 'to help tree' for details.
`)+"\n")
				os.Exit(2)
				return
			}

			cfg := config.Default
			shallowMergeConfigs(cfg, configs)
			root := parse(os.Stdin, cfg.Elements.ParserElements(), tabWidth)
			root = transformers(cfg.Elements).Transform(root)

			tree(root, modes) // exits on error
			return
		default:
			panic("unexpected cmd " + cmd)
		}
	case "help":
		if len(args) == 0 {
			help()
			return
		}

		cmd := args[0]
		allArgs := strings.Join(args, " ")
		switch cmd {
		case "build":
			fmt.Println(strings.TrimSpace(`
usage:   to build <format> [options] stdin
example: to build html < file.to

Build converts Touch formatted text to the given format.

Options:
	-config file
		configures templates and elements (sequentially shallow
		merged into the core config)
	-tabwidth int
		tab=<tabwidth> x spaces (default=8)
`))
			return
		case "fmt":
			fmt.Println(strings.TrimSpace(`
usage:           to fmt [options] stdin
format in place: to fmt < file.to 1<> file.to

Fmt formats Touch formatted text into its canonical form. Fmt is like
what is commonly known as prettify, but opinionated.

Options:
	-config file
		configures templates and elements (sequentially shallow
		merged into the core config)
	-tabwidth int
		tab=<tabwidth> x spaces (default=8)
	-linelength int
		hard-wrap prose at <linelength> column (default=0)
`))
			return
		case "tree":
			fmt.Println(strings.TrimSpace(`
usage:   to tree [options] stdin
example: to tree -mode printdata < file.to

Tree prints the node tree representation of Touch formatted text.

Options:
	-config file
		configures templates and elements (sequentially shallow
		merged into the core config)
	-tabwidth int
		tab=<tabwidth> x spaces (default=8)
	-mode
		dials the level of info to print

		modes: printData, printlocation
`))
			return
		default:
			fmt.Fprintf(os.Stderr, strings.TrimSpace(`
to help %s: unknown topic
Run 'to help'.
`)+"\n", allArgs)
			os.Exit(2)
			return
		}
	case "version":
		if len(args) > 0 {
			fmt.Fprintf(os.Stderr, strings.TrimSpace(`
to version: unexpected arguments: %s
Run 'to version'.
`)+"\n", strings.Join(args, " "))
			os.Exit(2)
			return
		}
		fmt.Printf("to %s\n", version)
		return
	default:
		fmt.Fprintf(os.Stderr, strings.TrimSpace(`
to %s: unknown command
Run 'to help %s' for details.
`)+"\n", cmd, cmd)
		os.Exit(2)
		return
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, strings.TrimSpace(`
usage: to <command> [arguments]
Run 'to help' for details.
`))
}

func help() {
	fmt.Fprintln(os.Stdout, strings.TrimSpace(`
Touch is a tool for managing Touch formatted text.

usage: to <command> [arguments]

Commands:
	build  	convert Touch formatted text
	fmt    	format Touch formatted text (prettify)
	tree   	print node tree
	help   	print help
	version	print version

Use "to help <command>" for details about a command.
`))
}

func isEmptyStdin() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		panic(fmt.Sprintf("os.Stdin.Stat() failed: %v", err))
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}

func shallowMergeConfigs(dst *config.Config, srcs []string) {
	for _, s := range srcs {
		var o *config.Config
		f, err := os.Open(s)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot open config file (%s): %v\n", s, err)
			os.Exit(2)
			return
		}
		if err := json.NewDecoder(f).Decode(&o); err != nil {
			fmt.Fprintf(os.Stderr, "cannot decode JSON from config file (%s): %v\n", s, err)
			os.Exit(2)
			return
		}
		_ = config.ShallowMerge(dst, o)
	}
}

func parse(in io.Reader, elements parser.Elements, tabWidth int) *node.Node {
	p := parser.Parser{
		Elements: elements,
		Matchers: matcher.Defaults(),
	}
	if tabWidth > 0 {
		p.TabWidth = tabWidth
	} else {
		p.TabWidth = 8
	}
	root, err := p.Parse(in)
	if err != nil {
		parser.PrintError(os.Stderr, err)
		os.Exit(1)
		return nil
	}
	return root
}

func transformers(elements config.Elements) transformer.Group {
	paragraphs := paragraph.Map{}
	lists := group.Map{}
	stickies := sticky.Map{}
	for n, e := range elements {
		var x node.Type
		if err := (&x).UnmarshalText([]byte(e.Type)); err == nil {
			// is a node element (can't be a group)
			continue
		}

		switch e.Type {
		case "paragraph":
			var t node.Type
			if err := (&t).UnmarshalText([]byte(e.Option)); err != nil {
				fmt.Fprintf(os.Stderr, "invalid paragraph option (%q)\n", e.Option)
				os.Exit(2)
				return transformer.Group{}
			}
			paragraphs[n] = t
		case "list":
			lists[n] = e.Element
		case "sticky":
			stickies[n] = sticky.Sticky{
				Element: e.Element,
				Target:  e.Target,
				After:   e.Option == "after",
			}
		default:
			fmt.Fprintf(os.Stderr, "unsupported group type (%q)\n", e.Type)
			os.Exit(2)
			return transformer.Group{}
		}
	}
	return transformer.Group{
		paragraph.Transformer{paragraphs},
		group.Transformer{lists},
		sticky.Transformer{stickies},
		transformer.Func(sequentialnumber.Transform),
	}
}

func build(cfg *config.Config, root *node.Node, format string) {
	aggregators := aggregator.Aggregators{}
	for n, a := range cfg.Aggregates {
		switch a.Type {
		case "sequentialNumber":
			aggregators[n] = seqnumaggregator.Aggregator{a.Elements}
		default:
			fmt.Fprintf(os.Stderr, "invalid config: unsupported aggregate type: %q\n", a.Type)
			os.Exit(2)
			return
		}
	}
	aggregates := aggregator.Apply(root, aggregators)

	tmpl := template.New(format)
	global := map[string]interface{}{
		"aggregates": aggregates,
	}
	tmpl.Funcs(totemplate.Funcs(tmpl, global))
	_, err := cfg.ParseTemplates(tmpl, format)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse templates failed (format=%q): %v\n", format, err)
		os.Exit(1)
		return
	}
	if err := tmpl.ExecuteTemplate(os.Stdout, "root", root); err != nil {
		fmt.Fprintf(os.Stderr, "execute template failed (\"root\"): %v\n", err)
		os.Exit(1)
		return
	}
}

func format(elements parser.Elements, lineLength int, root *node.Node) {
	if err := (printer.Printer{Elements: elements, LineLength: lineLength}).Fprint(os.Stdout, root); err != nil {
		fmt.Fprintf(os.Stderr, "fmt failed: %v\n", err)
		os.Exit(1)
		return
	}
}

func tree(root *node.Node, modes []string) {
	var m node.PrinterMode
	for _, s := range modes {
		var mm node.PrinterMode
		if err := (&mm).UnmarshalText([]byte(s)); err != nil {
			fmt.Fprintf(os.Stderr, strings.TrimSpace(`
to tree: invalid mode: %q

valid modes: printdata, printlocation

usage:   to tree [options] stdin
example: to tree -mode printdata < file.to
Run 'to help tree' for details.
`)+"\n", s)
			os.Exit(2)
			return
		}
		m = m | mm // set flag
	}
	if err := (node.Printer{m}).Fprint(os.Stdout, root); err != nil {
		fmt.Fprintf(os.Stderr, "print tree failed: %v\n", err)
		os.Exit(1)
		return
	}
}
