package parser

type Node interface {
	Pos() int // position of first character belonging to the node
	End() int // position of first character immediately after the node
}

// type ParentNode interface {
// 	Node
// 	AppendChild(child Node)
// 	Children() []Node
// }

type ParserNode interface {
	Node
	Parsers() []Parser
}

type Section struct {
	Start  int
	Length int
}

func NewSection(start, length int) *Section {
	return &Section{
		Start:  start,
		Length: length,
	}
}

func (d *Section) Pos() int {
	return d.Start
}
func (d *Section) End() int {
	return d.Start + d.Length
}
