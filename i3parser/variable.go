package i3parser

import "github.com/abibby/yabai3/parser"

type Variable struct {
	*parser.Section
	Set   *Exact
	Name  *VariableName
	Value *Identifier
}

func NewVariable(s *parser.Section) *Variable {
	return &Variable{
		Section: s,
	}
}

func ParseVariable(parent parser.Node, block *parser.Reader) (parser.Node, error) {
	tx := block.BeginTx()
	defer tx.Rollback()

	v := NewVariable(tx.Commit())
	set, err := ExactParser("set")(v, block)
	if err != nil {
		return nil, parser.ErrWrongParser
	}
	v.Set = set.(*Exact)

	block.SkipWhitespace()

	name, err := ParseVariableName(v, block)
	if err != nil {
		return nil, err
	}
	v.Name = name.(*VariableName)

	block.SkipWhitespace()

	value, err := ParseIdentifier(v, block)
	if err != nil {
		return nil, err
	}
	v.Value = value.(*Identifier)

	return v, nil
}
