package poller

import (
	"errors"
	"fmt"
)

var (
	ErrFetch         = errors.New("failed to fetch data")
	ErrParse         = errors.New("failed to parse response")
	ErrInvalidURL    = errors.New("invalid url")
	ErrUnknownSource = errors.New("unknown source")
)

type Error struct {
	Op     string
	Source string
	URL    string
	Err    error
}

func (e *Error) Error() string {
	return fmt.Sprintf("poller.%s [%s] %s: %v", e.Op, e.Source, e.URL, e.Err)
}

func (e *Error) Unwrap() error {
	return e.Err
}

func wrap(op string, source string, url string, err error) *Error {
	return &Error{Op: op, Source: source, URL: url, Err: err}
}
