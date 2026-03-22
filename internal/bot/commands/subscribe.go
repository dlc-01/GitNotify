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
	sender   *core.Sender
	log      *slog.Logger
	producer producer.Producer
}

func NewSubscribeCommand(repo repository.Repository, sender *core.Sender, log *slog.Logger, p producer.Producer) *SubscribeCommand {
	return &SubscribeCommand{repo: repo, sender: sender, log: log, producer: p}
}

func (c *SubscribeCommand) Name() string        { return "subscribe" }
func (c *SubscribeCommand) Description() string { return "Subscribe to a repository" }
func (c *SubscribeCommand) Usage() string       { return "/subscribe <repo_url>" }

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

	if err := c.producer.Produce(ctx, internalkafka.SubscriptionCreatedMessage{
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
