package i3parser

import (
	"fmt"

	"github.com/abibby/yabai3/parser"
)

type Identifier struct {
	*parser.Section
	Value string
}

func NewIdentifier(s *parser.Section, value string) *Identifier {
	return &Identifier{
		Section: s,
		Value:   value,
	}
}

func ParseIdentifier(parent parser.Node, block *parser.Reader) (parser.Node, error) {
	tx := block.BeginTx()
	defer tx.Rollback()

	id := ""
	for isIdentifierCharacter(block.Peak()) {
		id += string(block.ReadByte())
	}
	if len(id) == 0 {
		return nil, parser.NewError(block, fmt.Errorf("invalid identifier"))
	}
	return NewIdentifier(tx.Commit(), id), nil
}

func isIdentifierCharacter(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9')
}
