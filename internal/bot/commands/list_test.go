package commands

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/dlc-01/GitNotify/internal/bot/core"
	"github.com/dlc-01/GitNotify/internal/domain"
	"github.com/dlc-01/GitNotify/internal/repository"
)

func TestListCommand_Execute_Empty(t *testing.T) {
	repo := &mockRepository{}
	sender := &mockSender{}
	log := newTestLogger()

	repo.On("ListSubscriptions", mock.Anything, int64(123)).
		Return([]*domain.Subscription{}, nil)

	cmd := &ListCommand{repo: repo, sender: sender, log: log}
	cmd.Execute(context.Background(), 123, "")

	assert.Contains(t, sender.sent[0], "no subscriptions")
	repo.AssertExpectations(t)
}

func TestListCommand_Execute_WithSubscriptions(t *testing.T) {
	repo := &mockRepository{}
	sender := &mockSender{}
	log := newTestLogger()

	subs := []*domain.Subscription{
		{
			ID:          1,
			ChatID:      123,
			RepoURL:     "https://github.com/golang/go",
			MutedEvents: []domain.EventType{},
			CreatedAt:   time.Now(),
		},
		{
			ID:          2,
			ChatID:      123,
			RepoURL:     "https://github.com/torvalds/linux",
			MutedEvents: []domain.EventType{domain.EventPush},
			CreatedAt:   time.Now(),
		},
	}

	repo.On("ListSubscriptions", mock.Anything, int64(123)).
		Return(subs, nil)

	cmd := &ListCommand{repo: repo, sender: sender, log: log}
	cmd.Execute(context.Background(), 123, "")

	assert.Len(t, sender.sent, 2)
	assert.Contains(t, sender.sent[0], "https://github.com/golang/go")
	assert.Contains(t, sender.sent[1], "https://github.com/torvalds/linux")
	repo.AssertExpectations(t)
}

func TestListCommand_Execute_WithMutedEvents(t *testing.T) {
	repo := &mockRepository{}
	sender := &mockSender{}
	log := newTestLogger()

	subs := []*domain.Subscription{
		{
			ID:          1,
			ChatID:      123,
			RepoURL:     "https://github.com/golang/go",
			MutedEvents: []domain.EventType{domain.EventPush, domain.EventPR},
			CreatedAt:   time.Now(),
		},
	}

	repo.On("ListSubscriptions", mock.Anything, int64(123)).
		Return(subs, nil)

	cmd := &ListCommand{repo: repo, sender: sender, log: log}
	cmd.Execute(context.Background(), 123, "")

	assert.Contains(t, sender.sent[0], "muted")
	assert.Contains(t, sender.sent[0], "push")
	assert.Contains(t, sender.sent[0], "pr")
	repo.AssertExpectations(t)
}

func TestListCommand_Execute_RepoError(t *testing.T) {
	repo := &mockRepository{}
	sender := &mockSender{}
	log := newTestLogger()

	repo.On("ListSubscriptions", mock.Anything, int64(123)).
		Return(nil, repository.Wrap("ListSubscriptions", repository.ErrInvalidInput))

	cmd := &ListCommand{repo: repo, sender: sender, log: log}
	cmd.Execute(context.Background(), 123, "")

	assert.Contains(t, sender.sent[0], core.FormatError(core.Wrap("Execute", core.ErrInternal)))
	repo.AssertExpectations(t)
}
