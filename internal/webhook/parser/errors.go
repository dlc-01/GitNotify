package parser

import (
	"errors"
	"fmt"
)

var (
	ErrUnknownEvent = errors.New("unknown event type")
	ErrParsePayload = errors.New("failed to parse payload")
	ErrEmptyPayload = errors.New("empty payload")
	ErrMissingRepo  = errors.New("missing repository url")
)

type Error struct {
	Op     string
	Source string
	Err    error
}

func (e *Error) Error() string {
	return fmt.Sprintf("parser.%s [%s]: %v", e.Op, e.Source, e.Err)
}

func (e *Error) Unwrap() error {
	return e.Err
}

func wrap(op string, source string, err error) *Error {
	return &Error{Op: op, Source: source, Err: err}
}
