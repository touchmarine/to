package config

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"

	"github.com/touchmarine/to/aggregator"
	"github.com/touchmarine/to/aggregator/seqnum"
	"github.com/touchmarine/to/node"
	"github.com/touchmarine/to/parser"
	"github.com/touchmarine/to/printer"
	"github.com/touchmarine/to/transformer"
	"github.com/touchmarine/to/transformer/composite"
	"github.com/touchmarine/to/transformer/group"
	"github.com/touchmarine/to/transformer/paragraph"
	"github.com/touchmarine/to/transformer/sequence"
	"github.com/touchmarine/to/transformer/sticky"
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
	Elements   []Element   `json:"elements"`
	Composites []Composite `json:"composites"`
	Stickies   []Sticky    `json:"stickies"`
	Groups     []Group     `json:"groups"`
	Aggregates Aggregates  `json:"aggregates"`
}

type Element struct {
	Name        string            `json:"name"`
	Type        node.Type         `json:"type"`
	Delimiter   string            `json:"delimiter"`
	Matcher     string            `json:"matcher"`
	DoNotRemove bool              `json:"doNotRemove"`
	Templates   map[string]string `json:"templates"`
}

// Composite is a group of two inline elements, the PrimaryElement and the
// SecondaryElement.
//
// The SecondaryElement immediately follows the PrimaryElement.
type Composite struct {
	Name             string            `json:"name"`
	PrimaryElement   string            `json:"primaryElement"`
	SecondaryElement string            `json:"secondaryElement"`
	Templates        map[string]string `json:"templates"`
}

// Sticky is a group of two elements, the Element and an element the Element
// sticks to.
//
// The Element sticks to the preceding element if After is true. Otherwise,
// it sticks to the following element.
type Sticky struct {
	Name      string            `json:"name"`
	Element   string            `json:"element"`
	After     bool              `json:"after"`
	Templates map[string]string `json:"templates"`
}

// Group is a group of consecutive Elements.
type Group struct {
	Name       string            `json:"name"`
	Recognizer string            `json:"recognizer"`
	Element    string            `json:"element"`
	Templates  map[string]string `json:"templates"`
}

type Aggregates struct {
	SequentialNumbers map[string]struct {
		Elements []string `json:"elements"`
	} `json:"sequentialNumbers"`
}

func (c *Config) ParserElements() parser.ElementMap {
	m := parser.ElementMap{}
	for _, e := range c.Elements {
		m[e.Name] = parser.Element{
			Name:      e.Name,
			Type:      e.Type,
			Delimiter: e.Delimiter,
			Matcher:   e.Matcher,
		}
	}
	return m
}

func (c *Config) PrinterElements() printer.ElementMap {
	m := printer.ElementMap{}
	for _, e := range c.Elements {
		m[e.Name] = printer.Element{
			Name:        e.Name,
			Type:        e.Type,
			Delimiter:   e.Delimiter,
			Matcher:     e.Matcher,
			DoNotRemove: e.DoNotRemove,
		}
	}
	return m
}

func (c *Config) TransformerComposites() composite.Map {
	m := composite.Map{}
	for _, e := range c.Composites {
		m[e.PrimaryElement] = composite.Composite{
			Name:             e.Name,
			PrimaryElement:   e.PrimaryElement,
			SecondaryElement: e.SecondaryElement,
		}
	}
	return m
}

func (c *Config) GroupsByRecognizer(recognizer string) []Group {
	var groups []Group
	for _, g := range c.Groups {
		if g.Recognizer == recognizer {
			groups = append(groups, g)
		}
	}
	return groups
}

func (c *Config) TransformerGroups(recognizer string) group.Map {
	m := group.Map{}
	for _, g := range c.Groups {
		if g.Recognizer == recognizer {
			m[g.Element] = group.Group{
				Name:    g.Name,
				Element: g.Element,
			}
		}
	}
	return m
}

func (c *Config) TransformerStickies() sticky.Map {
	m := sticky.Map{}
	for _, s := range c.Stickies {
		m[s.Element] = sticky.Sticky{
			Name:    s.Name,
			Element: s.Element,
			After:   s.After,
		}
	}
	return m
}

func (c *Config) ParseTemplates(target *template.Template, name string) (*template.Template, error) {
	rootTmpl, ok := c.Root.Templates[name]
	if !ok {
		return nil, fmt.Errorf("root %s template not found", name)
	}
	if _, err := target.New("root").Parse(rootTmpl); err != nil {
		return nil, err
	}

	for _, el := range c.Elements {
		elTmpl, ok := el.Templates[name]
		if !ok {
			return nil, fmt.Errorf("element %s %s template not found", el.Name, name)
		}
		if _, err := target.New(el.Name).Parse(elTmpl); err != nil {
			return nil, err
		}
	}

	for _, composite := range c.Composites {
		cTmpl, ok := composite.Templates[name]
		if !ok {
			return nil, fmt.Errorf("composite %s %s template not found", composite.Name, name)
		}
		if _, err := target.New(composite.Name).Parse(cTmpl); err != nil {
			return nil, err
		}
	}

	for _, sticky := range c.Stickies {
		sTmpl, ok := sticky.Templates[name]
		if !ok {
			return nil, fmt.Errorf("sticky %s %s template not found", sticky.Name, name)
		}
		if _, err := target.New(sticky.Name).Parse(sTmpl); err != nil {
			return nil, err
		}
	}

	for _, group := range c.Groups {
		gTmpl, ok := group.Templates[name]
		if !ok {
			return nil, fmt.Errorf("group %s %s template not found", group.Name, name)
		}
		if _, err := target.New(group.Name).Parse(gTmpl); err != nil {
			return nil, err
		}
	}

	return target, nil
}

func (c *Config) DefaultTransformers() []transformer.Transformer {
	var transformers []transformer.Transformer

	paragraphGroups := c.GroupsByRecognizer("paragraph")
	if len(paragraphGroups) > 0 {
		paragrapher := paragraph.Transformer{paragraphGroups[0].Name}
		transformers = append(transformers, paragrapher)
	}

	grouper := group.Transformer{c.TransformerGroups("element")}
	compositer := composite.Transformer{c.TransformerComposites()}
	stickier := sticky.Transformer{c.TransformerStickies()}
	sequencer := sequence.Transformer{}

	transformers = append(transformers, []transformer.Transformer{
		grouper,
		compositer,
		stickier,
		sequencer,
	}...)
	return transformers
}

func (c *Config) DefaultAggregators() map[string]map[string]aggregator.Aggregator {
	m := map[string]map[string]aggregator.Aggregator{}
	for n, a := range c.Aggregates.SequentialNumbers {
		if m["sequentialNumbers"] == nil {
			m["sequentialNumbers"] = map[string]aggregator.Aggregator{}
		}

		m["sequentialNumbers"][n] = seqnum.Aggregator{a.Elements}
	}
	return m
}
