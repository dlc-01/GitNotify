package core

import (
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Senderer interface {
	Send(chatID int64, text string)
	SendMD(chatID int64, text string)
	SendWithKeyboard(chatID int64, text string, keyboard tgbotapi.InlineKeyboardMarkup)
	EditKeyboard(chatID int64, messageID int, keyboard tgbotapi.InlineKeyboardMarkup)
	EditText(chatID int64, messageID int, text string)
	AnswerCallback(callbackID string, text string)
	SendErr(chatID int64, err error)
	IsAdmin(chatID, userID int64) bool
}

type sender struct {
	api *tgbotapi.BotAPI
	log *slog.Logger
}

func NewSender(api *tgbotapi.BotAPI, log *slog.Logger) Senderer {
	return &sender{api: api, log: log}
}

func (s *sender) Send(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := s.api.Send(msg); err != nil {
		s.log.Error("send message",
			slog.Group("chat", slog.Int64("id", chatID)),
			slog.String("err", err.Error()),
		)
	}
}

func (s *sender) SendMD(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdownV2
	if _, err := s.api.Send(msg); err != nil {
		s.log.Error("send markdown",
			slog.Group("chat", slog.Int64("id", chatID)),
			slog.String("err", err.Error()),
		)
	}
}

func (s *sender) SendWithKeyboard(chatID int64, text string, keyboard tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	if _, err := s.api.Send(msg); err != nil {
		s.log.Error("send with keyboard",
			slog.Group("chat", slog.Int64("id", chatID)),
			slog.String("err", err.Error()),
		)
	}
}

func (s *sender) EditKeyboard(chatID int64, messageID int, keyboard tgbotapi.InlineKeyboardMarkup) {
	edit := tgbotapi.NewEditMessageReplyMarkup(chatID, messageID, keyboard)
	if _, err := s.api.Send(edit); err != nil {
		s.log.Error("edit keyboard",
			slog.Group("chat", slog.Int64("id", chatID)),
			slog.Int("message_id", messageID),
			slog.String("err", err.Error()),
		)
	}
}

func (s *sender) EditText(chatID int64, messageID int, text string) {
	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	if _, err := s.api.Send(edit); err != nil {
		s.log.Error("edit text",
			slog.Group("chat", slog.Int64("id", chatID)),
			slog.Int("message_id", messageID),
			slog.String("err", err.Error()),
		)
	}
}

func (s *sender) AnswerCallback(callbackID string, text string) {
	callback := tgbotapi.NewCallback(callbackID, text)
	if _, err := s.api.Request(callback); err != nil {
		s.log.Error("answer callback",
			slog.String("callback_id", callbackID),
			slog.String("err", err.Error()),
		)
	}
}

func (s *sender) SendErr(chatID int64, err error) {
	s.Send(chatID, FormatError(err))
}

func (s *sender) IsAdmin(chatID, userID int64) bool {
	member, err := s.api.GetChatMember(tgbotapi.GetChatMemberConfig{
		ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
			ChatID: chatID,
			UserID: userID,
		},
	})
	if err != nil {
		s.log.Error("get chat member",
			slog.Group("chat", slog.Int64("id", chatID)),
			slog.Group("user", slog.Int64("id", userID)),
			slog.String("err", err.Error()),
		)
		return false
	}
	return member.IsAdministrator() || member.IsCreator()
}
