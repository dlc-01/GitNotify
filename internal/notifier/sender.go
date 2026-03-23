package notifier

import (
	"context"
	"fmt"
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Senderer interface {
	Send(ctx context.Context, chatID int64, text string) error
}

type Sender struct {
	api *tgbotapi.BotAPI
	log *slog.Logger
}

func NewSender(token string, log *slog.Logger) (*Sender, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("init telegram bot: %w", err)
	}
	log.Info("notifier bot authorized",
		slog.String("username", api.Self.UserName),
	)
	return &Sender{api: api, log: log}, nil
}

func (s *Sender) Send(ctx context.Context, chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := s.api.Send(msg); err != nil {
		s.log.Error("send message",
			slog.Group("chat", slog.Int64("id", chatID)),
			slog.String("err", err.Error()),
		)
		return wrap("Send", "", ErrSendMessage)
	}
	return nil
}
