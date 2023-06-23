package i3parser

import (
	"github.com/abibby/yabai3/parser"
)

type Whitespace struct {
	*parser.Section
}

func NewWhitespace(s *parser.Section) *Whitespace {
	return &Whitespace{
		Section: s,
	}
}

func ParseWhitespace(parent parser.Node, block *parser.Reader) (parser.Node, error) {
	tx := block.BeginTx()
	defer tx.Rollback()

	hasWhitespace := false

	for isWhitespace(block.Peak()) {
		block.Advance(1)
		hasWhitespace = true
	}
	if !hasWhitespace {
		return nil, nil
	}
	return NewWhitespace(tx.Commit()), nil
}
