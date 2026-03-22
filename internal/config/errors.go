package config

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound  = errors.New("config file not found")
	ErrInvalid   = errors.New("config file is invalid")
	ErrUnmarshal = errors.New("failed to unmarshal config")
)

type Error struct {
	Path string
	Err  error
}

func (e *Error) Error() string {
	return fmt.Sprintf("config %q: %v", e.Path, e.Err)
}

func (e *Error) Unwrap() error {
	return e.Err
}

func wrap(path string, err error) *Error {
	return &Error{Path: path, Err: err}
}
