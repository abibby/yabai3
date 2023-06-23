package i3parser

import (
	"errors"
	"fmt"

	"github.com/abibby/yabai3/parser"
)

type Document struct {
	length   int
	children []parser.Node
}

func NewDocument(length int) *Document {
	return &Document{
		length:   length,
		children: []parser.Node{},
	}
}

func (d *Document) Pos() int {
	return 0
}
func (d *Document) End() int {
	return d.length
}
func (d *Document) Children() []parser.Node {
	return d.children
}

var docChildren = []parser.Parser{
	ParseVariable,
	ParseGaps,
	ParseSmartGaps,
	ParseForWindow,
	ParseWhitespace,
	ParseCommand,
}

func ParseDocument(parent parser.Node, block *parser.Reader) (parser.Node, error) {
	d := NewDocument(len(block.Source()))
	for {
		c, err := parser.NextNode(d, block, docChildren...)
		if errors.Is(err, parser.ErrNoNode) {
		} else if err != nil {
			return nil, err
		}
		if c == nil {
			if block.Peak() != parser.EOF {
				return nil, parser.NewError(block, fmt.Errorf("unexpected content \"%s\"", block.PeakLine()))
			}
			return d, nil
		}
		d.children = append(d.children, c)
	}
}
