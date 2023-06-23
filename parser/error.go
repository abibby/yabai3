package parser

import (
	"fmt"
)

type Error struct {
	file    string
	line    int
	column  int
	wrapped error
}

func NewError(r *Reader, err error) *Error {
	return &Error{
		file:    r.File(),
		line:    r.Line(),
		column:  r.Column(),
		wrapped: err,
	}
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s:%d:%d: %s", e.file, e.line, e.column, e.wrapped.Error())
}
func (e *Error) Unwrap() error {
	return e.wrapped
}
