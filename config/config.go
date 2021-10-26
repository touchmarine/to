package config

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"

	"github.com/touchmarine/to/aggregator"
	sequentialnumberag "github.com/touchmarine/to/aggregator/sequentialnumber"
	"github.com/touchmarine/to/node"
	"github.com/touchmarine/to/parser"
	"github.com/touchmarine/to/printer"
)

//go:embed to.json
var b []byte

var Default *Config

func init() {
	r := bytes.NewReader(b)
	if err := json.NewDecoder(r).Decode(&Default); err != nil {
		panic(err)
	}
}

type Config struct {
	Root struct {
		Templates map[string]string `json:"templates"`
	} `json:"root"`
	Elements   Elements   `json:"elements"`
	Groups     Groups     `json:"groups"`
	Aggregates Aggregates `json:"aggregates"`
}

// Elements maps Elements by name.
type Elements map[string]Element

type Element struct {
	Type        node.Type         `json:"type"`
	Delimiter   string            `json:"delimiter"`
	Matcher     string            `json:"matcher"`
	DoNotRemove bool              `json:"doNotRemove"`
	Templates   map[string]string `json:"templates"`
}

// Groups maps Groups by name.
type Groups map[string]Group

type Group struct {
	Type      string            `json:"type"`
	Element   string            `json:"element"`
	Target    string            `json:"target"`
	Option    string            `json:"option"`
	Templates map[string]string `json:"templates"`
}

type Aggregates struct {
	SequentialNumbers map[string]struct {
		Elements []string `json:"elements"`
	} `json:"sequentialNumbers"`
}

func (c Config) ParserElements() parser.Elements {
	return ToParserElements(c.Elements)
}

func (c Config) PrinterElements() printer.Elements {
	return ToPrinterElements(c.Elements)
}

func (c Config) GroupsByType(t string) Groups {
	m := Groups{}
	for name, g := range c.Groups {
		if g.Type == t {
			m[name] = g
		}
	}
	return m
}

func (c Config) ParseTemplates(target *template.Template, templateName string) (*template.Template, error) {
	rootTmpl, ok := c.Root.Templates[templateName]
	if !ok {
		return nil, fmt.Errorf("root %s template not found", templateName)
	}
	if _, err := target.New("root").Parse(rootTmpl); err != nil {
		return nil, err
	}

	for name, element := range c.Elements {
		tmpl, ok := element.Templates[templateName]
		if !ok {
			return nil, fmt.Errorf("element template not found (%s %s)", name, templateName)
		}
		if _, err := target.New(name).Parse(tmpl); err != nil {
			return nil, err
		}
	}

	for name, group := range c.Groups {
		tmpl, ok := group.Templates[templateName]
		if !ok {
			return nil, fmt.Errorf("group template not found (%s %s)", name, templateName)
		}
		if _, err := target.New(name).Parse(tmpl); err != nil {
			return nil, err
		}
	}

	return target, nil
}

func (c Config) DefaultAggregators() map[string]map[string]aggregator.Aggregator {
	m := map[string]map[string]aggregator.Aggregator{}
	for n, a := range c.Aggregates.SequentialNumbers {
		if m["sequentialNumbers"] == nil {
			m["sequentialNumbers"] = map[string]aggregator.Aggregator{}
		}

		m["sequentialNumbers"][n] = sequentialnumberag.Aggregator{a.Elements}
	}
	return m
}

func ToParserElements(elements Elements) parser.Elements {
	m := parser.Elements{}
	for name, e := range elements {
		m[name] = parser.Element{
			Name:      name,
			Type:      e.Type,
			Delimiter: e.Delimiter,
			Matcher:   e.Matcher,
		}
	}
	return m
}

func ToPrinterElements(elements Elements) printer.Elements {
	m := printer.Elements{}
	for name, e := range elements {
		m[name] = printer.Element{
			Name:        name,
			Type:        e.Type,
			Delimiter:   e.Delimiter,
			Matcher:     e.Matcher,
			DoNotRemove: e.DoNotRemove,
		}
	}
	return m
}
