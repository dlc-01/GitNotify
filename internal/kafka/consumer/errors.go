package consumer

import (
	"errors"
	"fmt"
)

var (
	ErrConnect   = errors.New("failed to connect to kafka")
	ErrConsume   = errors.New("failed to consume message")
	ErrUnmarshal = errors.New("failed to unmarshal message")
	ErrClosed    = errors.New("consumer is closed")
)

type Error struct {
	Op    string
	Topic string
	Err   error
}

func (e *Error) Error() string {
	if e.Topic != "" {
		return fmt.Sprintf("consumer.%s [%s]: %v", e.Op, e.Topic, e.Err)
	}
	return fmt.Sprintf("consumer.%s: %v", e.Op, e.Err)
}

func (e *Error) Unwrap() error {
	return e.Err
}

func wrap(op string, topic string, err error) *Error {
	return &Error{Op: op, Topic: topic, Err: err}
}
