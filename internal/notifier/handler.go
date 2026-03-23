package notifier

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/dlc-01/GitNotify/internal/domain"
	internalkafka "github.com/dlc-01/GitNotify/internal/kafka"
)

type Handler struct {
	repo   Repository
	sender Senderer
	log    *slog.Logger
}

func NewHandler(repo Repository, sender Senderer, log *slog.Logger) *Handler {
	return &Handler{repo: repo, sender: sender, log: log}
}

func (h *Handler) HandleSubscriptionCreated(ctx context.Context, data []byte) error {
	var msg internalkafka.SubscriptionCreatedMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return wrap("HandleSubscriptionCreated", internalkafka.TopicSubscriptionCreated.String(), ErrUnmarshal)
	}

	if err := h.repo.Subscribe(ctx, msg.ChatID, msg.RepoURL); err != nil {
		h.log.Error("subscribe",
			slog.Group("chat", slog.Int64("id", msg.ChatID)),
			slog.String("repo", msg.RepoURL),
			slog.String("err", err.Error()),
		)
		return wrap("HandleSubscriptionCreated", internalkafka.TopicSubscriptionCreated.String(), err)
	}

	h.log.Info("subscription created",
		slog.Group("chat", slog.Int64("id", msg.ChatID)),
		slog.String("repo", msg.RepoURL),
	)
	return nil
}

func (h *Handler) HandleSubscriptionDeleted(ctx context.Context, data []byte) error {
	var msg internalkafka.SubscriptionDeletedMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return wrap("HandleSubscriptionDeleted", internalkafka.TopicSubscriptionDeleted.String(), ErrUnmarshal)
	}

	if err := h.repo.Unsubscribe(ctx, msg.ChatID, msg.RepoURL); err != nil {
		if err == ErrNotFound {
			h.log.Warn("subscription not found on delete",
				slog.Group("chat", slog.Int64("id", msg.ChatID)),
				slog.String("repo", msg.RepoURL),
			)
			return nil
		}
		h.log.Error("unsubscribe",
			slog.Group("chat", slog.Int64("id", msg.ChatID)),
			slog.String("repo", msg.RepoURL),
			slog.String("err", err.Error()),
		)
		return wrap("HandleSubscriptionDeleted", internalkafka.TopicSubscriptionDeleted.String(), err)
	}

	h.log.Info("subscription deleted",
		slog.Group("chat", slog.Int64("id", msg.ChatID)),
		slog.String("repo", msg.RepoURL),
	)
	return nil
}

func (h *Handler) HandleSubscriptionMuted(ctx context.Context, data []byte) error {
	var msg internalkafka.SubscriptionMutedMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return wrap("HandleSubscriptionMuted", internalkafka.TopicSubscriptionMuted.String(), ErrUnmarshal)
	}

	event := domain.EventType(msg.Event)
	if err := h.repo.MuteEvent(ctx, msg.ChatID, msg.RepoURL, event); err != nil {
		h.log.Error("mute event",
			slog.Group("chat", slog.Int64("id", msg.ChatID)),
			slog.String("repo", msg.RepoURL),
			slog.String("event", msg.Event),
			slog.String("err", err.Error()),
		)
		return wrap("HandleSubscriptionMuted", internalkafka.TopicSubscriptionMuted.String(), err)
	}

	h.log.Info("subscription muted",
		slog.Group("chat", slog.Int64("id", msg.ChatID)),
		slog.String("repo", msg.RepoURL),
		slog.String("event", msg.Event),
	)
	return nil
}

func (h *Handler) HandleEvent(ctx context.Context, data []byte) error {
	var msg internalkafka.WebhookEventMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return wrap("HandleEvent", "", ErrUnmarshal)
	}

	event := domain.EventType(msg.EventType)
	chatIDs, err := h.repo.GetSubscribersByRepo(ctx, msg.RepoURL, event)
	if err != nil {
		h.log.Error("get subscribers",
			slog.String("repo", msg.RepoURL),
			slog.String("event", msg.EventType),
			slog.String("err", err.Error()),
		)
		return wrap("HandleEvent", "", err)
	}

	if len(chatIDs) == 0 {
		h.log.Debug("no subscribers",
			slog.String("repo", msg.RepoURL),
			slog.String("event", msg.EventType),
		)
		return nil
	}

	text := formatEventMessage(msg)
	for _, chatID := range chatIDs {
		if err := h.sender.Send(ctx, chatID, text); err != nil {
			h.log.Error("send notification",
				slog.Group("chat", slog.Int64("id", chatID)),
				slog.String("repo", msg.RepoURL),
				slog.String("err", err.Error()),
			)
		}
	}

	h.log.Info("event notifications sent",
		slog.String("repo", msg.RepoURL),
		slog.String("event", msg.EventType),
		slog.String("source", msg.Source),
		slog.Int("recipients", len(chatIDs)),
	)
	return nil
}

func formatEventMessage(msg internalkafka.WebhookEventMessage) string {
	switch domain.EventType(msg.EventType) {
	case domain.EventPush:
		return "🔔 New push to " + msg.RepoURL
	case domain.EventPR:
		return "🔔 New pull request in " + msg.RepoURL
	case domain.EventIssue:
		return "🔔 New issue in " + msg.RepoURL
	case domain.EventPipeline:
		return "🔔 Pipeline triggered in " + msg.RepoURL
	case domain.EventAnswer:
		return "🔔 New answer on " + msg.RepoURL
	case domain.EventPost:
		return "🔔 New post on " + msg.RepoURL
	case domain.EventVideo:
		return "🔔 New video on " + msg.RepoURL
	default:
		return "🔔 New event on " + msg.RepoURL
	}
}
