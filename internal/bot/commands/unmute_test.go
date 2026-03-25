package commands

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/dlc-01/GitNotify/internal/bot/core"
	"github.com/dlc-01/GitNotify/internal/domain"
	internalkafka "github.com/dlc-01/GitNotify/internal/kafka"
	"github.com/dlc-01/GitNotify/internal/repository"
)

func TestUnmuteCommand_Execute_Success(t *testing.T) {
	repo := &mockRepository{}
	prod := &mockMultiProducer{}
	sender := &mockSender{}
	log := newTestLogger()

	repo.On("UnmuteEvent", mock.Anything, int64(123), "https://github.com/golang/go", domain.EventPush).
		Return(nil)
	prod.On("ProduceTo", mock.Anything, internalkafka.TopicSubscriptionUnmuted, internalkafka.SubscriptionUnmutedMessage{
		ChatID:  123,
		RepoURL: "https://github.com/golang/go",
		Event:   "push",
	}).Return(nil)

	cmd := &UnmuteCommand{repo: repo, sender: sender, log: log, producer: prod, topic: internalkafka.TopicSubscriptionUnmuted}
	cmd.Execute(context.Background(), 123, "https://github.com/golang/go push")

	assert.Contains(t, sender.sent[0], "Unmuted")
	assert.Contains(t, sender.sent[0], "push")
	repo.AssertExpectations(t)
	prod.AssertExpectations(t)
}

func TestUnmuteCommand_Execute_EmptyArgs(t *testing.T) {
	repo := &mockRepository{}
	prod := &mockMultiProducer{}
	sender := &mockSender{}
	log := newTestLogger()

	cmd := &UnmuteCommand{repo: repo, sender: sender, log: log, producer: prod, topic: internalkafka.TopicSubscriptionUnmuted}
	cmd.Execute(context.Background(), 123, "")

	assert.Contains(t, sender.sent[0], core.FormatError(core.Wrap("Execute", core.ErrEmptyArgs)))
	repo.AssertNotCalled(t, "UnmuteEvent")
	prod.AssertNotCalled(t, "ProduceTo")
}

func TestUnmuteCommand_Execute_OnlyOneArg(t *testing.T) {
	repo := &mockRepository{}
	prod := &mockMultiProducer{}
	sender := &mockSender{}
	log := newTestLogger()

	cmd := &UnmuteCommand{repo: repo, sender: sender, log: log, producer: prod, topic: internalkafka.TopicSubscriptionUnmuted}
	cmd.Execute(context.Background(), 123, "https://github.com/golang/go")

	assert.Contains(t, sender.sent[0], core.FormatError(core.Wrap("Execute", core.ErrEmptyArgs)))
	repo.AssertNotCalled(t, "UnmuteEvent")
	prod.AssertNotCalled(t, "ProduceTo")
}

func TestUnmuteCommand_Execute_InvalidEvent(t *testing.T) {
	repo := &mockRepository{}
	prod := &mockMultiProducer{}
	sender := &mockSender{}
	log := newTestLogger()

	cmd := &UnmuteCommand{repo: repo, sender: sender, log: log, producer: prod, topic: internalkafka.TopicSubscriptionUnmuted}
	cmd.Execute(context.Background(), 123, "https://github.com/golang/go invalidEvent")

	assert.Contains(t, sender.sent[0], core.FormatError(core.Wrap("Execute", core.ErrInvalidEvent)))
	repo.AssertNotCalled(t, "UnmuteEvent")
	prod.AssertNotCalled(t, "ProduceTo")
}

func TestUnmuteCommand_Execute_NotFound(t *testing.T) {
	repo := &mockRepository{}
	prod := &mockMultiProducer{}
	sender := &mockSender{}
	log := newTestLogger()

	repo.On("UnmuteEvent", mock.Anything, int64(123), "https://github.com/golang/go", domain.EventPush).
		Return(repository.Wrap("UnmuteEvent", repository.ErrNotFound))

	cmd := &UnmuteCommand{repo: repo, sender: sender, log: log, producer: prod, topic: internalkafka.TopicSubscriptionUnmuted}
	cmd.Execute(context.Background(), 123, "https://github.com/golang/go push")

	assert.Contains(t, sender.sent[0], "Subscription not found")
	prod.AssertNotCalled(t, "ProduceTo")
	repo.AssertExpectations(t)
}

func TestUnmuteCommand_Execute_RepoError(t *testing.T) {
	repo := &mockRepository{}
	prod := &mockMultiProducer{}
	sender := &mockSender{}
	log := newTestLogger()

	repo.On("UnmuteEvent", mock.Anything, int64(123), "https://github.com/golang/go", domain.EventPush).
		Return(repository.Wrap("UnmuteEvent", repository.ErrInvalidInput))

	cmd := &UnmuteCommand{repo: repo, sender: sender, log: log, producer: prod, topic: internalkafka.TopicSubscriptionUnmuted}
	cmd.Execute(context.Background(), 123, "https://github.com/golang/go push")

	assert.Contains(t, sender.sent[0], core.FormatError(core.Wrap("Execute", core.ErrInternal)))
	prod.AssertNotCalled(t, "ProduceTo")
	repo.AssertExpectations(t)
}

func TestUnmuteCommand_Execute_ProduceError(t *testing.T) {
	repo := &mockRepository{}
	prod := &mockMultiProducer{}
	sender := &mockSender{}
	log := newTestLogger()

	repo.On("UnmuteEvent", mock.Anything, int64(123), "https://github.com/golang/go", domain.EventPush).
		Return(nil)
	prod.On("ProduceTo", mock.Anything, mock.Anything, mock.Anything).
		Return(assert.AnError)

	cmd := &UnmuteCommand{repo: repo, sender: sender, log: log, producer: prod, topic: internalkafka.TopicSubscriptionUnmuted}
	cmd.Execute(context.Background(), 123, "https://github.com/golang/go push")

	assert.Contains(t, sender.sent[0], "Unmuted")
	repo.AssertExpectations(t)
	prod.AssertExpectations(t)
}

func TestUnmuteCommand_Name(t *testing.T) {
	cmd := &UnmuteCommand{}
	assert.Equal(t, "unmute", cmd.Name())
}

func TestUnmuteCommand_Description(t *testing.T) {
	cmd := &UnmuteCommand{}
	assert.Equal(t, "Unmute an event type for a resource", cmd.Description())
}

func TestUnmuteCommand_Usage(t *testing.T) {
	cmd := &UnmuteCommand{}
	assert.Equal(t, "/unmute <url> <event>", cmd.Usage())
}
