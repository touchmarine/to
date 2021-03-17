package config

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"to/internal/node"
	"unicode/utf8"
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
	Text struct {
		Templates map[string]string `json:"templates"`
	} `json:"text"`
	Line struct {
		Templates map[string]string `json:"templates"`
	} `json:"line"`
	Elements []Element `json:"elements"`
}

func (c *Config) Element(name string) (Element, bool) {
	for _, el := range c.Elements {
		if el.Name == name {
			return el, true
		}
	}
	return Element{}, false
}

func (c *Config) ParseTemplates(target *template.Template, name string) (*template.Template, error) {
	textTmpl, ok := c.Text.Templates[name]
	if !ok {
		return nil, fmt.Errorf("template %s for Text not found", name)
	}
	if _, err := target.New("Text").Parse(textTmpl); err != nil {
		return nil, err
	}

	lineTmpl, ok := c.Line.Templates[name]
	if !ok {
		return nil, fmt.Errorf("template %s for Line not found", name)
	}
	if _, err := target.New("Line").Parse(lineTmpl); err != nil {
		return nil, err
	}

	for _, el := range c.Elements {
		elTmpl, ok := el.Templates[name]
		if !ok {
			return nil, fmt.Errorf("template %s for %s not found", name, el.Name)
		}
		if _, err := target.New(el.Name).Parse(elTmpl); err != nil {
			return nil, err
		}
	}

	return target, nil
}

// template encode:https://play.golang.org/p/ayrY0opKeEv
type Element struct {
	Name      string            `json:"name"`
	Type      node.Type         `json:"type"`
	Delimiter rune              `json:"delimiter"`
	Ranked    bool              `json:"ranked"`
	MinRank   uint              `json:"minRank"`
	Leaf      bool              `json:"leaf"`
	Templates map[string]string `json:"templates"`
}

func (e *Element) UnmarshalJSON(text []byte) error {
	var el struct {
		Name      string            `json:"name"`
		Type      node.Type         `json:"type"`
		Delimiter string            `json:"delimiter"`
		Ranked    bool              `json:"ranked"`
		MinRank   uint              `json:"minRank"`
		Leaf      bool              `json:"leaf"`
		Templates map[string]string `json:"templates"`
	}
	if err := json.Unmarshal(text, &el); err != nil {
		return err
	}

	r, w := utf8.DecodeRuneInString(el.Delimiter)
	if r == utf8.RuneError {
		return errors.New("invalid delimiter")
	}

	if len(el.Delimiter[w:]) > 0 {
		return errors.New("invalid delimiter")
	}

	*e = Element{el.Name, el.Type, r, el.Ranked, el.MinRank, el.Leaf, el.Templates}
	return nil
}
