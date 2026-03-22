package callback

import (
	"context"
	"strings"
)

type Handler interface {
	Action() string
	Execute(ctx context.Context, chatID int64, messageID int, args string)
}

type Registry struct {
	handlers map[string]Handler
}

func NewRegistry() *Registry {
	return &Registry{handlers: make(map[string]Handler)}
}

func (r *Registry) Register(h Handler) {
	r.handlers[h.Action()] = h
}

func (r *Registry) Get(data string) (Handler, string, bool) {
	parts := strings.SplitN(data, ":", 2)
	action := parts[0]
	args := ""
	if len(parts) == 2 {
		args = parts[1]
	}
	h, ok := r.handlers[action]
	return h, args, ok
}
