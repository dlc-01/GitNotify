package bot

import (
	"context"
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/dlc-01/GitNotify/internal/bot/callback"
	"github.com/dlc-01/GitNotify/internal/bot/core"
	"github.com/dlc-01/GitNotify/internal/domain"
	"github.com/dlc-01/GitNotify/internal/repository"
)

type Handler struct {
	api              *tgbotapi.BotAPI
	repo             repository.Repository
	log              *slog.Logger
	sender           *core.Sender
	registry         *core.Registry
	callbackRegistry *callback.Registry
	dispatch         core.HandlerFunc
}

func NewHandler(
	api *tgbotapi.BotAPI,
	repo repository.Repository,
	log *slog.Logger,
	sender *core.Sender,
	registry *core.Registry,
	callbackRegistry *callback.Registry,
) *Handler {
	h := &Handler{
		api:              api,
		repo:             repo,
		log:              log,
		sender:           sender,
		registry:         registry,
		callbackRegistry: callbackRegistry,
	}

	h.dispatch = core.Chain(
		h.handle,
		core.LogRequest(log),
		core.AdminOnly(sender, log),
	)

	return h
}

func (h *Handler) Dispatch(ctx context.Context, update tgbotapi.Update) {
	h.dispatch(ctx, update)
}

func (h *Handler) handle(ctx context.Context, update tgbotapi.Update) {
	switch {
	case update.Message != nil:
		h.handleMessage(ctx, update)
	case update.CallbackQuery != nil:
		h.handleCallback(ctx, update)
	}
}

func (h *Handler) handleMessage(ctx context.Context, update tgbotapi.Update) {
	if err := h.upsertUserAndChat(ctx, update.Message); err != nil {
		h.log.Error("upsert user and chat",
			slog.Group("chat",
				slog.Int64("id", update.Message.Chat.ID),
			),
			slog.String("err", err.Error()),
		)
		return
	}

	if !update.Message.IsCommand() {
		return
	}

	chatID := update.Message.Chat.ID
	cmdName := update.Message.Command()
	args := update.Message.CommandArguments()

	cmd, ok := h.registry.Get(cmdName)
	if !ok {
		h.sender.Send(chatID, "Unknown command. Use /help to see available commands")
		return
	}

	cmd.Execute(ctx, chatID, args)
}

func (h *Handler) handleCallback(ctx context.Context, update tgbotapi.Update) {
	query := update.CallbackQuery
	chatID := query.Message.Chat.ID
	messageID := query.Message.MessageID

	handler, args, ok := h.callbackRegistry.Get(query.Data)
	if !ok {
		h.log.Warn("unknown callback",
			slog.String("data", query.Data),
			slog.Group("user",
				slog.Int64("id", query.From.ID),
			),
		)
		h.sender.AnswerCallback(query.ID, "Unknown action")
		return
	}

	h.sender.AnswerCallback(query.ID, "")
	handler.Execute(ctx, chatID, messageID, args)
}

func (h *Handler) upsertUserAndChat(ctx context.Context, msg *tgbotapi.Message) error {
	user := &domain.User{
		UserID:   msg.From.ID,
		Username: msg.From.UserName,
	}
	if err := h.repo.UpsertUser(ctx, user); err != nil {
		return core.Wrap("upsertUserAndChat", err)
	}

	chat := &domain.Chat{
		ChatID:   msg.Chat.ID,
		ChatType: domain.ChatType(msg.Chat.Type),
	}
	if err := h.repo.UpsertChat(ctx, chat); err != nil {
		return core.Wrap("upsertUserAndChat", err)
	}
	return nil
}

func (h *Handler) SetupCommands() error {
	cfg := tgbotapi.NewSetMyCommands(h.registry.BotCommands()...)
	_, err := h.api.Request(cfg)
	return err
}
