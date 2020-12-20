package node

type ListType string

// list types
const (
	UnorderedList ListType = "unordered"
	NumberedList           = "numbers"
	//LowercaseLetters                = "lowercaseLetters"
	//UppercaseLetters                = "uppercaseLetters"
	//LowercaseRomanNumerals          = "lowercaseRomanNumerals"
	//UppercaseRomanNumerals          = "uppercaseRomanNumerals"
)

type Node interface {
	node() // dummy method to conform to interface
}

type Block interface {
	Node
	block() // dummy method
}

type Inline interface {
	Node
	inline() // dummy method
}

type Document struct {
	Children []Node
}

func (d *Document) node() {}

type Paragraph struct {
	Lines []*Line
}

func (p Paragraph) node()  {}
func (p Paragraph) block() {}

// currently line is used only by paragraph
type Line struct {
	Children []Inline
}

func (l Line) node()   {}
func (l Line) inline() {}

type Text struct {
	Value string
}

func (t Text) node()   {}
func (t Text) inline() {}

type Emphasis struct {
	Children []Inline
}

func (e Emphasis) node()   {}
func (e Emphasis) inline() {}

type Strong struct {
	Children []Inline
}

func (s Strong) node()   {}
func (s Strong) inline() {}

type Heading struct {
	Level      int
	IsNumbered bool
	Children   []Inline
}

func (h Heading) node()  {}
func (h Heading) block() {}

type Link struct {
	Destination string
	Children    []Inline
}

func (l Link) node()   {}
func (l Link) inline() {}

type CodeBlock struct {
	Language    string
	Filename    string
	MetadataRaw string
	Body        string
}

func (cb CodeBlock) node()  {}
func (cb CodeBlock) block() {}

type List struct {
	//IsContinued bool     // whether counting continues onward from the previous list
	Type      ListType // unordered or numbering type if ordered
	ListItems []*ListItem
}

func (l List) node()  {}
func (l List) block() {}

type ListItem struct {
	Children []Node
}

func (li ListItem) node()  {}
func (li ListItem) block() {}

// InlinesToNodes converts []Inline to []Node.
func InlinesToNodes(inlines []Inline) []Node {
	nodes := make([]Node, len(inlines))
	for i, v := range inlines {
		nodes[i] = Node(v)
	}
	return nodes
}

// LinesToNodes converts []*Line to []Node.
func LinesToNodes(lines []*Line) []Node {
	nodes := make([]Node, len(lines))
	for i, v := range lines {
		nodes[i] = Node(v)
	}
	return nodes
}

// ListItemsToNodes converts []*ListItem to []Node.
func ListItemsToNodes(listItems []*ListItem) []Node {
	nodes := make([]Node, len(listItems))
	for i, v := range listItems {
		nodes[i] = Node(v)
	}
	return nodes
}
