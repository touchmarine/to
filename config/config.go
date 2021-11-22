// package config provides structs and functions for configuring all of the core
// Touch packages.
package config

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"

	"github.com/touchmarine/to/node"
	"github.com/touchmarine/to/parser"
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

// Config holds abstracted, user-friendly options that can be used to configure
// all of the core Touch packages.
type Config struct {
	Templates  Templates
	Elements   Elements
	Aggregates Aggregates
}

// Templates maps formats to template strings.
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

// Elements maps element names to Elements.
type Elements map[string]Element

// Element is an abstraction of parser.Element, printer.Element, and groups
// (transformers).
type Element struct {
	Type      string    // node or group (transformer) type (e.g. walled, list)
	Delimiter string    // element delimiter
	Matcher   string    // element matcher name (e.g. url)
	Element   string    // group main element (e.g. list element)
	Target    string    // group target element (e.g. sticky target)
	Option    string    // extra option (primarily for one-off options)
	Templates Templates // map of formats to template strings
}

// ParserElements returns the elements with a valid node type converted to
// parser.Elements.
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

func (es Elements) parseTemplates(t *template.Template, format string) (*template.Template, error) {
	for n, e := range es {
		if _, err := e.Templates.parse(t, n, format); err != nil {
			return nil, err
		}
	}
	return t, nil
}

// Aggregates maps aggregate names to Aggregates.
type Aggregates map[string]Aggregate

// Aggregate holds aggregator options.
type Aggregate struct {
	Type     string   // which aggregator
	Elements []string // limit elements aggregator can aggregate from
}
