// Package config provides types and functions for configuring the core Touch
// packages.
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

// Default is the default Config.
var Default = defaultConfig()

func defaultConfig() Config {
	var c Config
	r := bytes.NewReader(b)
	if err := json.NewDecoder(r).Decode(&c); err != nil {
		panic(err)
	}
	return c
}

// Config holds abstracted options that can be used to configure the core Touch
// packages.
type Config struct {
	Templates  Templates
	Elements   Elements
	Aggregates Aggregates
}

// Templates is a map of formats to template strings.
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

// ParseTemplates parses config templates that match the given format as
// template bodies for the given template.
func (c Config) ParseTemplates(t *template.Template, format string) (*template.Template, error) {
	if _, err := c.Templates.parse(t, "root", format); err != nil {
		return nil, err
	}
	if _, err := c.Elements.parseTemplates(t, format); err != nil {
		return nil, err
	}
	return t, nil
}

// Elements is a map of element names to Elements.
type Elements map[string]Element

// Element is an abstraction of parser.Element and transformer options.
type Element struct {
	Type      string    // node type or transformer name (e.g. walled, list)
	Delimiter string    // element delimiter
	Matcher   string    // prefixed element matcher name (e.g. url)
	Element   string    // transformer main element (list element)
	Target    string    // transformer target element (sticky target)
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

// Aggregates is a map of aggregate names to Aggregates.
type Aggregates map[string]Aggregate

// Aggregate holds aggregator options.
type Aggregate struct {
	Type     string   // which aggregator (aggregator name)
	Elements []string // allowed elements to aggregate from
}
