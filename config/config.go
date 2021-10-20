package config

import (
	_ "embed"
	"fmt"
	"html/template"

	"github.com/touchmarine/to/aggregator"
	sequentialnumberag "github.com/touchmarine/to/aggregator/sequentialnumber"
	"github.com/touchmarine/to/node"
	"github.com/touchmarine/to/parser"
	"github.com/touchmarine/to/printer"
	"github.com/touchmarine/to/transformer"
	"github.com/touchmarine/to/transformer/composite"
	"github.com/touchmarine/to/transformer/group"
	"github.com/touchmarine/to/transformer/sequentialnumber"
	"github.com/touchmarine/to/transformer/sticky"
)

//go:embed to.json
var b []byte

var Default *Config

//func init() {
//	r := bytes.NewReader(b)
//	if err := json.NewDecoder(r).Decode(&Default); err != nil {
//		panic(err)
//	}
//}

type Config struct {
	Root struct {
		Templates map[string]string `json:"templates"`
	} `json:"root"`
	Elements   map[string]Element   `json:"elements"`
	Groups     map[string]Group     `json:"groups"`
	Composites map[string]Composite `json:"composites"`
	Stickies   map[string]Sticky    `json:"stickies"`
	Aggregates Aggregates           `json:"aggregates"`
}

type Element struct {
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
	Element   string            `json:"element"`
	After     bool              `json:"after"`
	Templates map[string]string `json:"templates"`
}

// Group is a group of consecutive Elements.
//type Group struct {
//	Recognizer string            `json:"recognizer"`
//	Element    string            `json:"element"`
//	Templates  map[string]string `json:"templates"`
//}

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

func (c Config) ParserElements() parser.ElementMap {
	m := parser.ElementMap{}
	for name, e := range c.Elements {
		m[name] = parser.Element{
			Name:      name,
			Type:      e.Type,
			Delimiter: e.Delimiter,
			Matcher:   e.Matcher,
		}
	}
	return m
}

func (c Config) PrinterElements() printer.ElementMap {
	m := printer.ElementMap{}
	for name, e := range c.Elements {
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

func (c Config) TransformerComposites() composite.Map {
	m := composite.Map{}
	for name, e := range c.Composites {
		m[e.PrimaryElement] = composite.Composite{
			Name:             name,
			PrimaryElement:   e.PrimaryElement,
			SecondaryElement: e.SecondaryElement,
		}
	}
	return m
}

func (c Config) GroupsByType(t string) map[string]Group {
	m := map[string]Group{}
	for name, g := range c.Groups {
		if g.Type == t {
			m[name] = g
		}
	}
	return m
}

func (c Config) TransformerGroups(t string) group.Map {
	m := group.Map{}
	for name, g := range c.Groups {
		if g.Type == t {
			m[g.Element] = group.Group{
				Name:    name,
				Element: g.Element,
			}
		}
	}
	return m
}

func (c Config) TransformerStickies() sticky.Map {
	m := sticky.Map{}
	for name, s := range c.Stickies {
		m[s.Element] = sticky.Sticky{
			Name:    name,
			Element: s.Element,
			After:   s.After,
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

	for name, composite := range c.Composites {
		tmpl, ok := composite.Templates[templateName]
		if !ok {
			return nil, fmt.Errorf("composite template not found (%s %s)", name, templateName)
		}
		if _, err := target.New(name).Parse(tmpl); err != nil {
			return nil, err
		}
	}

	for name, sticky := range c.Stickies {
		tmpl, ok := sticky.Templates[templateName]
		if !ok {
			return nil, fmt.Errorf("sticky template not found (%s %s)", name, templateName)
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

func (c Config) DefaultTransformers() []transformer.Transformer {
	var transformers []transformer.Transformer

	//paragraphGroups := c.GroupsByType("paragraph")
	//for name, _ := range paragraphGroups {
	//	paragrapher := paragraph.Transformer{name}
	//	transformers = append(transformers, paragrapher)
	//}

	grouper := group.Transformer{c.TransformerGroups("element")}
	compositer := composite.Transformer{c.TransformerComposites()}
	stickier := sticky.Transformer{c.TransformerStickies()}

	transformers = append(transformers, []transformer.Transformer{
		grouper,
		compositer,
		stickier,
		transformer.Func(sequentialnumber.Transform),
	}...)
	return transformers
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
