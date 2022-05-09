// Touch is a tool for managing Touch formatted text.
//
// Usage:
// 	to <command> [arguments]
//
// Commands:
// 	build  	convert Touch formatted text
// 	fmt    	format Touch formatted text (prettify)
// 	tree   	print node tree
// 	tool    run specified Touch tool
// 	help   	print help
// 	version	print version
//
// Use "to help <command>" for details about a command.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/cbroglie/mustache"
	"github.com/gosimple/slug"
	"github.com/touchmarine/to/config"
	"github.com/touchmarine/to/matcher"
	"github.com/touchmarine/to/node"
	"github.com/touchmarine/to/parser"
	"github.com/touchmarine/to/printer"
	totemplate "github.com/touchmarine/to/template"
	"github.com/touchmarine/to/tools/extjson"
	"github.com/touchmarine/to/transformer"
	"github.com/touchmarine/to/transformer/group"
	"github.com/touchmarine/to/transformer/paragraph"
	"github.com/touchmarine/to/transformer/sequentialnumber"
	"github.com/touchmarine/to/transformer/sticky"
)

func init() {
	slug.MaxLength = 20
}

const version = "1.0.0-beta.1"

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
			configs  string
			tabWidth int
		)
		registerWorkFlags := func(fs *flag.FlagSet) {
			fs.StringVar(&configs, "config", "", "comma-separated list of configs to use")
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

			if isStdinEmpty() {
				fmt.Fprintf(os.Stderr, strings.TrimSpace(`
to build: empty stdin

usage:   to build <format> [options] stdin
example: to build html < file.to
Run 'to help build' for details.
`)+"\n")
				os.Exit(2)
				return
			}

			cfg := &config.Default
			for _, p := range strings.Split(configs, ",") {
				if p == "" {
					continue
				}
				c := yamlDecodeConfigFile(p) // exits on error
				config.ShallowMerge(cfg, c)
			}
			src, err := io.ReadAll(os.Stdin)
			if err != nil {
				fmt.Fprintf(os.Stderr, "read stdint failed: %v\n", err)
				os.Exit(1)
				return
			}
			root := parse(src, cfg.Elements.ParserElements(), tabWidth)
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

			if isStdinEmpty() {
				fmt.Fprintf(os.Stderr, strings.TrimSpace(`
to fmt: empty stdin

usage:   to fmt [options] stdin
example: to fmt < file.to
Run 'to help fmt' for details.
`)+"\n")
				os.Exit(2)
				return
			}

			cfg := &config.Default
			for _, p := range strings.Split(configs, ",") {
				if p == "" {
					continue
				}
				c := yamlDecodeConfigFile(p) // exits on error
				config.ShallowMerge(cfg, c)
			}
			src, err := io.ReadAll(os.Stdin)
			if err != nil {
				fmt.Fprintf(os.Stderr, "read stdint failed: %v\n", err)
				os.Exit(1)
				return
			}
			root := parse(src, cfg.Elements.ParserElements(), tabWidth)
			root = transformers(cfg.Elements).Transform(root)

			format(cfg.Elements.ParserElements(), *lineLength, root) // exits on error
			return
		case "tree":
			fs := flag.NewFlagSet("to tree", flag.ContinueOnError)
			fs.Usage = func() {
				fmt.Fprintln(os.Stderr, strings.TrimSpace(`
usage: to tree [options] stdin
Run 'to help tree' for details.
`))
			}
			modes := fs.String("mode", "", "comma-separated list of modes to use")
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

			if isStdinEmpty() {
				fmt.Fprintf(os.Stderr, strings.TrimSpace(`
to tree: empty stdin

usage:   to tree [options] stdin
example: to tree < file.to
Run 'to help tree' for details.
`)+"\n")
				os.Exit(2)
				return
			}

			cfg := &config.Default
			for _, p := range strings.Split(configs, ",") {
				if p == "" {
					continue
				}
				c := yamlDecodeConfigFile(p) // exits on error
				config.ShallowMerge(cfg, c)
			}
			src, err := io.ReadAll(os.Stdin)
			if err != nil {
				fmt.Fprintf(os.Stderr, "read stdint failed: %v\n", err)
				os.Exit(1)
				return
			}
			root := parse(src, cfg.Elements.ParserElements(), tabWidth)
			root = transformers(cfg.Elements).Transform(root)

			var m []string
			if *modes != "" {
				m = strings.Split(*modes, ",")
			}
			tree(root, m) // exits on error
			return
		default:
			panic("unexpected cmd " + cmd)
		}
	case "tool":
		if len(args) == 0 {
			fmt.Println(strings.TrimSpace(`
to tool: missing <tool>
Run 'to help tool' for details.
`))
			return
		}

		cmd, args := args[0], args[1:]
		switch cmd {
		case "extjson":
			if len(args) > 0 {
				fmt.Fprintf(os.Stderr, strings.TrimSpace(`
to tool extjson: unexpected arguments: %s
Run 'to help tool extjson' for details.
`)+"\n", strings.Join(args, " "))
				os.Exit(2)
				return
			}

			if isStdinEmpty() {
				fmt.Fprintf(os.Stderr, strings.TrimSpace(`
to tool extjson: empty stdin

usage:   to tool extjson stdin
example: to tool extjson < config.extjson
Run 'to help tool extjson' for details.
`)+"\n")
				os.Exit(2)
				return
			}

			extjson.Convert(os.Stdout, os.Stdin)
		default:
			fmt.Fprintf(os.Stderr, strings.TrimSpace(`
to tool %s: unknown tool
Run 'to help tool'.
`)+"\n", cmd)
			os.Exit(2)
			return

		}
	case "help":
		if len(args) == 0 {
			help()
			return
		}

		cmd := args[0]
		switch cmd {
		case "build":
			fmt.Println(strings.TrimSpace(`
usage:   to build <format> [options] stdin
example: to build html < file.to

Build converts Touch formatted text to the given format.

Options:
	-config file,list
		a comma-separated list of configs to use. Configs are
		shallow merged (sequentially) into the default config.
		(Shallow merge adds or overrides only whole objects, it
		cannot override specific properties.)
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
	-config file,list
		a comma-separated list of configs to use. Configs are
		shallow merged (sequentially) into the default config.
		(Shallow merge adds or overrides only whole objects, it
		cannot override specific properties.)
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
	-config file,list
		a comma-separated list of configs to use. Configs are
		shallow merged (sequentially) into the default config.
		(Shallow merge adds or overrides only whole objects, it
		cannot override specific properties.)
	-tabwidth int
		tab=<tabwidth> x spaces (default=8)
	-mode   mode,list
		a comma-separated list of modes to use:
		printdata, printoffsets, printlocation
`))
			return
		case "tool":
			args := args[1:]
			if len(args) == 0 {
				fmt.Fprintln(os.Stdout, strings.TrimSpace(`
usage: to tool <tool> [arguments]

Tool runs the Touch tool.

Tools:
	extjson  convert extended JSON to plain JSON

Use "to help tool <tool>" for details about a tool.
`))
				return
			}

			cmd, args := args[0], args[1:]
			if len(args) > 0 {
				fmt.Fprintf(os.Stderr, strings.TrimSpace(`
to help tool %s: unexpected arguments: %s
Run 'to help tool'.
`)+"\n", cmd, strings.Join(args, " "))
				os.Exit(2)
				return
			}

			switch cmd {
			case "extjson":
				fmt.Println(strings.TrimSpace(`
extjson reads and converts extended JSON to plain JSON from stdin.

usage:   to tool extjson stdin
example: to tool extjson < config.extjson > config.json

Extended JSON is a superset of JSON and converts to plain JSON. It
makes it easier to write JSON Touch configs.

Features:
- raw multiline strings
	Raw multiline strings are delimited by triple single
	quotes and convert to regular JSON strings. Immediate
	newline after the delimiter is discarded if present.

	For example:
		"Templates": {
			"html": '''
		<blockquote>
			{{template "children" .}}
		</blockquote>
			'''
		}
	converts to:
		"Templates": {
			"html": "<blockquote>\n\t{{template \"children\" .}}\n</blockquote>\n"
		}
`))
				return
			default:
				fmt.Fprintf(os.Stderr, strings.TrimSpace(`
to help tool %s: unknown topic
Run 'to help tool'.
`)+"\n", strings.Join(args, " "))
				os.Exit(2)
				return
			}
		default:
			fmt.Fprintf(os.Stderr, strings.TrimSpace(`
to help %s: unknown topic
Run 'to help'.
`)+"\n", strings.Join(args, " "))
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
	tool    run specified Touch tool
	help   	print help
	version	print version

Use "to help <command>" for details about a command.
`))
}

func isStdinEmpty() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		panic(fmt.Sprintf("os.Stdin.Stat() failed: %v", err))
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}

func yamlDecodeConfigFile(path string) *config.Config {
	var c config.Config
	f, err := os.Open(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot open config file (%s): %v\n", path, err)
		os.Exit(2)
		return nil
	}
	dec := yaml.NewDecoder(f)
	dec.SetStrict(true)
	if err := dec.Decode(&c); err != nil {
		fmt.Fprintf(os.Stderr, "cannot decode YAML from config file (%s): %v\n", path, err)
		os.Exit(2)
		return nil
	}
	return &c
}

func parse(src []byte, elements parser.Elements, tabWidth int) *node.Node {
	p := parser.Parser{
		Elements: elements,
		Matchers: matcher.Defaults(),
	}
	if tabWidth > 0 {
		p.TabWidth = tabWidth
	} else {
		p.TabWidth = 8
	}
	root, err := p.Parse(nil, src)
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
		if e.Disabled {
			continue
		}
		var x node.Type
		if err := (&x).UnmarshalText([]byte(e.Type)); err == nil {
			// is a node element (can't be a group)
			continue
		}

		switch e.Type {
		case "template":
			// TODO: added for templates in config.elements?
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
			fmt.Fprintf(os.Stderr, "unsupported group type: %q (element=%q)\n", e.Type, n)
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

	tpl, err := elementTemplate(cfg, "root", format)
	if err != nil {
		fmt.Fprintf(os.Stderr, "no root template: %v", err)
		os.Exit(1)
		return
	}

	o, err := render(cfg, root, format, nil, tpl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "render failed: %v\n", err)
		os.Exit(1)
		return
	}
	io.WriteString(os.Stdout, o)
	return

	//aggregators := aggregator.Aggregators{}
	//for n, a := range cfg.Aggregates {
	//	switch a.Type {
	//	case "sequentialNumber":
	//		aggregators[n] = seqnumaggregator.Aggregator{a.Elements}
	//	default:
	//		fmt.Fprintf(os.Stderr, "invalid config: unsupported aggregate type: %q\n", a.Type)
	//		os.Exit(2)
	//		return
	//	}
	//}
	//aggregates := aggregator.Apply(root, aggregators)

	//tmpl := template.New(format)
	//global := map[string]any{
	//	"aggregates": aggregates,
	//}
	//tmpl.Funcs(totemplate.Funcs(tmpl, global))
	//_, err := cfg.ParseTemplates(tmpl, format)
	//if err != nil {
	//	fmt.Fprintf(os.Stderr, "parse templates failed (format=%q): %v\n", format, err)
	//	os.Exit(1)
	//	return
	//}
	//if err := tmpl.Execute(os.Stdout, root); err != nil {
	//	fmt.Fprintf(os.Stderr, "execute template failed: %v\n", err)
	//	os.Exit(1)
	//	return
	//}
}

// tpl overrides the default tepmlate found in config
func render(cfg *config.Config, n *node.Node, format string, attrs map[string]any, tpl string) (string, error) {
	if attrs == nil {
		attrs = map[string]any{}
	}

	// get element content—text value if text otherwise rendered children
	var cont string
	if n.Value != "" {
		cont = n.Value
	} else {
		var err error
		cont, err = renderChildren(cfg, n, format)
		if err != nil {
			return "", fmt.Errorf("renderChildren: %v", err)
		}
	}

	chldn, err := renderChildrenSeparate(cfg, n, format)
	if err != nil {
		return "", fmt.Errorf("renderChildrenSeparate: %v", err)
	}

	nd, err := extractNodeData(n.Data)
	if err != nil {
		return "", fmt.Errorf("extractNodeData: %v", err)
	}

	// add id attribute to all ranked hanging elements (they usually denote
	// sections) and to elements with an id attribute
	if v, ok := attrs["id"]; ok {
		var new string
		switch id := v.(type) {
		case string:
			// TODO: must start with a letter
			new = id
		default:
			return "", fmt.Errorf("attribute id is not a string (%T %s)", v, v)
		}
		attrs["id"] = new
	} else {
		if n.Type == node.TypeRankedHanging {
			id := slug.Make(n.TextContent())
			attrs["id"] = id
		}
	}

	// serialize attributes map to html-formatted attribute string
	attrStr := string(totemplate.AttributesToHTML(attrs))

	e := cfg.Elements[n.Element]

	// get sticky element and its target
	var st, tgt string
	if e.Type == sticky.Key {
		s, t, m, err := handleSticky(n, e)
		if err != nil {
			return "", fmt.Errorf("handleSticky: %v", err)
		}

		var stTpl string // override template
		if e.StickyTemplates != nil {
			stTpl = e.StickyTemplates[format]
		}
		st, err = render(cfg, s, format, copyMap(m), stTpl)
		if err != nil {
			return "", fmt.Errorf("render sticky: %v", err)
		}

		var tgtTpl string // override template
		if e.TargetTemplates != nil {
			tgtTpl = e.TargetTemplates[format]
		}
		tgt, err = render(cfg, t, format, copyMap(m), tgtTpl)
		if err != nil {
			return "", fmt.Errorf("render sticky target: %v", err)
		}
	}

	// render template
	m := map[string]any{
		"content":     cont,           // text value if text otherwise rendered children
		"children":    chldn,          // list of rendered element children
		"rank":        nd.rank,        // node's rank
		"openingText": nd.openingText, // node's opening text
		"seqNum":      nd.seqNum,      // node's sequential number
		"attributes":  attrStr,        // passed from parent (from sticky elements)
		"textContent": n.TextContent(),
		"sticky":      st,  // sticky element
		"target":      tgt, // the element onto which sticky sticks to
	}
	if tpl == "" {
		// no override template
		var err error
		tpl, err = elementTemplate(cfg, n.Element, format)
		if err != nil {
			return "", fmt.Errorf("elementTemplate: %v", err)
		}
	}
	out, err := mustache.Render(tpl, m)
	if err != nil {
		return "", fmt.Errorf("mustache render: %v", err)
	}
	return out, nil
}

type extractedData struct {
	rank, openingText, seqNum string
}

func extractNodeData(d node.Data) (extractedData, error) {
	var ed extractedData
	if v, ok := d[parser.KeyRank]; ok {
		switch r := v.(type) {
		case int:
			ed.rank = strconv.Itoa(r)
		case string:
			ed.rank = r
		default:
			return extractedData{}, fmt.Errorf("rank is neither int nor string (%T %s)", v, v)
		}
	}
	if v, ok := d[parser.KeyOpeningText]; ok {
		switch r := v.(type) {
		case string:
			ed.openingText = r
		default:
			return extractedData{}, fmt.Errorf("opening text is not a string (%T %s)", v, v)
		}
	}
	if v, ok := d[sequentialnumber.Key]; ok {
		switch r := v.(type) {
		case string:
			ed.seqNum = r
		default:
			return extractedData{}, fmt.Errorf("seqNum is not a string (%T %s)", v, v)
		}
	}
	return ed, nil
}

func copyMap(m map[string]any) map[string]any {
	if m == nil {
		return nil
	}
	n := make(map[string]any, len(m))
	for k, v := range m {
		n[k] = v
	}
	return n
}

func elementTemplate(cfg *config.Config, nm string, format string) (string, error) {
	e := cfg.Elements[nm]
	if e.Disabled {
		return "", nil
	}
	tpl, ok := e.Templates[format]
	if !ok {
		return "", fmt.Errorf("element template not found: element=%q format=%q", nm, format)
	}
	return tpl, nil
}

func handleSticky(n *node.Node, e config.Element) (*node.Node, *node.Node, map[string]any, error) {
	if e.Type != sticky.Key {
		return nil, nil, nil, fmt.Errorf("element not of type sticky")
	}

	var s, t *node.Node
	if e.Option == "after" {
		s = n.LastChild
		t = n.FirstChild
	} else {
		s = n.FirstChild
		t = n.LastChild
	}
	if s == nil {
		return nil, nil, nil, fmt.Errorf("nil sticky node")
	}
	if t == nil {
		return nil, nil, nil, fmt.Errorf("nil sticky target node")
	}

	var a *node.Node
	switch "attributes" {
	case s.Element:
		a = s
	case t.Element:
		a = t
	}
	var m map[string]any
	if a != nil {
		m = totemplate.ParseAttributes(a.TextContent())
	}
	return s, t, m, nil
}

func renderChildren(cfg *config.Config, n *node.Node, format string) (string, error) {
	chldn, err := renderChildrenSeparate(cfg, n, format)
	if err != nil {
		return "", fmt.Errorf("renderChildrenSeparate: %v", err)
	}
	return strings.Join(chldn, ""), nil
}

func renderChildrenSeparate(cfg *config.Config, n *node.Node, format string) ([]string, error) {
	chldn := totemplate.ElementChildren(n)
	p := make([]string, 0, len(chldn))
	for _, c := range chldn {
		o, err := render(cfg, c, format, nil, "")
		if err != nil {
			return nil, fmt.Errorf("render: %v", err)
		}
		p = append(p, o)
	}
	return p, nil
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

valid modes: printdata, printoffsets, printlocation

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
