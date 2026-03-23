package commands

import (
	"context"

	"github.com/dlc-01/GitNotify/internal/bot/core"
)

type StartCommand struct {
	sender   core.Senderer
	registry *core.Registry
}

func NewStartCommand(sender core.Senderer, registry *core.Registry) *StartCommand {
	return &StartCommand{sender: sender, registry: registry}
}

func (c *StartCommand) Name() string        { return "start" }
func (c *StartCommand) Description() string { return "Start the bot" }
func (c *StartCommand) Usage() string       { return "/start" }

func (c *StartCommand) Execute(ctx context.Context, chatID int64, args string) {
	help := &HelpCommand{sender: c.sender, registry: c.registry}
	c.sender.Send(chatID, help.format())
}
