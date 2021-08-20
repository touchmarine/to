package config

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/touchmarine/to/node"
	"html/template"
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
	RootTemplates map[string]string `json:"rootTemplates"`
	Text          struct {
		Templates map[string]string `json:"templates"`
	}
	Autolink struct {
		Templates map[string]string `json:"templates"`
	}
	TextBlock struct {
		Templates map[string]string `json:"templates"`
	}
	Paragraph struct {
		Templates map[string]string `json:"templates"`
	}
	Elements   []Element   `json:"elements"`
	Composites []Composite `json:"composites"`
	Stickies   []Sticky    `json:"stickies"`
	Groups     []Group     `json:"groups"`
	Aggregates []Aggregate `json:"aggregates"`
}

func (c *Config) Element(name string) (Element, bool) {
	for _, el := range c.Elements {
		if el.Name == name {
			return el, true
		}
	}
	return Element{}, false
}

func (c *Config) Composite(name string) (Composite, bool) {
	for _, comp := range c.Composites {
		if comp.Name == name {
			return comp, true
		}
	}
	return Composite{}, false
}

func (c *Config) Sticky(name string) (Sticky, bool) {
	for _, sticky := range c.Stickies {
		if sticky.Name == name {
			return sticky, true
		}
	}
	return Sticky{}, false
}

func (c *Config) Group(name string) (Group, bool) {
	for _, grp := range c.Groups {
		if grp.Name == name {
			return grp, true
		}
	}
	return Group{}, false
}

func (c *Config) ParseTemplates(target *template.Template, name string) (*template.Template, error) {
	rootTmpl, ok := c.RootTemplates[name]
	if !ok {
		return nil, fmt.Errorf("root %s template not found", name)
	}
	if _, err := target.New("root").Parse(rootTmpl); err != nil {
		return nil, err
	}

	textTmpl, ok := c.Text.Templates[name]
	if !ok {
		return nil, fmt.Errorf("Text %s template not found", name)
	}
	if _, err := target.New("Text").Parse(textTmpl); err != nil {
		return nil, err
	}

	autolinkTmpl, ok := c.Autolink.Templates[name]
	if !ok {
		return nil, fmt.Errorf("Autolink %s template not found", name)
	}
	if _, err := target.New("Autolink").Parse(autolinkTmpl); err != nil {
		return nil, err
	}

	textBlockTmpl, ok := c.TextBlock.Templates[name]
	if !ok {
		return nil, fmt.Errorf("TextBlock %s template not found", name)
	}
	if _, err := target.New("TextBlock").Parse(textBlockTmpl); err != nil {
		return nil, err
	}

	paraTmpl, ok := c.Paragraph.Templates[name]
	if !ok {
		return nil, fmt.Errorf("Paragraph %s template not found", name)
	}
	if _, err := target.New("Paragraph").Parse(paraTmpl); err != nil {
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

// template encode:https://play.golang.org/p/ayrY0opKeEv
type Element struct {
	Name      string            `json:"name"`
	Type      node.Type         `json:"type"`
	Delimiter string            `json:"delimiter"`
	Templates map[string]string `json:"templates"`
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
	Name      string            `json:"name"`
	Element   string            `json:"element"`
	Templates map[string]string `json:"templates"`
}

// Aggregate is an aggregate of Elements.
//
// One current usage is generating table of contents based on an aggregate of
// headings.
type Aggregate struct {
	Name     string   `json:"name"`
	Elements []string `json"elements"`
}
