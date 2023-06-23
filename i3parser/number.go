package i3parser

import (
	"fmt"
	"strconv"

	"github.com/abibby/yabai3/parser"
)

type Number struct {
	*parser.Section
	Value float64
}

func NewNumber(s *parser.Section, value float64) *Number {
	return &Number{
		Section: s,
		Value:   value,
	}
}

func ParseNumber(parent parser.Node, block *parser.Reader) (parser.Node, error) {
	tx := block.BeginTx()
	defer tx.Rollback()

	result := ""
	b := block.Peak()
	for (b >= '0' && b <= '9') || (b == '.' && result != "") {
		block.Advance(1)
		result += string(b)
		b = block.Peak()
	}

	if !isWhitespace(b) || result == "" {
		return nil, fmt.Errorf("expected [0-9.] received %c", b)
	}

	value, err := strconv.ParseFloat(result, 64)
	if err != nil {
		return nil, err
	}
	return NewNumber(tx.Commit(), value), nil
}
