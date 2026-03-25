package repository

import (
	"context"
	"log/slog"
	"time"

	"github.com/dlc-01/GitNotify/internal/domain"
)

type LoggingRepository struct {
	repo Repository
	log  *slog.Logger
}

func NewLoggingRepository(repo Repository, log *slog.Logger) Repository {
	return &LoggingRepository{repo: repo, log: log}
}

func (r *LoggingRepository) UpsertUser(ctx context.Context, user *domain.User) error {
	start := time.Now()
	err := r.repo.UpsertUser(ctx, user)
	r.log.Debug("UpsertUser",
		slog.Group("user",
			slog.Int64("id", user.UserID),
			slog.String("username", user.Username),
		),
		slog.Duration("duration", time.Since(start)),
		slog.Any("err", err),
	)
	return err
}

func (r *LoggingRepository) UpsertChat(ctx context.Context, chat *domain.Chat) error {
	start := time.Now()
	err := r.repo.UpsertChat(ctx, chat)
	r.log.Debug("UpsertChat",
		slog.Group("chat",
			slog.Int64("id", chat.ChatID),
			slog.String("type", string(chat.ChatType)),
		),
		slog.Duration("duration", time.Since(start)),
		slog.Any("err", err),
	)
	return err
}

func (r *LoggingRepository) Subscribe(ctx context.Context, chatID int64, repoURL string) (*domain.Subscription, error) {
	start := time.Now()
	sub, err := r.repo.Subscribe(ctx, chatID, repoURL)
	r.log.Debug("Subscribe",
		slog.Group("chat",
			slog.Int64("id", chatID),
		),
		slog.String("repo", repoURL),
		slog.Duration("duration", time.Since(start)),
		slog.Any("err", err),
	)
	return sub, err
}

func (r *LoggingRepository) Unsubscribe(ctx context.Context, chatID int64, repoURL string) error {
	start := time.Now()
	err := r.repo.Unsubscribe(ctx, chatID, repoURL)
	r.log.Debug("Unsubscribe",
		slog.Group("chat",
			slog.Int64("id", chatID),
		),
		slog.String("repo", repoURL),
		slog.Duration("duration", time.Since(start)),
		slog.Any("err", err),
	)
	return err
}

func (r *LoggingRepository) ListSubscriptions(ctx context.Context, chatID int64) ([]*domain.Subscription, error) {
	start := time.Now()
	subs, err := r.repo.ListSubscriptions(ctx, chatID)
	r.log.Debug("ListSubscriptions",
		slog.Group("chat",
			slog.Int64("id", chatID),
		),
		slog.Int("count", len(subs)),
		slog.Duration("duration", time.Since(start)),
		slog.Any("err", err),
	)
	return subs, err
}

func (r *LoggingRepository) MuteEvent(ctx context.Context, chatID int64, repoURL string, event domain.EventType) error {
	start := time.Now()
	err := r.repo.MuteEvent(ctx, chatID, repoURL, event)
	r.log.Debug("MuteEvent",
		slog.Group("chat",
			slog.Int64("id", chatID),
		),
		slog.String("repo", repoURL),
		slog.String("event", string(event)),
		slog.Duration("duration", time.Since(start)),
		slog.Any("err", err),
	)
	return err
}

func (r *LoggingRepository) UnmuteEvent(ctx context.Context, chatID int64, repoURL string, event domain.EventType) error {
	start := time.Now()
	err := r.repo.UnmuteEvent(ctx, chatID, repoURL, event)
	r.log.Debug("UnmuteEvent",
		slog.Group("chat", slog.Int64("id", chatID)),
		slog.String("repo", repoURL),
		slog.String("event", string(event)),
		slog.Duration("duration", time.Since(start)),
		slog.Any("err", err),
	)
	return err
}
