package core

import (
	"context"
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type HandlerFunc func(ctx context.Context, update tgbotapi.Update)

type Middleware func(next HandlerFunc) HandlerFunc

func Chain(h HandlerFunc, middlewares ...Middleware) HandlerFunc {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

func LogRequest(log *slog.Logger) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, update tgbotapi.Update) {
			if update.Message != nil {
				log.Debug("incoming message",
					slog.Group("chat",
						slog.Int64("id", update.Message.Chat.ID),
						slog.String("type", update.Message.Chat.Type),
					),
					slog.Group("user",
						slog.Int64("id", update.Message.From.ID),
						slog.String("username", update.Message.From.UserName),
					),
					slog.String("command", update.Message.Command()),
				)
			}

			if update.CallbackQuery != nil {
				log.Debug("incoming callback",
					slog.Group("user",
						slog.Int64("id", update.CallbackQuery.From.ID),
						slog.String("username", update.CallbackQuery.From.UserName),
					),
					slog.String("data", update.CallbackQuery.Data),
				)
			}

			next(ctx, update)
		}
	}
}

func AdminOnly(sender Senderer, log *slog.Logger) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, update tgbotapi.Update) {
			if update.Message == nil {
				next(ctx, update)
				return
			}

			chat := update.Message.Chat
			if chat.IsPrivate() {
				next(ctx, update)
				return
			}

			if !sender.IsAdmin(chat.ID, update.Message.From.ID) {
				log.Warn("non-admin tried to use command",
					slog.Group("chat",
						slog.Int64("id", chat.ID),
						slog.String("type", chat.Type),
					),
					slog.Group("user",
						slog.Int64("id", update.Message.From.ID),
						slog.String("username", update.Message.From.UserName),
					),
					slog.String("command", update.Message.Command()),
				)
				sender.Send(chat.ID, FormatError(Wrap("AdminOnly", ErrNotAdmin)))
				return
			}

			next(ctx, update)
		}
	}
}
