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
	Line struct {
		Templates map[string]string `json:"templates"`
	}
	Paragraph struct {
		Templates map[string]string `json:"templates"`
	}
	Elements   []Element   `json:"elements"`
	Composites []Composite `json:"composites"`
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

	lineTmpl, ok := c.Line.Templates[name]
	if !ok {
		return nil, fmt.Errorf("Line %s template not found", name)
	}
	if _, err := target.New("Line").Parse(lineTmpl); err != nil {
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

// Composite is like an inline Group. It is an element created by traversing
// nodes and trying to recognize patterns between inline elements.
//
// The pattern is the PrimaryElement followed by the SecondaryElement.
type Composite struct {
	Name             string            `json:"name"`
	PrimaryElement   string            `json:"primaryElement"`
	SecondaryElement string            `json:"secondaryElement"`
	Templates        map[string]string `json:"templates"`
}

type Group struct {
	Name      string            `json:"name"`
	Element   string            `json:"element"`
	Templates map[string]string `json:"templates"`
}

type Aggregate struct {
	Name     string   `json:"name"`
	Elements []string `json"elements"`
}
