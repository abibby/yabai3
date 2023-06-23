package i3parser

import (
	"github.com/abibby/yabai3/parser"
)

var KindBorder = CommandKind("border")

type Border struct {
	*parser.Section
	Border *Exact
}

func (b *Border) Kind() CommandKind {
	return KindBorder
}

func ParseBorder(parent parser.Node, block *parser.Reader) (parser.Node, error) {
	tx := block.BeginTx()
	defer tx.Rollback()

	b := &Border{}

	border, err := ExactParser("border")(b, block)
	if err != nil {
		return nil, parser.ErrWrongParser
	}
	b.Border = border.(*Exact)

	b.Section = tx.Commit()
	return b, nil
}
