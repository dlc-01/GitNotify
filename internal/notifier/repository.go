package notifier

import (
	"context"

	"github.com/dlc-01/GitNotify/internal/domain"
)

type Repository interface {
	Subscribe(ctx context.Context, chatID int64, repoURL string) error
	Unsubscribe(ctx context.Context, chatID int64, repoURL string) error
	MuteEvent(ctx context.Context, chatID int64, repoURL string, event domain.EventType) error
	UnmuteEvent(ctx context.Context, chatID int64, repoURL string, event domain.EventType) error
	GetSubscribersByRepo(ctx context.Context, repoURL string, event domain.EventType) ([]int64, error)
}
