package commands

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dlc-01/GitNotify/internal/bot/core"
)

func TestHelpCommand_Execute(t *testing.T) {
	registry := core.NewRegistry()
	sender := &mockSender{}

	registry.Register(&SubscribeCommand{})
	registry.Register(&UnsubscribeCommand{})
	registry.Register(&ListCommand{})
	registry.Register(&MuteCommand{})

	cmd := &HelpCommand{sender: sender, registry: registry}
	cmd.Execute(context.Background(), 123, "")

	assert.Len(t, sender.sent, 1)
	assert.Contains(t, sender.sent[0], "GitNotify")
	assert.Contains(t, sender.sent[0], "subscribe")
	assert.Contains(t, sender.sent[0], "unsubscribe")
	assert.Contains(t, sender.sent[0], "list")
	assert.Contains(t, sender.sent[0], "mute")
}

func TestHelpCommand_Execute_Empty(t *testing.T) {
	registry := core.NewRegistry()
	sender := &mockSender{}

	cmd := &HelpCommand{sender: sender, registry: registry}
	cmd.Execute(context.Background(), 123, "")

	assert.Len(t, sender.sent, 1)
	assert.Contains(t, sender.sent[0], "GitNotify")
}

func TestHelpCommand_Name(t *testing.T) {
	cmd := &HelpCommand{}
	assert.Equal(t, "help", cmd.Name())
}

func TestHelpCommand_Description(t *testing.T) {
	cmd := &HelpCommand{}
	assert.Equal(t, "Show available commands", cmd.Description())
}

func TestHelpCommand_Usage(t *testing.T) {
	cmd := &HelpCommand{}
	assert.Equal(t, "/help", cmd.Usage())
}

func TestStartCommand_Execute(t *testing.T) {
	registry := core.NewRegistry()
	sender := &mockSender{}

	registry.Register(&SubscribeCommand{})
	registry.Register(&ListCommand{})

	cmd := &StartCommand{sender: sender, registry: registry}
	cmd.Execute(context.Background(), 123, "")

	assert.Len(t, sender.sent, 1)
	assert.Contains(t, sender.sent[0], "GitNotify")
}
