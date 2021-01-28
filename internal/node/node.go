package node

type Block interface {
	block()
}

type Blockquote struct {
	Children []Block
}

func (bq *Blockquote) block() {}

type ListItem struct {
	Children []Block
}

func (li ListItem) block() {}

type CodeBlock struct {
	Head string
	Body string
}

func (cb *CodeBlock) block() {}

type Paragraph struct {
	Children []Block
}

func (p *Paragraph) block() {}

type Line struct {
	Children []Inline
}

func (p Line) block() {}

type Inline interface {
	inline()
}

type Emphasis struct {
	Children []Inline
}

func (e *Emphasis) inline() {}

type Text string

func (t Text) inline() {}
