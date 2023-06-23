package i3parser

import (
	"fmt"

	"github.com/abibby/yabai3/parser"
)

type Condition struct {
	Type  *Identifier
	Value *String
}

type Conditions struct {
	*parser.Section

	Conditions []*Condition
}

func NewConditions(s *parser.Section) *Conditions {
	return &Conditions{
		Section:    s,
		Conditions: []*Condition{},
	}
}

func ParseConditions(parent parser.Node, block *parser.Reader) (parser.Node, error) {
	tx := block.BeginTx()
	defer tx.Rollback()

	b := block.ReadByte()

	if b != '[' {
		return nil, parser.NewError(block, fmt.Errorf("expected [ received %c", b))
	}

	c := NewConditions(tx.Commit())
	for block.Peak() != ']' {
		t, err := ParseIdentifier(c, block)
		if err != nil {
			return nil, parser.NewError(block, err)
		}

		block.SkipWhitespace()

		b = block.ReadByte()
		if b != '=' {
			return nil, parser.NewError(block, fmt.Errorf("expected = received %c", b))
		}

		block.SkipWhitespace()

		v, err := ParseString(c, block)
		if err != nil {
			return nil, parser.NewError(block, err)
		}
		c.Conditions = append(c.Conditions, &Condition{
			Type:  t.(*Identifier),
			Value: v.(*String),
		})
		block.SkipWhitespace()
	}
	block.Advance(1)

	return c, nil
}
