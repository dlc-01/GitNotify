package validator

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidSignature = errors.New("invalid signature")
	ErrInvalidToken     = errors.New("invalid token")
	ErrMissingHeader    = errors.New("missing required header")
	ErrEmptySecret      = errors.New("empty secret")
)

type Error struct {
	Op     string
	Source string
	Err    error
}

func (e *Error) Error() string {
	return fmt.Sprintf("validator.%s [%s]: %v", e.Op, e.Source, e.Err)
}

func (e *Error) Unwrap() error {
	return e.Err
}

func wrap(op string, source string, err error) *Error {
	return &Error{Op: op, Source: source, Err: err}
}
