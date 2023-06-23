package i3parser

import (
	"fmt"

	"github.com/abibby/yabai3/parser"
)

type VariableName struct {
	*parser.Section
	Value string
}

func NewVariableName(s *parser.Section, value string) *VariableName {
	return &VariableName{
		Section: s,
		Value:   value,
	}
}

func ParseVariableName(parent parser.Node, block *parser.Reader) (parser.Node, error) {
	tx := block.BeginTx()
	defer tx.Rollback()

	word := block.ReadWord()
	if len(word) == 0 {
		return nil, fmt.Errorf("could not find variable name")
	}
	if word[0] != '$' {
		return nil, fmt.Errorf("invalid variable name %s: must start with $", word)
	}

	return NewVariableName(tx.Commit(), string(word)), nil
}
