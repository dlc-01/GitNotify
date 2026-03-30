package notifier

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/dlc-01/GitNotify/internal/domain"
	internalkafka "github.com/dlc-01/GitNotify/internal/kafka"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
}

type mockRepository struct {
	mock.Mock
}

func (m *mockRepository) Subscribe(ctx context.Context, chatID int64, repoURL string) error {
	args := m.Called(ctx, chatID, repoURL)
	return args.Error(0)
}

func (m *mockRepository) Unsubscribe(ctx context.Context, chatID int64, repoURL string) error {
	args := m.Called(ctx, chatID, repoURL)
	return args.Error(0)
}

func (m *mockRepository) MuteEvent(ctx context.Context, chatID int64, repoURL string, event domain.EventType) error {
	args := m.Called(ctx, chatID, repoURL, event)
	return args.Error(0)
}

func (m *mockRepository) UnmuteEvent(ctx context.Context, chatID int64, repoURL string, event domain.EventType) error {
	args := m.Called(ctx, chatID, repoURL, event)
	return args.Error(0)
}

func (m *mockRepository) GetSubscribersByRepo(ctx context.Context, repoURL string, event domain.EventType) ([]int64, error) {
	args := m.Called(ctx, repoURL, event)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]int64), args.Error(1)
}

type mockSender struct {
	mock.Mock
}

func (m *mockSender) Send(ctx context.Context, chatID int64, text string) error {
	args := m.Called(ctx, chatID, text)
	return args.Error(0)
}

func TestHandler_HandleEvent_Success(t *testing.T) {
	repo := &mockRepository{}
	sender := &mockSender{}
	log := newTestLogger()

	msg := internalkafka.WebhookEventMessage{
		RepoURL:   "https://github.com/golang/go",
		EventType: "push",
		Source:    "github",
	}
	data, _ := json.Marshal(msg)

	repo.On("GetSubscribersByRepo", mock.Anything, "https://github.com/golang/go", domain.EventPush).
		Return([]int64{123, 456}, nil)
	sender.On("Send", mock.Anything, int64(123), mock.Anything).Return(nil)
	sender.On("Send", mock.Anything, int64(456), mock.Anything).Return(nil)

	h := NewHandler(repo, sender, log)
	err := h.HandleEvent(context.Background(), data)

	require.NoError(t, err)
	repo.AssertExpectations(t)
	sender.AssertExpectations(t)
}

func TestHandler_HandleEvent_NoSubscribers(t *testing.T) {
	repo := &mockRepository{}
	sender := &mockSender{}
	log := newTestLogger()

	msg := internalkafka.WebhookEventMessage{
		RepoURL:   "https://github.com/golang/go",
		EventType: "push",
		Source:    "github",
	}
	data, _ := json.Marshal(msg)

	repo.On("GetSubscribersByRepo", mock.Anything, "https://github.com/golang/go", domain.EventPush).
		Return([]int64{}, nil)

	h := NewHandler(repo, sender, log)
	err := h.HandleEvent(context.Background(), data)

	require.NoError(t, err)
	sender.AssertNotCalled(t, "Send")
	repo.AssertExpectations(t)
}

