package repository

import (
	"context"

	"github.com/dlc-01/GitNotify/internal/domain"
)

type Repository interface {
	UpsertUser(ctx context.Context, user *domain.User) error

	UpsertChat(ctx context.Context, chat *domain.Chat) error

	Subscribe(ctx context.Context, chatID int64, repoURL string) (*domain.Subscription, error)
	Unsubscribe(ctx context.Context, chatID int64, repoURL string) error
	ListSubscriptions(ctx context.Context, chatID int64) ([]*domain.Subscription, error)
	MuteEvent(ctx context.Context, chatID int64, repoURL string, event domain.EventType) error
}
