-- +goose Up
CREATE TABLE users (
                       user_id    BIGINT PRIMARY KEY,
                       username   TEXT NOT NULL DEFAULT '',
                       created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE chats (
                       chat_id    BIGINT PRIMARY KEY,
                       chat_type  TEXT NOT NULL,
                       created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE subscriptions (
                               id           BIGSERIAL PRIMARY KEY,
                               chat_id      BIGINT NOT NULL REFERENCES chats(chat_id) ON DELETE CASCADE,
                               repo_url     TEXT NOT NULL,
                               muted_events TEXT[] NOT NULL DEFAULT '{}',
                               created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                               UNIQUE (chat_id, repo_url)
);

CREATE INDEX idx_subscriptions_chat_id ON subscriptions(chat_id);
CREATE INDEX idx_subscriptions_repo_url ON subscriptions(repo_url);

-- +goose Down
DROP INDEX idx_subscriptions_repo_url;
DROP INDEX idx_subscriptions_chat_id;
DROP TABLE subscriptions;
DROP TABLE chats;
DROP TABLE users;
