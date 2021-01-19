package node

type Block interface {
	block()
}

type Paragraph struct {
	Children []Block
}

func (p *Paragraph) block() {}

type Lines []string

func (p Lines) block() {}
