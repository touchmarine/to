package parser

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/touchmarine/to/node"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// pass -dbg when testing
var dbg = flag.Bool("dbg", false, "debug ParseFromHTML")

func OutHTML(els Elements, hels map[string]atom.Atom, root *node.Node) (*html.Node, error) {
	ctr := &html.Node{
		Type:     html.ElementNode,
		DataAtom: atom.Body,
		Data:     atom.Body.String(),
	}
	var f func(*html.Node, *node.Node)
	f = func(tr *html.Node, n *node.Node) {
		if a, ok := hels[n.Element]; ok {
			m := &html.Node{
				Type:     html.ElementNode,
				DataAtom: a,
				Data:     a.String(),
			}
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				f(m, c)
			}
			tr.AppendChild(m)
			return
		}

		switch n.Type {
		case node.TypeError:
			panic(fmt.Sprintf("encountered an error node: %v", n.String()))
		case node.TypeContainer, node.TypeLeaf:
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				f(tr, c)
			}
			return
		case node.TypeText:
			m := &html.Node{
				Type: html.TextNode,
				Data: n.Value,
			}
			tr.AppendChild(m)
			return
		}

		if n.Element == "" {
			panic(fmt.Sprintf("blank element for non-text node: %v", n))
		}
		var a atom.Atom
		if n.IsBlock() {
			a = atom.Div
		} else {
			a = atom.Span
		}
		m := &html.Node{
			Type:     html.ElementNode,
			DataAtom: a,
			Data:     a.String(),
			Attr: map2attrs(map[string]string{
				"class": n.Element,
			}),
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(m, c)
		}
		tr.AppendChild(m)
		return
	}
	f(ctr, root)
	return ctr, nil
}

func ParseFromHTML(els Elements, hels map[atom.Atom]string, src []byte) (*node.Node, error) {
	doc, err := html.Parse(bytes.NewReader(src))
	if err != nil {
		return nil, fmt.Errorf("parse html fragment: %v", err)
	}

	if *dbg {
		if err := html.Render(os.Stdout, doc); err != nil {
			return nil, fmt.Errorf("render html: %v", err)
		}
	}

	var f func(*node.Node, *html.Node)
	f = func(tr *node.Node, n *html.Node) {
		if *dbg {
			fmt.Println(n.Type, n.DataAtom, n.Data)
		}
		switch n.Type {
		case html.ElementNode:
			as := attrs2map(n.Attr)
			cl := as["class"]
			cls := strings.Fields(cl)
			if e, ok := searchElements(els, cls); ok {
				// found Element name in class attr
				m := &node.Node{
					Element: e.Name,
					Type:    e.Type,
				}
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					f(m, c)
				}
				tr.AppendChild(m)
				return
			}

			switch n.DataAtom {
			case atom.Div:
				m := &node.Node{
					Element: hels[n.DataAtom],
					Type:    node.TypeContainer,
				}
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					f(m, c)
				}
				tr.AppendChild(m)
				return
			case atom.B:
				m := &node.Node{
					Element: hels[n.DataAtom],
					Type:    node.TypeUniform,
				}
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					f(m, c)
				}
				tr.AppendChild(m)
				return
			case atom.P:
				m := &node.Node{
					Element: hels[n.DataAtom],
					Type:    node.TypeLeaf,
				}
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					f(m, c)
				}
				tr.AppendChild(m)
				return
			}
		case html.TextNode:
			m := &node.Node{
				Type:  node.TypeText,
				Value: n.Data,
			}
			tr.AppendChild(m)
			return
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(tr, c)
		}
	}
	ctr := &node.Node{
		Type: node.TypeContainer,
	}
	f(ctr, doc)
	return ctr, nil
}

func map2attrs(m map[string]string) []html.Attribute {
	as := make([]html.Attribute, 0, len(m))
	for k, v := range m {
		as = append(as, html.Attribute{
			Namespace: "",
			Key:       k,
			Val:       v,
		})
	}
	return as
}

func attrs2map(as []html.Attribute) map[string]string {
	m := make(map[string]string, len(as))
	for _, a := range as {
		if a.Namespace == "" {
			m[a.Key] = a.Val
		}
	}
	return m
}

func searchElements(els Elements, p []string) (Element, bool) {
	for _, s := range p {
		if e, ok := els[s]; ok {
			return e, true
		}
	}
	return Element{}, false
}
