package config

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"

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
	Templates  Templates  `json:"templates"`
	Elements   Elements   `json:"elements"`
	Aggregates Aggregates `json:"aggregates"`
}

type Templates map[string]string

func (ts Templates) parse(t *template.Template, name, format string) (*template.Template, error) {
	s, ok := ts[format]
	if !ok {
		return nil, fmt.Errorf("template not found (name=%q format=%q)", name, format)
	}
	if _, err := t.New(name).Parse(s); err != nil {
		return nil, err
	}
	return t, nil
}

// ParseTemplates parses all config templates as template bodies for the given
// template. It parses templates that match the given format.
func (c Config) ParseTemplates(t *template.Template, format string) (*template.Template, error) {
	if _, err := c.Templates.parse(t, "root", format); err != nil {
		return nil, err
	}
	if _, err := c.Elements.parseTemplates(t, format); err != nil {
		return nil, err
	}
	return t, nil
}

// Elements maps Elements by name.
type Elements map[string]Element

type Element struct {
	Type        string    `json:"type"`
	Delimiter   string    `json:"delimiter"`
	Matcher     string    `json:"matcher"`
	DoNotRemove bool      `json:"doNotRemove"`
	Element     string    `json:"element"`
	Target      string    `json:"target"`
	Option      string    `json:"option"`
	Templates   Templates `json:"templates"`
}

// ParserElements returns the elements converted to parser.Elements.
func (es Elements) ParserElements() parser.Elements {
	m := parser.Elements{}
	for n, e := range es {
		var t node.Type
		if err := (&t).UnmarshalText([]byte(e.Type)); err != nil {
			// isn't a node element
			continue
		}
		m[n] = parser.Element{
			Name:      n,
			Type:      t,
			Delimiter: e.Delimiter,
			Matcher:   e.Matcher,
		}
	}
	return m
}

// PrinterElements returns the elements converted to printer.Elements.
func (es Elements) PrinterElements() printer.Elements {
	m := printer.Elements{}
	for n, e := range es {
		var t node.Type
		if err := (&t).UnmarshalText([]byte(e.Type)); err != nil {
			// isn't a node element
			continue
		}
		m[n] = printer.Element{
			Name:        n,
			Type:        t,
			Delimiter:   e.Delimiter,
			Matcher:     e.Matcher,
			DoNotRemove: e.DoNotRemove,
		}
	}
	return m
}

func (es Elements) parseTemplates(t *template.Template, format string) (*template.Template, error) {
	for n, e := range es {
		if _, err := e.Templates.parse(t, n, format); err != nil {
			return nil, err
		}
	}
	return t, nil
}

// Aggregates maps Aggregates by name.
type Aggregates map[string]Aggregate

type Aggregate struct {
	Type     string   `json:"type"`
	Elements []string `json:"elements"`
}
