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

func attrs2map(as []html.Attribute) map[string]string {
	m := map[string]string{}
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
