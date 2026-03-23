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

func TestMuteCommand_Execute_Success(t *testing.T) {
	repo := &mockRepository{}
	prod := &mockMultiProducer{}
	sender := &mockSender{}
	log := newTestLogger()

	repo.On("MuteEvent", mock.Anything, int64(123), "https://github.com/golang/go", domain.EventPush).
		Return(nil)
	prod.On("ProduceTo", mock.Anything, internalkafka.TopicSubscriptionMuted, internalkafka.SubscriptionMutedMessage{
		ChatID:  123,
		RepoURL: "https://github.com/golang/go",
		Event:   "push",
	}).Return(nil)

	cmd := &MuteCommand{repo: repo, sender: sender, log: log, producer: prod, topic: internalkafka.TopicSubscriptionMuted}
	cmd.Execute(context.Background(), 123, "https://github.com/golang/go push")

	assert.Contains(t, sender.sent[0], "Muted")
	assert.Contains(t, sender.sent[0], "push")
	repo.AssertExpectations(t)
	prod.AssertExpectations(t)
}

func TestMuteCommand_Execute_EmptyArgs(t *testing.T) {
	repo := &mockRepository{}
	prod := &mockMultiProducer{}
	sender := &mockSender{}
	log := newTestLogger()

	cmd := &MuteCommand{repo: repo, sender: sender, log: log, producer: prod, topic: internalkafka.TopicSubscriptionMuted}
	cmd.Execute(context.Background(), 123, "")

	assert.Contains(t, sender.sent[0], core.FormatError(core.Wrap("Execute", core.ErrEmptyArgs)))
	repo.AssertNotCalled(t, "MuteEvent")
	prod.AssertNotCalled(t, "ProduceTo")
}

func TestMuteCommand_Execute_OnlyOneArg(t *testing.T) {
	repo := &mockRepository{}
	prod := &mockMultiProducer{}
	sender := &mockSender{}
	log := newTestLogger()

	cmd := &MuteCommand{repo: repo, sender: sender, log: log, producer: prod, topic: internalkafka.TopicSubscriptionMuted}
	cmd.Execute(context.Background(), 123, "https://github.com/golang/go")

	assert.Contains(t, sender.sent[0], core.FormatError(core.Wrap("Execute", core.ErrEmptyArgs)))
	repo.AssertNotCalled(t, "MuteEvent")
	prod.AssertNotCalled(t, "ProduceTo")
}

func TestMuteCommand_Execute_InvalidEvent(t *testing.T) {
	repo := &mockRepository{}
	prod := &mockMultiProducer{}
	sender := &mockSender{}
	log := newTestLogger()

	cmd := &MuteCommand{repo: repo, sender: sender, log: log, producer: prod, topic: internalkafka.TopicSubscriptionMuted}
	cmd.Execute(context.Background(), 123, "https://github.com/golang/go invalidEvent")

	assert.Contains(t, sender.sent[0], core.FormatError(core.Wrap("Execute", core.ErrInvalidEvent)))
	repo.AssertNotCalled(t, "MuteEvent")
	prod.AssertNotCalled(t, "ProduceTo")
}

func TestMuteCommand_Execute_NotFound(t *testing.T) {
	repo := &mockRepository{}
	prod := &mockMultiProducer{}
	sender := &mockSender{}
	log := newTestLogger()

	repo.On("MuteEvent", mock.Anything, int64(123), "https://github.com/golang/go", domain.EventPush).
		Return(repository.Wrap("MuteEvent", repository.ErrNotFound))

	cmd := &MuteCommand{repo: repo, sender: sender, log: log, producer: prod, topic: internalkafka.TopicSubscriptionMuted}
	cmd.Execute(context.Background(), 123, "https://github.com/golang/go push")

	assert.Contains(t, sender.sent[0], "Subscription not found")
	prod.AssertNotCalled(t, "ProduceTo")
	repo.AssertExpectations(t)
}

func TestMuteCommand_Execute_RepoError(t *testing.T) {
	repo := &mockRepository{}
	prod := &mockMultiProducer{}
	sender := &mockSender{}
	log := newTestLogger()

	repo.On("MuteEvent", mock.Anything, int64(123), "https://github.com/golang/go", domain.EventPush).
		Return(repository.Wrap("MuteEvent", repository.ErrInvalidInput))

	cmd := &MuteCommand{repo: repo, sender: sender, log: log, producer: prod, topic: internalkafka.TopicSubscriptionMuted}
	cmd.Execute(context.Background(), 123, "https://github.com/golang/go push")

	assert.Contains(t, sender.sent[0], core.FormatError(core.Wrap("Execute", core.ErrInternal)))
	prod.AssertNotCalled(t, "ProduceTo")
	repo.AssertExpectations(t)
}

func TestMuteCommand_Execute_ProduceError(t *testing.T) {
	repo := &mockRepository{}
	prod := &mockMultiProducer{}
	sender := &mockSender{}
	log := newTestLogger()

	repo.On("MuteEvent", mock.Anything, int64(123), "https://github.com/golang/go", domain.EventPush).
		Return(nil)
	prod.On("ProduceTo", mock.Anything, mock.Anything, mock.Anything).
		Return(assert.AnError)

	cmd := &MuteCommand{repo: repo, sender: sender, log: log, producer: prod, topic: internalkafka.TopicSubscriptionMuted}
	cmd.Execute(context.Background(), 123, "https://github.com/golang/go push")

	assert.Contains(t, sender.sent[0], "Muted")
	repo.AssertExpectations(t)
	prod.AssertExpectations(t)
}
