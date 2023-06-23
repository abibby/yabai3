package i3parser

import (
	"github.com/abibby/yabai3/parser"
)

func Parse(file string) (parser.Node, error) {
	r, err := parser.NewReaderFromFile(file)
	if err != nil {
		return nil, err
	}
	return ParseDocument(nil, r)
}
