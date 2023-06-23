package i3parser

import (
	"fmt"

	"github.com/abibby/yabai3/parser"
)

type Bool struct {
	*parser.Section
	Value bool
}

func NewBool(s *parser.Section, value bool) *Bool {
	return &Bool{
		Section: s,
		Value:   value,
	}
}

func ParseBool(parent parser.Node, block *parser.Reader) (parser.Node, error) {
	tx := block.BeginTx()
	defer tx.Rollback()

	valueStr := string(block.ReadWord())

	if valueStr != "on" && valueStr != "off" {
		return nil, fmt.Errorf("invalid bool value %s, must be on or off", valueStr)
	}

	return NewBool(tx.Commit(), valueStr == "on"), nil
}
