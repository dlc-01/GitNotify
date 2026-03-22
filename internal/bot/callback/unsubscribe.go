package callback

import (
	"context"
	"errors"
	"log/slog"

	"github.com/dlc-01/GitNotify/internal/bot/core"
	"github.com/dlc-01/GitNotify/internal/repository"
)

type UnsubscribeHandler struct {
	repo   repository.Repository
	sender *core.Sender
	log    *slog.Logger
}

func NewUnsubscribeHandler(repo repository.Repository, sender *core.Sender, log *slog.Logger) *UnsubscribeHandler {
	return &UnsubscribeHandler{repo: repo, sender: sender, log: log}
}

func (h *UnsubscribeHandler) Action() string { return "unsubscribe" }

func (h *UnsubscribeHandler) Execute(ctx context.Context, chatID int64, messageID int, args string) {
	if err := h.repo.Unsubscribe(ctx, chatID, args); err != nil {
		var repoErr *repository.Error
		if errors.As(err, &repoErr) && errors.Is(repoErr, repository.ErrNotFound) {
			h.sender.AnswerCallback("", "Subscription not found")
			return
		}
		h.log.Error("callback unsubscribe",
			slog.Group("chat",
				slog.Int64("id", chatID),
			),
			slog.String("repo", args),
			slog.String("err", err.Error()),
		)
		h.sender.AnswerCallback("", "Internal error, please try again later")
		return
	}

	h.log.Info("unsubscribed via callback",
		slog.Group("chat",
			slog.Int64("id", chatID),
		),
		slog.String("repo", args),
	)

	h.sender.EditText(chatID, messageID, "✅ Unsubscribed from "+args)
}
