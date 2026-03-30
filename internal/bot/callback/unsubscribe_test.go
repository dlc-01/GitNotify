package callback

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	internalkafka "github.com/dlc-01/GitNotify/internal/kafka"
	"github.com/dlc-01/GitNotify/internal/repository"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
}

func TestUnsubscribeHandler_Execute_Success(t *testing.T) {
	repo := &mockRepository{}
	sender := &mockSender{}
	prod := &mockMultiProducer{}
	log := newTestLogger()

	repo.On("Unsubscribe", mock.Anything, int64(123), "https://github.com/golang/go").Return(nil)
	prod.On("ProduceTo", mock.Anything, internalkafka.TopicSubscriptionDeleted, mock.Anything).Return(nil)

	h := NewUnsubscribeHandler(repo, sender, log, prod)
	h.Execute(context.Background(), 123, 1, "https://github.com/golang/go")

	assert.Contains(t, sender.edited[0], "Unsubscribed from")
	repo.AssertExpectations(t)
	prod.AssertExpectations(t)
}

func TestUnsubscribeHandler_Execute_NotFound(t *testing.T) {
	repo := &mockRepository{}
	sender := &mockSender{}
	prod := &mockMultiProducer{}
	log := newTestLogger()

	repo.On("Unsubscribe", mock.Anything, int64(123), "https://github.com/golang/go").
		Return(repository.Wrap("Unsubscribe", repository.ErrNotFound))

	h := NewUnsubscribeHandler(repo, sender, log, prod)
	h.Execute(context.Background(), 123, 1, "https://github.com/golang/go")

	assert.Contains(t, sender.answered[0], "Subscription not found")
	assert.Empty(t, sender.edited)
	prod.AssertNotCalled(t, "ProduceTo")
	repo.AssertExpectations(t)
}

func TestUnsubscribeHandler_Execute_RepoError(t *testing.T) {
	repo := &mockRepository{}
	sender := &mockSender{}
	prod := &mockMultiProducer{}
	log := newTestLogger()

	repo.On("Unsubscribe", mock.Anything, int64(123), "https://github.com/golang/go").
		Return(repository.Wrap("Unsubscribe", repository.ErrInvalidInput))

	h := NewUnsubscribeHandler(repo, sender, log, prod)
	h.Execute(context.Background(), 123, 1, "https://github.com/golang/go")

	assert.Contains(t, sender.answered[0], "Internal error")
	assert.Empty(t, sender.edited)
	prod.AssertNotCalled(t, "ProduceTo")
	repo.AssertExpectations(t)
}

func TestUnsubscribeHandler_Execute_ProduceError(t *testing.T) {
	repo := &mockRepository{}
	sender := &mockSender{}
	prod := &mockMultiProducer{}
	log := newTestLogger()

	repo.On("Unsubscribe", mock.Anything, int64(123), "https://github.com/golang/go").Return(nil)
	prod.On("ProduceTo", mock.Anything, mock.Anything, mock.Anything).Return(assert.AnError)

	h := NewUnsubscribeHandler(repo, sender, log, prod)
	h.Execute(context.Background(), 123, 1, "https://github.com/golang/go")

	assert.Contains(t, sender.edited[0], "Unsubscribed from")
	repo.AssertExpectations(t)
	prod.AssertExpectations(t)
}

func TestUnsubscribeHandler_Action(t *testing.T) {
	h := &UnsubscribeHandler{}
	assert.Equal(t, "unsubscribe", h.Action())
}
