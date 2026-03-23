package notifier

import (
	"errors"
	"fmt"
)

var (
	ErrHandleMessage = errors.New("failed to handle message")
	ErrSendMessage   = errors.New("failed to send message")
	ErrUnmarshal     = errors.New("failed to unmarshal message")
	ErrNoSubscribers = errors.New("no subscribers found")
)

type Error struct {
	Op    string
	Topic string
	Err   error
}

func (e *Error) Error() string {
	if e.Topic != "" {
		return fmt.Sprintf("notifier.%s [%s]: %v", e.Op, e.Topic, e.Err)
	}
	return fmt.Sprintf("notifier.%s: %v", e.Op, e.Err)
}

func (e *Error) Unwrap() error {
	return e.Err
}

func wrap(op string, topic string, err error) *Error {
	return &Error{Op: op, Topic: topic, Err: err}
}
