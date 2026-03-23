package webhook

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidSignature = errors.New("invalid signature")
	ErrInvalidToken     = errors.New("invalid token")
	ErrUnknownSource    = errors.New("unknown webhook source")
	ErrUnknownEvent     = errors.New("unknown event type")
	ErrParsePayload     = errors.New("failed to parse payload")
	ErrEmptyPayload     = errors.New("empty payload")
)

type Error struct {
	Op  string
	Err error
}

func (e *Error) Error() string {
	return fmt.Sprintf("webhook.%s: %v", e.Op, e.Err)
}

func (e *Error) Unwrap() error {
	return e.Err
}

func wrap(op string, err error) *Error {
	return &Error{Op: op, Err: err}
}
