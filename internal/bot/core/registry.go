package core

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Command interface {
	Name() string
	Description() string
	Usage() string
	Execute(ctx context.Context, chatID int64, args string)
}

type Registry struct {
	commands map[string]Command
}

func NewRegistry() *Registry {
	return &Registry{commands: make(map[string]Command)}
}

func (r *Registry) Register(cmd Command) {
	r.commands[cmd.Name()] = cmd
}

func (r *Registry) Get(name string) (Command, bool) {
	cmd, ok := r.commands[name]
	return cmd, ok
}

func (r *Registry) All() []Command {
	cmds := make([]Command, 0, len(r.commands))
	for _, cmd := range r.commands {
		cmds = append(cmds, cmd)
	}
	return cmds
}

func (r *Registry) BotCommands() []tgbotapi.BotCommand {
	cmds := make([]tgbotapi.BotCommand, 0, len(r.commands))
	for _, cmd := range r.commands {
		cmds = append(cmds, tgbotapi.BotCommand{
			Command:     cmd.Name(),
			Description: cmd.Description(),
		})
	}
	return cmds
}
