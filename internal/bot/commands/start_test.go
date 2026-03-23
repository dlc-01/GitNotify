package commands

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dlc-01/GitNotify/internal/bot/core"
)

func TestStartCommand_Execute_ContainsHelp(t *testing.T) {
	registry := core.NewRegistry()
	sender := &mockSender{}

	registry.Register(&SubscribeCommand{})
	registry.Register(&UnsubscribeCommand{})
	registry.Register(&ListCommand{})
	registry.Register(&MuteCommand{})

	cmd := &StartCommand{sender: sender, registry: registry}
	cmd.Execute(context.Background(), 123, "")

	assert.Len(t, sender.sent, 1)
	assert.Contains(t, sender.sent[0], "GitNotify")
	assert.Contains(t, sender.sent[0], "subscribe")
	assert.Contains(t, sender.sent[0], "unsubscribe")
	assert.Contains(t, sender.sent[0], "list")
	assert.Contains(t, sender.sent[0], "mute")
}

func TestStartCommand_Execute_Empty(t *testing.T) {
	registry := core.NewRegistry()
	sender := &mockSender{}

	cmd := &StartCommand{sender: sender, registry: registry}
	cmd.Execute(context.Background(), 123, "")

	assert.Len(t, sender.sent, 1)
	assert.Contains(t, sender.sent[0], "GitNotify")
}

func TestStartCommand_Name(t *testing.T) {
	cmd := &StartCommand{}
	assert.Equal(t, "start", cmd.Name())
}

func TestStartCommand_Description(t *testing.T) {
	cmd := &StartCommand{}
	assert.Equal(t, "Start the bot", cmd.Description())
}

func TestStartCommand_Usage(t *testing.T) {
	cmd := &StartCommand{}
	assert.Equal(t, "/start", cmd.Usage())
}
