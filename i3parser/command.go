package i3parser

import "github.com/abibby/yabai3/parser"

type CommandKind string

type Command interface {
	parser.Node
	Kind() CommandKind
}

func ParseCommand(parent parser.Node, block *parser.Reader) (parser.Node, error) {
	return parser.NextNode(parent, block)
}
