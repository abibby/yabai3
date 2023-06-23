package i3parser

import (
	"fmt"

	"github.com/abibby/yabai3/parser"
)

type Gaps struct {
	*parser.Section
	Gaps *Exact
	Type *Identifier
	With *Number
}

func NewGaps(s *parser.Section) *Gaps {
	return &Gaps{
		Section: s,
	}
}

func ParseGaps(parent parser.Node, block *parser.Reader) (parser.Node, error) {
	tx := block.BeginTx()
	defer tx.Rollback()

	g := NewGaps(tx.Commit())

	gaps, err := ExactParser("gaps")(g, block)
	if err != nil {
		return nil, parser.ErrWrongParser
	}
	g.Gaps = gaps.(*Exact)

	block.SkipWhitespace()

	gapType, err := ParseIdentifier(g, block)
	if err != nil {
		return nil, err
	}
	g.Type = gapType.(*Identifier)

	block.SkipWhitespace()

	width, err := ParseNumber(g, block)
	if err != nil {
		return nil, err
	}
	g.With = width.(*Number)

	if g.Type.Value != "inner" && g.Type.Value != "outer" {
		return nil, fmt.Errorf("invalid gap type %s, must be inner or outer", g.Type.Value)
	}

	return g, nil
}
