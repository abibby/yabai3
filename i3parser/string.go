package i3parser

import (
	"fmt"

	"github.com/abibby/yabai3/parser"
)

type String struct {
	*parser.Section
	Value string
}

func NewString(s *parser.Section, value string) *String {
	return &String{
		Section: s,
		Value:   value,
	}
}

func ParseString(parent parser.Node, block *parser.Reader) (parser.Node, error) {
	tx := block.BeginTx()
	defer tx.Rollback()

	b := block.ReadByte()
	if b != '"' {
		return nil, parser.NewError(block, fmt.Errorf("expected \" recieved %c", b))
	}

	result := ""
	quote := b
	escape := false
	for {
		b = block.ReadByte()
		if escape {
			result += string(b)
		}
		if b == '\\' {
			escape = true
			continue
		}
		if b == quote {
			return NewString(tx.Commit(), result), nil
		}
		result += string(b)
	}
}
