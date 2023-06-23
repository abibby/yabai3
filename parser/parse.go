package parser

import (
	"errors"
)

var (
	ErrNoNode      = errors.New("no node")
	ErrWrongParser = errors.New("wrong parser")
)

type Parser func(parent Node, block *Reader) (Node, error)

func NextNode(parent Node, block *Reader, parsers ...Parser) (Node, error) {
	for _, p := range parsers {
		n, err := p(parent, block)
		if errors.Is(err, ErrWrongParser) {
		} else if err != nil {
			return nil, err
		}
		if n != nil {
			return n, nil
		}
	}

	return nil, ErrNoNode
}
