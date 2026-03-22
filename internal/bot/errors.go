package bot

import (
	"errors"
	"fmt"

	"github.com/dlc-01/GitNotify/internal/domain"
)

var (
	ErrInvalidRepoURL = errors.New("invalid repo url")
	ErrInvalidEvent   = errors.New("invalid event type")
	ErrEmptyArgs      = errors.New("empty arguments")
)

type Error struct {
	Op  string
	Err error
}

func (e *Error) Error() string {
	return fmt.Sprintf("bot.%s: %v", e.Op, e.Err)
}

func (e *Error) Unwrap() error {
	return e.Err
}

func Wrap(op string, err error) *Error {
	return &Error{Op: op, Err: err}
}

func formatError(err error) string {
	var botErr *Error
	if errors.As(err, &botErr) {
		switch {
		case errors.Is(botErr, ErrEmptyArgs):
			return "❌ No arguments provided. Use /help to see usage"
		case errors.Is(botErr, ErrInvalidRepoURL):
			return "❌ Invalid URL. Only github.com and gitlab.com are supported"
		case errors.Is(botErr, ErrInvalidEvent):
			return fmt.Sprintf("❌ Invalid event type. Available: %s", formatEventTypes())
		}
	}
	return "❌ Something went wrong, please try again later"
}

func formatEventTypes() string {
	types := make([]string, len(domain.AllEventTypes))
	for i, e := range domain.AllEventTypes {
		types[i] = string(e)
	}
	return fmt.Sprintf("%v", types)
}
