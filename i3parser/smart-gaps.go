package i3parser

import (
	"github.com/abibby/yabai3/parser"
)

type SmartGaps struct {
	*parser.Section
	SmartGaps *Exact
	Value     *Bool
}

func NewSmartGaps(s *parser.Section) *SmartGaps {
	return &SmartGaps{
		Section: s,
	}
}

func ParseSmartGaps(parent parser.Node, block *parser.Reader) (parser.Node, error) {
	tx := block.BeginTx()
	defer tx.Rollback()

	s := NewSmartGaps(tx.Commit())

	smartGaps, err := ExactParser("smart_gaps")(s, block)
	if err != nil {
		return nil, parser.ErrWrongParser
	}
	s.SmartGaps = smartGaps.(*Exact)

	block.SkipWhitespace()

	value, err := ParseBool(s, block)
	if err != nil {
		return nil, err
	}
	s.Value = value.(*Bool)

	return s, nil
}
