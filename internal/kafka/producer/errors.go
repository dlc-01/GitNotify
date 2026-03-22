package producer

import (
	"errors"
	"fmt"
)

var (
	ErrConnect = errors.New("failed to connect to kafka")
	ErrProduce = errors.New("failed to produce message")
	ErrMarshal = errors.New("failed to marshal message")
	ErrClosed  = errors.New("producer is closed")
)

type Error struct {
	Op    string
	Topic string
	Err   error
}

func (e *Error) Error() string {
	if e.Topic != "" {
		return fmt.Sprintf("producer.%s [%s]: %v", e.Op, e.Topic, e.Err)
	}
	return fmt.Sprintf("producer.%s: %v", e.Op, e.Err)
}

func (e *Error) Unwrap() error {
	return e.Err
}

func wrap(op string, topic string, err error) *Error {
	return &Error{Op: op, Topic: topic, Err: err}
}
