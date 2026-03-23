package callback

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

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
	log := newTestLogger()

	repo.On("Unsubscribe", mock.Anything, int64(123), "https://github.com/golang/go").
		Return(nil)

	h := NewUnsubscribeHandler(repo, sender, log)
	h.Execute(context.Background(), 123, 1, "https://github.com/golang/go")

	assert.Contains(t, sender.edited[0], "Unsubscribed from")
	repo.AssertExpectations(t)
}

func TestUnsubscribeHandler_Execute_NotFound(t *testing.T) {
	repo := &mockRepository{}
	sender := &mockSender{}
	log := newTestLogger()

	repo.On("Unsubscribe", mock.Anything, int64(123), "https://github.com/golang/go").
		Return(repository.Wrap("Unsubscribe", repository.ErrNotFound))

	h := NewUnsubscribeHandler(repo, sender, log)
	h.Execute(context.Background(), 123, 1, "https://github.com/golang/go")

	assert.Contains(t, sender.answered[0], "Subscription not found")
	assert.Empty(t, sender.edited)
	repo.AssertExpectations(t)
}

func TestUnsubscribeHandler_Execute_RepoError(t *testing.T) {
	repo := &mockRepository{}
	sender := &mockSender{}
	log := newTestLogger()

	repo.On("Unsubscribe", mock.Anything, int64(123), "https://github.com/golang/go").
		Return(repository.Wrap("Unsubscribe", repository.ErrInvalidInput))

	h := NewUnsubscribeHandler(repo, sender, log)
	h.Execute(context.Background(), 123, 1, "https://github.com/golang/go")

	assert.Contains(t, sender.answered[0], "Internal error")
	assert.Empty(t, sender.edited)
	repo.AssertExpectations(t)
}

func TestUnsubscribeHandler_Action(t *testing.T) {
	h := &UnsubscribeHandler{}
	assert.Equal(t, "unsubscribe", h.Action())
}
