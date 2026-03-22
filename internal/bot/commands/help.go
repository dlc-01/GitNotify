package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/dlc-01/GitNotify/internal/bot/core"
)

type HelpCommand struct {
	sender   *core.Sender
	registry *core.Registry
}

func NewHelpCommand(sender *core.Sender, registry *core.Registry) *HelpCommand {
	return &HelpCommand{sender: sender, registry: registry}
}

func (c *HelpCommand) Name() string        { return "help" }
func (c *HelpCommand) Description() string { return "Show available commands" }
func (c *HelpCommand) Usage() string       { return "/help" }

func (c *HelpCommand) Execute(ctx context.Context, chatID int64, args string) {
	c.sender.Send(chatID, c.format())
}

func (c *HelpCommand) format() string {
	var sb strings.Builder
	sb.WriteString("GitNotify — GitHub & GitLab notifications in Telegram\n\n")
	sb.WriteString("Commands:\n")
	for _, cmd := range c.registry.All() {
		sb.WriteString(fmt.Sprintf("%-30s — %s\n", cmd.Usage(), cmd.Description()))
	}
	return sb.String()
}
