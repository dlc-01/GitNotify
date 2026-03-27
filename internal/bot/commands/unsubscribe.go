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
	sender   core.Senderer
	log      *slog.Logger
	producer producer.MultiProducer
	topic    internalkafka.Topic
}

func NewUnsubscribeCommand(repo repository.Repository, sender core.Senderer, log *slog.Logger, p producer.MultiProducer, topic internalkafka.Topic) *UnsubscribeCommand {
	return &UnsubscribeCommand{repo: repo, sender: sender, log: log, producer: p, topic: topic}
}

func (c *UnsubscribeCommand) Name() string        { return "unsubscribe" }
func (c *UnsubscribeCommand) Description() string { return "Unsubscribe from a repository or resource" }
func (c *UnsubscribeCommand) Usage() string       { return "/unsubscribe <url>" }
func (c *UnsubscribeCommand) Execute(ctx context.Context, chatID int64, args string) {
	repoURL := normalizeURL(strings.TrimSpace(args))
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

	if err := c.producer.ProduceTo(ctx, c.topic, internalkafka.SubscriptionDeletedMessage{
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
