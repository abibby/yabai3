package i3parser

import (
	"github.com/abibby/yabai3/parser"
)

type ForWindow struct {
	*parser.Section
	ForWindow  *Exact
	Conditions *Conditions
	Command    Command
}

func NewForWindow(s *parser.Section) *ForWindow {
	return &ForWindow{
		Section: s,
	}
}

func ParseForWindow(parent parser.Node, block *parser.Reader) (parser.Node, error) {
	tx := block.BeginTx()
	defer tx.Rollback()

	w := NewForWindow(tx.Commit())

	forWindow, err := ExactParser("for_window")(w, block)
	if err != nil {
		return nil, parser.ErrWrongParser
	}
	w.ForWindow = forWindow.(*Exact)

	block.SkipWhitespace()

	conditions, err := ParseConditions(parent, block)
	if err != nil {
		return nil, err
	}
	w.Conditions = conditions.(*Conditions)

	block.SkipWhitespace()

	command, err := ParseCommand(parent, block)
	if err != nil {
		return nil, err
	}
	w.Command = command.(Command)

	return w, nil
}
