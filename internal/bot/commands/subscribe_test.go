package commands

import (
	"context"
	"testing"
	"time"

	"log/slog"
	"os"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/dlc-01/GitNotify/internal/bot/core"
	"github.com/dlc-01/GitNotify/internal/domain"
	internalkafka "github.com/dlc-01/GitNotify/internal/kafka"
	"github.com/dlc-01/GitNotify/internal/repository"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
}

func TestSubscribeCommand_Execute_Success(t *testing.T) {
	repo := &mockRepository{}
	prod := &mockMultiProducer{}
	sender := &mockSender{}
	log := newTestLogger()

	sub := &domain.Subscription{
		ID:        1,
		ChatID:    123,
		RepoURL:   "https://github.com/golang/go",
		CreatedAt: time.Now(),
	}

	repo.On("Subscribe", mock.Anything, int64(123), "https://github.com/golang/go").
		Return(sub, nil)
	prod.On("ProduceTo", mock.Anything, internalkafka.TopicSubscriptionCreated, internalkafka.SubscriptionCreatedMessage{
		ChatID:  123,
		RepoURL: "https://github.com/golang/go",
	}).Return(nil)

	cmd := &SubscribeCommand{repo: repo, sender: sender, log: log, producer: prod, topic: internalkafka.TopicSubscriptionCreated}
	cmd.Execute(context.Background(), 123, "https://github.com/golang/go")

	assert.Contains(t, sender.sent[0], "Subscribed to")
	repo.AssertExpectations(t)
	prod.AssertExpectations(t)
}

func TestSubscribeCommand_Execute_EmptyArgs(t *testing.T) {
	repo := &mockRepository{}
	prod := &mockMultiProducer{}
	sender := &mockSender{}
	log := newTestLogger()

	cmd := &SubscribeCommand{repo: repo, sender: sender, log: log, producer: prod, topic: internalkafka.TopicSubscriptionCreated}
	cmd.Execute(context.Background(), 123, "")

	assert.Contains(t, sender.sent[0], core.FormatError(core.Wrap("Execute", core.ErrEmptyArgs)))
	repo.AssertNotCalled(t, "Subscribe")
	prod.AssertNotCalled(t, "ProduceTo")
}

func TestSubscribeCommand_Execute_InvalidURL(t *testing.T) {
	repo := &mockRepository{}
	prod := &mockMultiProducer{}
	sender := &mockSender{}
	log := newTestLogger()

	cmd := &SubscribeCommand{repo: repo, sender: sender, log: log, producer: prod, topic: internalkafka.TopicSubscriptionCreated}
	cmd.Execute(context.Background(), 123, "https://bitbucket.org/user/repo")

	assert.Contains(t, sender.sent[0], core.FormatError(core.Wrap("Execute", core.ErrInvalidRepoURL)))
	repo.AssertNotCalled(t, "Subscribe")
	prod.AssertNotCalled(t, "ProduceTo")
}

func TestSubscribeCommand_Execute_AlreadyExists(t *testing.T) {
	repo := &mockRepository{}
	prod := &mockMultiProducer{}
	sender := &mockSender{}
	log := newTestLogger()

	repo.On("Subscribe", mock.Anything, int64(123), "https://github.com/golang/go").
		Return(nil, repository.Wrap("Subscribe", repository.ErrAlreadyExists))

	cmd := &SubscribeCommand{repo: repo, sender: sender, log: log, producer: prod, topic: internalkafka.TopicSubscriptionCreated}
	cmd.Execute(context.Background(), 123, "https://github.com/golang/go")

	assert.Contains(t, sender.sent[0], "Already subscribed")
	prod.AssertNotCalled(t, "ProduceTo")
	repo.AssertExpectations(t)
}

func TestSubscribeCommand_Execute_RepoError(t *testing.T) {
	repo := &mockRepository{}
	prod := &mockMultiProducer{}
	sender := &mockSender{}
	log := newTestLogger()

	repo.On("Subscribe", mock.Anything, int64(123), "https://github.com/golang/go").
		Return(nil, repository.Wrap("Subscribe", repository.ErrInvalidInput))

	cmd := &SubscribeCommand{repo: repo, sender: sender, log: log, producer: prod, topic: internalkafka.TopicSubscriptionCreated}
	cmd.Execute(context.Background(), 123, "https://github.com/golang/go")

	assert.Contains(t, sender.sent[0], core.FormatError(core.Wrap("Execute", core.ErrInternal)))
	prod.AssertNotCalled(t, "ProduceTo")
	repo.AssertExpectations(t)
}

func TestSubscribeCommand_Execute_ProduceError(t *testing.T) {
	repo := &mockRepository{}
	prod := &mockMultiProducer{}
	sender := &mockSender{}
	log := newTestLogger()

	sub := &domain.Subscription{
		ID:        1,
		ChatID:    123,
		RepoURL:   "https://github.com/golang/go",
		CreatedAt: time.Now(),
	}

	repo.On("Subscribe", mock.Anything, int64(123), "https://github.com/golang/go").
		Return(sub, nil)
	prod.On("ProduceTo", mock.Anything, mock.Anything, mock.Anything).
		Return(assert.AnError)

	cmd := &SubscribeCommand{repo: repo, sender: sender, log: log, producer: prod, topic: internalkafka.TopicSubscriptionCreated}
	cmd.Execute(context.Background(), 123, "https://github.com/golang/go")

	assert.Contains(t, sender.sent[0], "Subscribed to")
	repo.AssertExpectations(t)
	prod.AssertExpectations(t)
}

func TestSubscribeCommand_Name(t *testing.T) {
	cmd := &SubscribeCommand{}
	assert.Equal(t, "subscribe", cmd.Name())
}

func TestSubscribeCommand_Description(t *testing.T) {
	cmd := &SubscribeCommand{}
	assert.Equal(t, "Subscribe to a repository or resource", cmd.Description())
}

func TestSubscribeCommand_Usage(t *testing.T) {
	cmd := &SubscribeCommand{}
	assert.Equal(t, "/subscribe <url>", cmd.Usage())
}
