package commands

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/dlc-01/GitNotify/internal/bot/core"
	"github.com/dlc-01/GitNotify/internal/domain"
	internalkafka "github.com/dlc-01/GitNotify/internal/kafka"
	"github.com/dlc-01/GitNotify/internal/kafka/producer"
	"github.com/dlc-01/GitNotify/internal/repository"
)

type MuteCommand struct {
	repo     repository.Repository
	sender   core.Senderer
	log      *slog.Logger
	producer producer.MultiProducer
	topic    internalkafka.Topic
}

func NewMuteCommand(repo repository.Repository, sender core.Senderer, log *slog.Logger, p producer.MultiProducer, topic internalkafka.Topic) *MuteCommand {
	return &MuteCommand{repo: repo, sender: sender, log: log, producer: p, topic: topic}
}

func (c *MuteCommand) Name() string        { return "mute" }
func (c *MuteCommand) Description() string { return "Mute an event type for a resource" }
func (c *MuteCommand) Usage() string       { return "/mute <url> <event>" }

func (c *MuteCommand) Execute(ctx context.Context, chatID int64, args string) {
	parts := strings.Fields(args)
	if len(parts) < 2 {
		c.sender.SendErr(chatID, core.Wrap("Execute", core.ErrEmptyArgs))
		return
	}

	repoURL := normalizeURL(parts[0])
	event := domain.EventType(parts[1])

	if !event.Valid() {
		c.sender.SendErr(chatID, core.Wrap("Execute", core.ErrInvalidEvent))
		return
	}

	if err := c.repo.MuteEvent(ctx, chatID, repoURL, event); err != nil {
		var repoErr *repository.Error
		if errors.As(err, &repoErr) && errors.Is(repoErr, repository.ErrNotFound) {
			c.log.Warn("subscription not found on mute",
				slog.Group("chat", slog.Int64("id", chatID)),
				slog.String("repo", repoURL),
				slog.String("event", string(event)),
			)
			c.sender.Send(chatID, "Subscription not found")
			return
		}
		c.log.Error("mute event",
			slog.Group("chat", slog.Int64("id", chatID)),
			slog.String("repo", repoURL),
			slog.String("event", string(event)),
			slog.String("err", err.Error()),
		)
		c.sender.SendErr(chatID, core.Wrap("Execute", core.ErrInternal))
		return
	}

	if err := c.producer.ProduceTo(ctx, c.topic, internalkafka.SubscriptionMutedMessage{
		ChatID:  chatID,
		RepoURL: repoURL,
		Event:   string(event),
	}); err != nil {
		c.log.Error("produce subscription muted",
			slog.Group("chat", slog.Int64("id", chatID)),
			slog.String("repo", repoURL),
			slog.String("event", string(event)),
			slog.String("err", err.Error()),
		)
	}

	c.log.Info("muted event",
		slog.Group("chat", slog.Int64("id", chatID)),
		slog.String("repo", repoURL),
		slog.String("event", string(event)),
	)
	c.sender.Send(chatID, fmt.Sprintf("🔕 Muted %s events for %s", string(event), repoURL))
}
