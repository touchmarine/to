package node

type Block interface {
	block()
}

type Blockquote struct {
	Children []Block
}

func (bq *Blockquote) block() {}

type Paragraph struct {
	Children []Block
}

func (p *Paragraph) block() {}

type Lines []string

func (p Lines) block() {}

type ListItem struct {
	Children []Block
}

func (li ListItem) block() {}

type CodeBlock struct {
	Head string
	Body string
}

func (cb *CodeBlock) block() {}

type Inline interface {
	inline()
}

type Emphasis struct {
	Children []Inline
}

func (e *Emphasis) inline() {}
