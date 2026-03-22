package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log/slog"
)

type Sender struct {
	api *tgbotapi.BotAPI
	log *slog.Logger
}

func NewSender(api *tgbotapi.BotAPI, log *slog.Logger) *Sender {
	return &Sender{api: api, log: log}
}

func (s *Sender) Send(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := s.api.Send(msg); err != nil {
		s.log.Error("send message", "err", err)
	}
}

func (s *Sender) SendErr(chatID int64, err error) {
	s.Send(chatID, formatError(err))
}
