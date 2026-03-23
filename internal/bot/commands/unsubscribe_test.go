package commands

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/dlc-01/GitNotify/internal/bot/core"
	internalkafka "github.com/dlc-01/GitNotify/internal/kafka"
	"github.com/dlc-01/GitNotify/internal/repository"
)

func TestUnsubscribeCommand_Execute_Success(t *testing.T) {
	repo := &mockRepository{}
	prod := &mockMultiProducer{}
	sender := &mockSender{}
	log := newTestLogger()

	repo.On("Unsubscribe", mock.Anything, int64(123), "https://github.com/golang/go").
		Return(nil)
	prod.On("ProduceTo", mock.Anything, internalkafka.TopicSubscriptionDeleted, internalkafka.SubscriptionDeletedMessage{
		ChatID:  123,
		RepoURL: "https://github.com/golang/go",
	}).Return(nil)

	cmd := &UnsubscribeCommand{repo: repo, sender: sender, log: log, producer: prod, topic: internalkafka.TopicSubscriptionDeleted}
	cmd.Execute(context.Background(), 123, "https://github.com/golang/go")

	assert.Contains(t, sender.sent[0], "Unsubscribed from")
	repo.AssertExpectations(t)
	prod.AssertExpectations(t)
}

func TestUnsubscribeCommand_Execute_EmptyArgs(t *testing.T) {
	repo := &mockRepository{}
	prod := &mockMultiProducer{}
	sender := &mockSender{}
	log := newTestLogger()

	cmd := &UnsubscribeCommand{repo: repo, sender: sender, log: log, producer: prod, topic: internalkafka.TopicSubscriptionDeleted}
	cmd.Execute(context.Background(), 123, "")

	assert.Contains(t, sender.sent[0], core.FormatError(core.Wrap("Execute", core.ErrEmptyArgs)))
	repo.AssertNotCalled(t, "Unsubscribe")
	prod.AssertNotCalled(t, "ProduceTo")
}

func TestUnsubscribeCommand_Execute_NotFound(t *testing.T) {
	repo := &mockRepository{}
	prod := &mockMultiProducer{}
	sender := &mockSender{}
	log := newTestLogger()

	repo.On("Unsubscribe", mock.Anything, int64(123), "https://github.com/golang/go").
		Return(repository.Wrap("Unsubscribe", repository.ErrNotFound))

	cmd := &UnsubscribeCommand{repo: repo, sender: sender, log: log, producer: prod, topic: internalkafka.TopicSubscriptionDeleted}
	cmd.Execute(context.Background(), 123, "https://github.com/golang/go")

	assert.Contains(t, sender.sent[0], "Subscription not found")
	prod.AssertNotCalled(t, "ProduceTo")
	repo.AssertExpectations(t)
}

func TestUnsubscribeCommand_Execute_RepoError(t *testing.T) {
	repo := &mockRepository{}
	prod := &mockMultiProducer{}
	sender := &mockSender{}
	log := newTestLogger()

	repo.On("Unsubscribe", mock.Anything, int64(123), "https://github.com/golang/go").
		Return(repository.Wrap("Unsubscribe", repository.ErrInvalidInput))

	cmd := &UnsubscribeCommand{repo: repo, sender: sender, log: log, producer: prod, topic: internalkafka.TopicSubscriptionDeleted}
	cmd.Execute(context.Background(), 123, "https://github.com/golang/go")

	assert.Contains(t, sender.sent[0], core.FormatError(core.Wrap("Execute", core.ErrInternal)))
	prod.AssertNotCalled(t, "ProduceTo")
	repo.AssertExpectations(t)
}

func TestUnsubscribeCommand_Execute_ProduceError(t *testing.T) {
	repo := &mockRepository{}
	prod := &mockMultiProducer{}
	sender := &mockSender{}
	log := newTestLogger()

	repo.On("Unsubscribe", mock.Anything, int64(123), "https://github.com/golang/go").
		Return(nil)
	prod.On("ProduceTo", mock.Anything, mock.Anything, mock.Anything).
		Return(assert.AnError)

	cmd := &UnsubscribeCommand{repo: repo, sender: sender, log: log, producer: prod, topic: internalkafka.TopicSubscriptionDeleted}
	cmd.Execute(context.Background(), 123, "https://github.com/golang/go")

	assert.Contains(t, sender.sent[0], "Unsubscribed from")
	repo.AssertExpectations(t)
	prod.AssertExpectations(t)
}
