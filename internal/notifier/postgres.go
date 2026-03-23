package notifier

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dlc-01/GitNotify/internal/domain"
)

type postgresRepository struct {
	db  *pgxpool.Pool
	log *slog.Logger
}

func NewPostgresRepository(db *pgxpool.Pool, log *slog.Logger) Repository {
	return &postgresRepository{db: db, log: log}
}

func (r *postgresRepository) Subscribe(ctx context.Context, chatID int64, repoURL string) error {
	start := time.Now()
	_, err := r.db.Exec(ctx, `
		INSERT INTO subscriptions (chat_id, repo_url)
		VALUES ($1, $2)
		ON CONFLICT (chat_id, repo_url) DO NOTHING
	`, chatID, repoURL)
	r.log.Debug("Subscribe",
		slog.Group("chat", slog.Int64("id", chatID)),
		slog.String("repo", repoURL),
		slog.Duration("duration", time.Since(start)),
		slog.Any("err", err),
	)
	return err
}

func (r *postgresRepository) Unsubscribe(ctx context.Context, chatID int64, repoURL string) error {
	start := time.Now()
	tag, err := r.db.Exec(ctx, `
		DELETE FROM subscriptions WHERE chat_id = $1 AND repo_url = $2
	`, chatID, repoURL)
	r.log.Debug("Unsubscribe",
		slog.Group("chat", slog.Int64("id", chatID)),
		slog.String("repo", repoURL),
		slog.Duration("duration", time.Since(start)),
		slog.Any("err", err),
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *postgresRepository) MuteEvent(ctx context.Context, chatID int64, repoURL string, event domain.EventType) error {
	start := time.Now()
	_, err := r.db.Exec(ctx, `
		UPDATE subscriptions
		SET muted_events = array_append(
			array_remove(muted_events, $3::text),
			$3::text
		)
		WHERE chat_id = $1 AND repo_url = $2
	`, chatID, repoURL, string(event))
	r.log.Debug("MuteEvent",
		slog.Group("chat", slog.Int64("id", chatID)),
		slog.String("repo", repoURL),
		slog.String("event", string(event)),
		slog.Duration("duration", time.Since(start)),
		slog.Any("err", err),
	)
	return err
}

func (r *postgresRepository) GetSubscribersByRepo(ctx context.Context, repoURL string, event domain.EventType) ([]int64, error) {
	start := time.Now()
	rows, err := r.db.Query(ctx, `
		SELECT chat_id FROM subscriptions
		WHERE repo_url = $1
		  AND NOT ($2::text = ANY(muted_events))
	`, repoURL, string(event))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chatIDs []int64
	for rows.Next() {
		var chatID int64
		if err := rows.Scan(&chatID); err != nil {
			return nil, err
		}
		chatIDs = append(chatIDs, chatID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	r.log.Debug("GetSubscribersByRepo",
		slog.String("repo", repoURL),
		slog.String("event", string(event)),
		slog.Int("count", len(chatIDs)),
		slog.Duration("duration", time.Since(start)),
		slog.Any("err", err),
	)

	return chatIDs, nil
}

var ErrNotFound = errors.New("not found")

func isNoRows(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}
