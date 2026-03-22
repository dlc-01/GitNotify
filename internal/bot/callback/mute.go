package callback

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/dlc-01/GitNotify/internal/bot/core"
	"github.com/dlc-01/GitNotify/internal/domain"
	"github.com/dlc-01/GitNotify/internal/repository"
)

type MuteHandler struct {
	repo   repository.Repository
	sender *core.Sender
	log    *slog.Logger
}

func NewMuteHandler(repo repository.Repository, sender *core.Sender, log *slog.Logger) *MuteHandler {
	return &MuteHandler{repo: repo, sender: sender, log: log}
}

func (h *MuteHandler) Action() string { return "mute" }

func (h *MuteHandler) Execute(ctx context.Context, chatID int64, messageID int, args string) {
	parts := strings.SplitN(args, ":", 2)
	if len(parts) < 2 {
		h.sender.AnswerCallback("", "Invalid callback data")
		return
	}

	repoURL := parts[0]
	event := domain.EventType(parts[1])

	if !event.Valid() {
		h.sender.AnswerCallback("", "Invalid event type")
		return
	}

	if err := h.repo.MuteEvent(ctx, chatID, repoURL, event); err != nil {
		var repoErr *repository.Error
		if errors.As(err, &repoErr) && errors.Is(repoErr, repository.ErrNotFound) {
			h.sender.AnswerCallback("", "Subscription not found")
			return
		}
		h.log.Error("callback mute",
			slog.Group("chat",
				slog.Int64("id", chatID),
			),
			slog.String("repo", repoURL),
			slog.String("event", string(event)),
			slog.String("err", err.Error()),
		)
		h.sender.AnswerCallback("", "Internal error, please try again later")
		return
	}

	h.log.Info("muted via callback",
		slog.Group("chat",
			slog.Int64("id", chatID),
		),
		slog.String("repo", repoURL),
		slog.String("event", string(event)),
	)

	h.sender.EditText(chatID, messageID, "🔕 Muted "+string(event)+" events for "+repoURL)
}
