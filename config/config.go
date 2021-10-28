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
	Root       Root       `json:"root"`
	Elements   Elements   `json:"elements"`
	Groups     Groups     `json:"groups"`
	Aggregates Aggregates `json:"aggregates"`
}

// ParseTemplates parses root, element, and group templates as template bodies
// for t. It parses templates that match the given name.
func (c Config) ParseTemplates(t *template.Template, name string) (*template.Template, error) {
	if _, err := c.Root.parseTemplates(t, name); err != nil {
		return nil, err
	}
	if _, err := c.Elements.parseTemplates(t, name); err != nil {
		return nil, err
	}
	if _, err := c.Groups.parseTemplates(t, name); err != nil {
		return nil, err
	}
	return t, nil
}

// Root holds data related to the root node.
type Root struct {
	Templates map[string]string `json:"templates"`
}

func (r Root) parseTemplates(t *template.Template, name string) (*template.Template, error) {
	s, ok := r.Templates[name]
	if !ok {
		return nil, fmt.Errorf("root template not found (%s)", name)
	}
	if _, err := t.New("root").Parse(s); err != nil {
		return nil, err
	}
	return t, nil
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

// ParserElemetns returns the elements converted to parser.Elements.
func (es Elements) ParserElements() parser.Elements {
	m := parser.Elements{}
	for name, e := range es {
		m[name] = parser.Element{
			Name:      name,
			Type:      e.Type,
			Delimiter: e.Delimiter,
			Matcher:   e.Matcher,
		}
	}
	return m
}

// PrinterElements returns the elements converted to printer.Elements.
func (es Elements) PrinterElements() printer.Elements {
	m := printer.Elements{}
	for name, e := range es {
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

func (es Elements) parseTemplates(t *template.Template, name string) (*template.Template, error) {
	for n, e := range es {
		s, ok := e.Templates[name]
		if !ok {
			return nil, fmt.Errorf("element template not found (%s.%s)", n, name)
		}
		if _, err := t.New(n).Parse(s); err != nil {
			return nil, err
		}
	}
	return t, nil
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

func (gs Groups) parseTemplates(t *template.Template, name string) (*template.Template, error) {
	for n, g := range gs {
		s, ok := g.Templates[name]
		if !ok {
			return nil, fmt.Errorf("group template not found (%s.%s)", n, name)
		}
		if _, err := t.New(n).Parse(s); err != nil {
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
