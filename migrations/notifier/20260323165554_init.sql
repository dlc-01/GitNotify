-- +goose Up
CREATE TABLE subscriptions (
                               chat_id      BIGINT NOT NULL,
                               repo_url     TEXT NOT NULL,
                               muted_events TEXT[] NOT NULL DEFAULT '{}',
                               created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                               UNIQUE (chat_id, repo_url)
);

CREATE INDEX idx_subscriptions_repo_url ON subscriptions(repo_url);

-- +goose Down
DROP INDEX idx_subscriptions_repo_url;
DROP TABLE subscriptions;