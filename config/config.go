// Package config provides types and functions for configuring the core Touch
// packages.
package config

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"

	"gopkg.in/yaml.v2"

	"github.com/touchmarine/to/node"
	"github.com/touchmarine/to/parser"
)

//go:embed to.yaml
var b []byte

// Default is the default Config.
var Default = defaultConfig()

func defaultConfig() Config {
	var c Config
	r := bytes.NewReader(b)
	dec := yaml.NewDecoder(r)
	dec.SetStrict(true)
	if err := dec.Decode(&c); err != nil {
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

// ParseTemplates parses config templates that match the given format as
// template bodies for the given template.
func (c Config) ParseTemplates(t *template.Template, format string) (*template.Template, error) {
	for n, e := range c.Elements {
		if e.Disabled {
			continue
		}
		s, ok := e.Templates[format]
		if !ok {
			return nil, fmt.Errorf("template not found: name=%q format=%q", n, format)
		}
		if _, err := t.New(n).Parse(s); err != nil {
			return nil, err
		}
	}
	s, ok := c.Templates[format]
	if !ok {
		return nil, fmt.Errorf("template not found: format=%q", format)
	}
	if _, err := t.Parse(s); err != nil {
		return nil, err
	}
	return t, nil
}

// Templates is a map of formats to template strings.
type Templates map[string]string

// Elements is a map of element names to Elements.
type Elements map[string]Element

// Element is an abstraction of parser.Element and transformer options.
type Element struct {
	Disabled        bool      // disabled=as if the element wasn't present
	Type            string    // node type or transformer name (e.g. walled, list)
	Delimiter       string    // element delimiter
	Matcher         string    // prefixed element matcher name (e.g. url)
	Element         string    // transformer main element (list element)
	Target          string    // transformer target element (sticky target)
	Option          string    // extra option (primarily for one-off options)
	Templates       Templates // map of formats to template strings
	StickyTemplates Templates // map of formats to template strings
	TargetTemplates Templates // map of formats to template strings
}

// ParserElements returns the elements with a valid node type converted to
// parser.Elements.
func (es Elements) ParserElements() parser.Elements {
	m := parser.Elements{}
	for n, e := range es {
		if e.Disabled {
			continue
		}
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

// Aggregates is a map of aggregate names to Aggregates.
type Aggregates map[string]Aggregate

// Aggregate holds aggregator options.
type Aggregate struct {
	Type      string   // which aggregator (aggregator name)
	Elements  []string // allowed elements to aggregate from
	Templates map[string]map[string]string
}
