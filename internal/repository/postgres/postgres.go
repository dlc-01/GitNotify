package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dlc-01/GitNotify/internal/domain"
	"github.com/dlc-01/GitNotify/internal/repository"
)

type Repository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) UpsertUser(ctx context.Context, user *domain.User) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO users (user_id, username)
		VALUES ($1, $2)
		ON CONFLICT (user_id) DO UPDATE SET username = EXCLUDED.username
	`, user.UserID, user.Username)
	if err != nil {
		return repository.Wrap("UpsertUser", err)
	}
	return nil
}

func (r *Repository) UpsertChat(ctx context.Context, chat *domain.Chat) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO chats (chat_id, chat_type)
		VALUES ($1, $2)
		ON CONFLICT (chat_id) DO UPDATE SET chat_type = EXCLUDED.chat_type
	`, chat.ChatID, chat.ChatType)
	if err != nil {
		return repository.Wrap("UpsertChat", err)
	}
	return nil
}

func (r *Repository) Subscribe(ctx context.Context, chatID int64, repoURL string) (*domain.Subscription, error) {
	sub := &domain.Subscription{}
	err := r.db.QueryRow(ctx, `
		INSERT INTO subscriptions (chat_id, repo_url)
		VALUES ($1, $2)
		ON CONFLICT (chat_id, repo_url) DO NOTHING
		RETURNING id, chat_id, repo_url, muted_events, created_at
	`, chatID, repoURL).Scan(
		&sub.ID, &sub.ChatID, &sub.RepoURL, &sub.MutedEvents, &sub.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.Wrap("Subscribe", repository.ErrAlreadyExists)
		}
		return nil, repository.Wrap("Subscribe", err)
	}
	return sub, nil
}

func (r *Repository) Unsubscribe(ctx context.Context, chatID int64, repoURL string) error {
	tag, err := r.db.Exec(ctx, `
		DELETE FROM subscriptions WHERE chat_id = $1 AND repo_url = $2
	`, chatID, repoURL)
	if err != nil {
		return repository.Wrap("Unsubscribe", err)
	}
	if tag.RowsAffected() == 0 {
		return repository.Wrap("Unsubscribe", repository.ErrNotFound)
	}
	return nil
}

func (r *Repository) ListSubscriptions(ctx context.Context, chatID int64) ([]*domain.Subscription, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, chat_id, repo_url, muted_events, created_at
		FROM subscriptions
		WHERE chat_id = $1
		ORDER BY created_at DESC
	`, chatID)
	if err != nil {
		return nil, repository.Wrap("ListSubscriptions", err)
	}
	defer rows.Close()

	var subs []*domain.Subscription
	for rows.Next() {
		s := &domain.Subscription{}
		if err := rows.Scan(&s.ID, &s.ChatID, &s.RepoURL, &s.MutedEvents, &s.CreatedAt); err != nil {
			return nil, repository.Wrap("ListSubscriptions", err)
		}
		subs = append(subs, s)
	}
	if err := rows.Err(); err != nil {
		return nil, repository.Wrap("ListSubscriptions", err)
	}
	return subs, nil
}

func (r *Repository) MuteEvent(ctx context.Context, chatID int64, repoURL string, event domain.EventType) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE subscriptions
		SET muted_events = array_append(
			array_remove(muted_events, $3::text),
			$3::text
		)
		WHERE chat_id = $1 AND repo_url = $2
	`, chatID, repoURL, string(event))
	if err != nil {
		return repository.Wrap("MuteEvent", err)
	}
	if tag.RowsAffected() == 0 {
		return repository.Wrap("MuteEvent", repository.ErrNotFound)
	}
	return nil
}

func (r *Repository) UnmuteEvent(ctx context.Context, chatID int64, repoURL string, event domain.EventType) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE subscriptions
		SET muted_events = array_remove(muted_events, $3::text)
		WHERE chat_id = $1 AND repo_url = $2
	`, chatID, repoURL, string(event))
	if err != nil {
		return repository.Wrap("UnmuteEvent", err)
	}
	if tag.RowsAffected() == 0 {
		return repository.Wrap("UnmuteEvent", repository.ErrNotFound)
	}
	return nil
}
