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

type SubscribeCommand struct {
	repo     repository.Repository
	sender   core.Senderer
	log      *slog.Logger
	producer producer.MultiProducer
	topic    internalkafka.Topic
}

func NewSubscribeCommand(repo repository.Repository, sender core.Senderer, log *slog.Logger, p producer.MultiProducer, topic internalkafka.Topic) *SubscribeCommand {
	return &SubscribeCommand{repo: repo, sender: sender, log: log, producer: p, topic: topic}
}

func (c *SubscribeCommand) Name() string        { return "subscribe" }
func (c *SubscribeCommand) Description() string { return "Subscribe to a repository or resource" }
func (c *SubscribeCommand) Usage() string       { return "/subscribe <url>" }

func (c *SubscribeCommand) Execute(ctx context.Context, chatID int64, args string) {
	repoURL := strings.TrimSpace(args)
	if repoURL == "" {
		c.sender.SendErr(chatID, core.Wrap("Execute", core.ErrEmptyArgs))
		return
	}
	if !isValidRepoURL(repoURL) {
		c.sender.SendErr(chatID, core.Wrap("Execute", core.ErrInvalidRepoURL))
		return
	}

	sub, err := c.repo.Subscribe(ctx, chatID, repoURL)
	if err != nil {
		var repoErr *repository.Error
		if errors.As(err, &repoErr) && errors.Is(repoErr, repository.ErrAlreadyExists) {
			c.log.Warn("already subscribed",
				slog.Group("chat", slog.Int64("id", chatID)),
				slog.String("repo", repoURL),
			)
			c.sender.Send(chatID, "Already subscribed to this repository")
			return
		}
		c.log.Error("subscribe",
			slog.Group("chat", slog.Int64("id", chatID)),
			slog.String("repo", repoURL),
			slog.String("err", err.Error()),
		)
		c.sender.SendErr(chatID, core.Wrap("Execute", core.ErrInternal))
		return
	}

	if err := c.producer.ProduceTo(ctx, c.topic, internalkafka.SubscriptionCreatedMessage{
		ChatID:  chatID,
		RepoURL: repoURL,
	}); err != nil {
		c.log.Error("produce subscription created",
			slog.Group("chat", slog.Int64("id", chatID)),
			slog.String("repo", repoURL),
			slog.String("err", err.Error()),
		)
	}

	c.log.Info("subscribed",
		slog.Group("chat", slog.Int64("id", chatID)),
		slog.String("repo", sub.RepoURL),
	)
	c.sender.Send(chatID, "✅ Subscribed to "+sub.RepoURL)
}
