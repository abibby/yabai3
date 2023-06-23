package i3parser

import (
	"fmt"

	"github.com/abibby/yabai3/parser"
)

type Exact struct {
	*parser.Section
	Value string
}

func NewExact(s *parser.Section, value string) *Exact {
	return &Exact{
		Section: s,
		Value:   value,
	}
}

func ExactParser(value string) parser.Parser {
	return func(parent parser.Node, block *parser.Reader) (parser.Node, error) {
		tx := block.BeginTx()
		defer tx.Rollback()

		b := block.ReadN(len(value))
		if string(b) != value {
			return nil, fmt.Errorf("expected %s received %s", value, b)
		}

		return NewExact(tx.Commit(), value), nil
	}
}