func TestHandler_HandleSubscriptionCreated_Success(t *testing.T) {
	repo := &mockRepository{}
	sender := &mockSender{}
	log := newTestLogger()

	msg := internalkafka.SubscriptionCreatedMessage{
		ChatID:  123,
		RepoURL: "https://github.com/golang/go",
	}
	data, _ := json.Marshal(msg)

	repo.On("Subscribe", mock.Anything, int64(123), "https://github.com/golang/go").
		Return(nil)

	h := NewHandler(repo, sender, log)
	err := h.HandleSubscriptionCreated(context.Background(), data)

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestHandler_HandleSubscriptionCreated_InvalidJSON(t *testing.T) {
	repo := &mockRepository{}
	sender := &mockSender{}
	log := newTestLogger()

	h := NewHandler(repo, sender, log)
	err := h.HandleSubscriptionCreated(context.Background(), []byte("invalid json"))

	require.Error(t, err)
	var notifierErr *Error
	require.ErrorAs(t, err, &notifierErr)
	assert.ErrorIs(t, notifierErr, ErrUnmarshal)
}

func TestHandler_HandleSubscriptionCreated_RepoError(t *testing.T) {
	repo := &mockRepository{}
	sender := &mockSender{}
	log := newTestLogger()

	msg := internalkafka.SubscriptionCreatedMessage{
		ChatID:  123,
		RepoURL: "https://github.com/golang/go",
	}
	data, _ := json.Marshal(msg)

	repo.On("Subscribe", mock.Anything, int64(123), "https://github.com/golang/go").
		Return(assert.AnError)

	h := NewHandler(repo, sender, log)
	err := h.HandleSubscriptionCreated(context.Background(), data)

	require.Error(t, err)
	repo.AssertExpectations(t)
}

func TestHandler_HandleSubscriptionDeleted_Success(t *testing.T) {
	repo := &mockRepository{}
	sender := &mockSender{}
	log := newTestLogger()

	msg := internalkafka.SubscriptionDeletedMessage{
		ChatID:  123,
		RepoURL: "https://github.com/golang/go",
	}
	data, _ := json.Marshal(msg)

	repo.On("Unsubscribe", mock.Anything, int64(123), "https://github.com/golang/go").
		Return(nil)

	h := NewHandler(repo, sender, log)
	err := h.HandleSubscriptionDeleted(context.Background(), data)

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestHandler_HandleSubscriptionDeleted_NotFound(t *testing.T) {
	repo := &mockRepository{}
	sender := &mockSender{}
	log := newTestLogger()

	msg := internalkafka.SubscriptionDeletedMessage{
		ChatID:  123,
		RepoURL: "https://github.com/golang/go",
	}
	data, _ := json.Marshal(msg)

	repo.On("Unsubscribe", mock.Anything, int64(123), "https://github.com/golang/go").
		Return(ErrNotFound)

	h := NewHandler(repo, sender, log)
	err := h.HandleSubscriptionDeleted(context.Background(), data)

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestHandler_HandleSubscriptionDeleted_InvalidJSON(t *testing.T) {
	repo := &mockRepository{}
	sender := &mockSender{}
	log := newTestLogger()

	h := NewHandler(repo, sender, log)
	err := h.HandleSubscriptionDeleted(context.Background(), []byte("invalid json"))

	require.Error(t, err)
	var notifierErr *Error
	require.ErrorAs(t, err, &notifierErr)
	assert.ErrorIs(t, notifierErr, ErrUnmarshal)
}

func TestHandler_HandleSubscriptionMuted_Success(t *testing.T) {
	repo := &mockRepository{}
	sender := &mockSender{}
	log := newTestLogger()

	msg := internalkafka.SubscriptionMutedMessage{
		ChatID:  123,
		RepoURL: "https://github.com/golang/go",
		Event:   "push",
	}
	data, _ := json.Marshal(msg)

	repo.On("MuteEvent", mock.Anything, int64(123), "https://github.com/golang/go", domain.EventPush).
		Return(nil)

	h := NewHandler(repo, sender, log)
	err := h.HandleSubscriptionMuted(context.Background(), data)

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestHandler_HandleSubscriptionMuted_InvalidJSON(t *testing.T) {
	repo := &mockRepository{}
	sender := &mockSender{}
	log := newTestLogger()

	h := NewHandler(repo, sender, log)
	err := h.HandleSubscriptionMuted(context.Background(), []byte("invalid json"))

	require.Error(t, err)
	var notifierErr *Error
	require.ErrorAs(t, err, &notifierErr)
	assert.ErrorIs(t, notifierErr, ErrUnmarshal)
}

func TestHandler_HandleSubscriptionUnmuted_Success(t *testing.T) {
	repo := &mockRepository{}
	sender := &mockSender{}
	log := newTestLogger()

	msg := internalkafka.SubscriptionUnmutedMessage{
		ChatID:  123,
		RepoURL: "https://github.com/golang/go",
		Event:   "push",
	}
	data, _ := json.Marshal(msg)

	repo.On("UnmuteEvent", mock.Anything, int64(123), "https://github.com/golang/go", domain.EventPush).
		Return(nil)

	h := NewHandler(repo, sender, log)
	err := h.HandleSubscriptionUnmuted(context.Background(), data)

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestHandler_HandleSubscriptionUnmuted_InvalidJSON(t *testing.T) {
	repo := &mockRepository{}
	sender := &mockSender{}
	log := newTestLogger()

	h := NewHandler(repo, sender, log)
	err := h.HandleSubscriptionUnmuted(context.Background(), []byte("invalid json"))

	require.Error(t, err)
	var notifierErr *Error
	require.ErrorAs(t, err, &notifierErr)
	assert.ErrorIs(t, notifierErr, ErrUnmarshal)
}

func TestHandler_HandleEvent_InvalidJSON(t *testing.T) {
	repo := &mockRepository{}
	sender := &mockSender{}
	log := newTestLogger()

	h := NewHandler(repo, sender, log)
	err := h.HandleEvent(context.Background(), []byte("invalid json"))

	require.Error(t, err)
	var notifierErr *Error
	require.ErrorAs(t, err, &notifierErr)
	assert.ErrorIs(t, notifierErr, ErrUnmarshal)
}

func TestHandler_HandleEvent_RepoError(t *testing.T) {
	repo := &mockRepository{}
	sender := &mockSender{}
	log := newTestLogger()

	msg := internalkafka.WebhookEventMessage{
		RepoURL:   "https://github.com/golang/go",
		EventType: "push",
		Source:    "github",
	}
	data, _ := json.Marshal(msg)

	repo.On("GetSubscribersByRepo", mock.Anything, "https://github.com/golang/go", domain.EventPush).
		Return(nil, assert.AnError)

	h := NewHandler(repo, sender, log)
	err := h.HandleEvent(context.Background(), data)

	require.Error(t, err)
	repo.AssertExpectations(t)
}

func TestHandler_HandleEvent_SendError(t *testing.T) {
	repo := &mockRepository{}
	sender := &mockSender{}
	log := newTestLogger()

	msg := internalkafka.WebhookEventMessage{
		RepoURL:   "https://github.com/golang/go",
		EventType: "push",
		Source:    "github",
	}
	data, _ := json.Marshal(msg)

	repo.On("GetSubscribersByRepo", mock.Anything, "https://github.com/golang/go", domain.EventPush).
		Return([]int64{123}, nil)
	sender.On("Send", mock.Anything, int64(123), mock.Anything).Return(assert.AnError)

	h := NewHandler(repo, sender, log)
	err := h.HandleEvent(context.Background(), data)

	require.NoError(t, err)
	repo.AssertExpectations(t)
	sender.AssertExpectations(t)
}

func TestFormatEventMessage(t *testing.T) {
	tests := []struct {
		eventType string
		expected  string
	}{
		{"push", "🔔 New push to https://github.com/golang/go"},
		{"pr", "🔔 New pull request in https://github.com/golang/go"},
		{"issue", "🔔 New issue in https://github.com/golang/go"},
		{"pipeline", "🔔 Pipeline triggered in https://github.com/golang/go"},
		{"answer", "🔔 New answer on https://github.com/golang/go"},
		{"post", "🔔 New post on https://github.com/golang/go"},
		{"video", "🔔 New video on https://github.com/golang/go"},
		{"unknown", "🔔 New event on https://github.com/golang/go"},
	}

	for _, tt := range tests {
		t.Run(tt.eventType, func(t *testing.T) {
			msg := internalkafka.WebhookEventMessage{
				RepoURL:   "https://github.com/golang/go",
				EventType: tt.eventType,
			}
			assert.Equal(t, tt.expected, formatEventMessage(msg))
		})
	}
}
