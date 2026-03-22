package postgres

import (
	"errors"
	"fmt"
)

var (
	ErrConnect = errors.New("failed to connect to postgres")
	ErrPing    = errors.New("failed to ping postgres")
)

type Error struct {
	Op  string
	Err error
}

func (e *Error) Error() string {
	return fmt.Sprintf("postgres.%s: %v", e.Op, e.Err)
}

func (e *Error) Unwrap() error {
	return e.Err
}

func wrap(op string, err error) *Error {
	return &Error{Op: op, Err: err}
}
