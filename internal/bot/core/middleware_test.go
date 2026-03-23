package core

import (
	"context"
	"log/slog"
	"os"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/assert"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
}

type mockSender struct {
	sent    []string
	isAdmin bool
}

func (m *mockSender) Send(chatID int64, text string)   { m.sent = append(m.sent, text) }
func (m *mockSender) SendErr(chatID int64, err error)  { m.sent = append(m.sent, FormatError(err)) }
func (m *mockSender) SendMD(chatID int64, text string) {}
func (m *mockSender) SendWithKeyboard(chatID int64, text string, keyboard tgbotapi.InlineKeyboardMarkup) {
}
func (m *mockSender) EditKeyboard(chatID int64, messageID int, keyboard tgbotapi.InlineKeyboardMarkup) {
}
func (m *mockSender) EditText(chatID int64, messageID int, text string) {}
func (m *mockSender) AnswerCallback(callbackID string, text string)     {}
func (m *mockSender) IsAdmin(chatID, userID int64) bool                 { return m.isAdmin }

func TestChain_Order(t *testing.T) {
	order := []string{}

	first := func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, update tgbotapi.Update) {
			order = append(order, "first")
			next(ctx, update)
		}
	}

	second := func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, update tgbotapi.Update) {
			order = append(order, "second")
			next(ctx, update)
		}
	}

	handler := func(ctx context.Context, update tgbotapi.Update) {
		order = append(order, "handler")
	}

	chained := Chain(handler, first, second)
	chained(context.Background(), tgbotapi.Update{})

	assert.Equal(t, []string{"first", "second", "handler"}, order)
}

func TestLogRequest_Message(t *testing.T) {
	log := newTestLogger()
	called := false

	handler := func(ctx context.Context, update tgbotapi.Update) {
		called = true
	}

	middleware := LogRequest(log)
	chained := middleware(handler)

	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			Text: "/subscribe https://github.com/golang/go",
			From: &tgbotapi.User{
				ID:       123,
				UserName: "testuser",
			},
			Chat: &tgbotapi.Chat{
				ID:   100,
				Type: "private",
			},
		},
	}

	chained(context.Background(), update)
	assert.True(t, called)
}

func TestLogRequest_Callback(t *testing.T) {
	log := newTestLogger()
	called := false

	handler := func(ctx context.Context, update tgbotapi.Update) {
		called = true
	}

	middleware := LogRequest(log)
	chained := middleware(handler)

	update := tgbotapi.Update{
		CallbackQuery: &tgbotapi.CallbackQuery{
			ID:   "callback123",
			Data: "unsubscribe:https://github.com/golang/go",
			From: &tgbotapi.User{
				ID:       123,
				UserName: "testuser",
			},
		},
	}

	chained(context.Background(), update)
	assert.True(t, called)
}

func TestAdminOnly_PrivateChat_Passes(t *testing.T) {
	log := newTestLogger()
	sender := &mockSender{isAdmin: false}
	called := false

	handler := func(ctx context.Context, update tgbotapi.Update) {
		called = true
	}

	middleware := AdminOnly(sender, log)
	chained := middleware(handler)

	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			From: &tgbotapi.User{ID: 123, UserName: "testuser"},
			Chat: &tgbotapi.Chat{ID: 123, Type: "private"},
		},
	}

	chained(context.Background(), update)
	assert.True(t, called)
	assert.Empty(t, sender.sent)
}

func TestAdminOnly_Group_Admin_Passes(t *testing.T) {
	log := newTestLogger()
	sender := &mockSender{isAdmin: true}
	called := false

	handler := func(ctx context.Context, update tgbotapi.Update) {
		called = true
	}

	middleware := AdminOnly(sender, log)
	chained := middleware(handler)

	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			From: &tgbotapi.User{ID: 123, UserName: "testuser"},
			Chat: &tgbotapi.Chat{ID: 456, Type: "group"},
		},
	}

	chained(context.Background(), update)
	assert.True(t, called)
	assert.Empty(t, sender.sent)
}

func TestAdminOnly_Group_NonAdmin_Blocked(t *testing.T) {
	log := newTestLogger()
	sender := &mockSender{isAdmin: false}
	called := false

	handler := func(ctx context.Context, update tgbotapi.Update) {
		called = true
	}

	middleware := AdminOnly(sender, log)
	chained := middleware(handler)

	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			From: &tgbotapi.User{ID: 123, UserName: "testuser"},
			Chat: &tgbotapi.Chat{ID: 456, Type: "group"},
		},
	}

	chained(context.Background(), update)
	assert.False(t, called)
	assert.NotEmpty(t, sender.sent)
	assert.Contains(t, sender.sent[0], "Only admins can use this command in groups")
}

func TestAdminOnly_NoMessage_Passes(t *testing.T) {
	log := newTestLogger()
	sender := &mockSender{isAdmin: false}
	called := false

	handler := func(ctx context.Context, update tgbotapi.Update) {
		called = true
	}

	middleware := AdminOnly(sender, log)
	chained := middleware(handler)

	chained(context.Background(), tgbotapi.Update{})
	assert.True(t, called)
}
