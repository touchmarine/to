package config

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"to/internal/node"
	"unicode/utf8"
)

//go:embed to.json
var b []byte
var Default Config

func init() {
	r := bytes.NewReader(b)
	if err := json.NewDecoder(r).Decode(&Default); err != nil {
		panic(err)
	}
}

type Config struct {
	Elements []Element `json:"elements"`
}

type Element struct {
	Name      string    `json:"name"`
	Type      node.Type `json:"type"`
	Delimiter rune      `json:"delimiter"`
	Ranked    bool      `json:"ranked"`
}

func (e *Element) UnmarshalJSON(text []byte) error {
	var el struct {
		Name      string    `json:"name"`
		Type      node.Type `json:"type"`
		Delimiter string    `json:"delimiter"`
		Ranked    bool      `json:"ranked"`
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

	*e = Element{el.Name, el.Type, r, el.Ranked}
	return nil
}
