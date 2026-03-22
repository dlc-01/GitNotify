package commands

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/dlc-01/GitNotify/internal/bot/core"
	internalkafka "github.com/dlc-01/GitNotify/internal/kafka"
	"github.com/dlc-01/GitNotify/internal/kafka/producer"
	"github.com/dlc-01/GitNotify/internal/repository"
)

type UnsubscribeCommand struct {
	repo     repository.Repository
	sender   *core.Sender
	log      *slog.Logger
	producer producer.Producer
}

func NewUnsubscribeCommand(repo repository.Repository, sender *core.Sender, log *slog.Logger, p producer.Producer) *UnsubscribeCommand {
	return &UnsubscribeCommand{repo: repo, sender: sender, log: log, producer: p}
}

func (c *UnsubscribeCommand) Name() string        { return "unsubscribe" }
func (c *UnsubscribeCommand) Description() string { return "Unsubscribe from a repository" }
func (c *UnsubscribeCommand) Usage() string       { return "/unsubscribe <repo_url>" }

func (c *UnsubscribeCommand) Execute(ctx context.Context, chatID int64, args string) {
	repoURL := strings.TrimSpace(args)
	if repoURL == "" {
		c.sender.SendErr(chatID, core.Wrap("Execute", core.ErrEmptyArgs))
		return
	}

	if err := c.repo.Unsubscribe(ctx, chatID, repoURL); err != nil {
		var repoErr *repository.Error
		if errors.As(err, &repoErr) && errors.Is(repoErr, repository.ErrNotFound) {
			c.log.Warn("subscription not found",
				slog.Group("chat", slog.Int64("id", chatID)),
				slog.String("repo", repoURL),
			)
			c.sender.Send(chatID, "Subscription not found")
			return
		}
		c.log.Error("unsubscribe",
			slog.Group("chat", slog.Int64("id", chatID)),
			slog.String("repo", repoURL),
			slog.String("err", err.Error()),
		)
		c.sender.SendErr(chatID, core.Wrap("Execute", core.ErrInternal))
		return
	}

	if err := c.producer.Produce(ctx, internalkafka.SubscriptionDeletedMessage{
		ChatID:  chatID,
		RepoURL: repoURL,
	}); err != nil {
		c.log.Error("produce subscription deleted",
			slog.Group("chat", slog.Int64("id", chatID)),
			slog.String("repo", repoURL),
			slog.String("err", err.Error()),
		)
	}

	c.log.Info("unsubscribed",
		slog.Group("chat", slog.Int64("id", chatID)),
		slog.String("repo", repoURL),
	)

	c.sender.Send(chatID, "✅ Unsubscribed from "+repoURL)
}
