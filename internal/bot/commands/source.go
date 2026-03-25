package commands

import (
	"context"
	"strings"

	"github.com/dlc-01/GitNotify/internal/bot/core"
)

type SourcesCommand struct {
	sender core.Senderer
}

func NewSourcesCommand(sender core.Senderer) *SourcesCommand {
	return &SourcesCommand{sender: sender}
}

func (c *SourcesCommand) Name() string        { return "sources" }
func (c *SourcesCommand) Description() string { return "Show supported sources and event types" }
func (c *SourcesCommand) Usage() string       { return "/sources" }

func (c *SourcesCommand) Execute(ctx context.Context, chatID int64, args string) {
	c.sender.Send(chatID, formatSources())
}

func formatSources() string {
	var sb strings.Builder
	sb.WriteString("Supported sources:\n\n")

	sb.WriteString("GitHub\n")
	sb.WriteString("  /subscribe https://github.com/user/repo\n")
	sb.WriteString("  events: push, pr, issue, pipeline\n\n")

	sb.WriteString("GitLab\n")
	sb.WriteString("  /subscribe https://gitlab.com/user/repo\n")
	sb.WriteString("  events: push, pr, issue, pipeline\n\n")

	sb.WriteString("Stack Overflow\n")
	sb.WriteString("  /subscribe https://stackoverflow.com/questions/tagged/golang\n")
	sb.WriteString("  events: answer\n\n")

	sb.WriteString("Reddit\n")
	sb.WriteString("  /subscribe https://reddit.com/r/golang\n")
	sb.WriteString("  events: post\n\n")

	sb.WriteString("YouTube\n")
	sb.WriteString("  /subscribe https://youtube.com/@channel\n")
	sb.WriteString("  events: video\n")

	return sb.String()
}
