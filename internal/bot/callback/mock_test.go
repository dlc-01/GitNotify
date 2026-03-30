package callback

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/mock"

	"github.com/dlc-01/GitNotify/internal/domain"
	internalkafka "github.com/dlc-01/GitNotify/internal/kafka"
)

type mockRepository struct {
	mock.Mock
}

func (m *mockRepository) UpsertUser(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockRepository) UpsertChat(ctx context.Context, chat *domain.Chat) error {
	args := m.Called(ctx, chat)
	return args.Error(0)
}

func (m *mockRepository) Subscribe(ctx context.Context, chatID int64, repoURL string) (*domain.Subscription, error) {
	args := m.Called(ctx, chatID, repoURL)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Subscription), args.Error(1)
}

func (m *mockRepository) Unsubscribe(ctx context.Context, chatID int64, repoURL string) error {
	args := m.Called(ctx, chatID, repoURL)
	return args.Error(0)
}

func (m *mockRepository) ListSubscriptions(ctx context.Context, chatID int64) ([]*domain.Subscription, error) {
	args := m.Called(ctx, chatID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Subscription), args.Error(1)
}

func (m *mockRepository) MuteEvent(ctx context.Context, chatID int64, repoURL string, event domain.EventType) error {
	args := m.Called(ctx, chatID, repoURL, event)
	return args.Error(0)
}

func (m *mockRepository) UnmuteEvent(ctx context.Context, chatID int64, repoURL string, event domain.EventType) error {
	args := m.Called(ctx, chatID, repoURL, event)
	return args.Error(0)
}

type mockMultiProducer struct {
	mock.Mock
}

func (m *mockMultiProducer) ProduceTo(ctx context.Context, topic internalkafka.Topic, msg any) error {
	args := m.Called(ctx, topic, msg)
	return args.Error(0)
}

func (m *mockMultiProducer) Close() error {
	args := m.Called()
	return args.Error(0)
}

type mockSender struct {
	sent     []string
	answered []string
	edited   []string
}

func (m *mockSender) Send(chatID int64, text string)   { m.sent = append(m.sent, text) }
func (m *mockSender) SendErr(chatID int64, err error)  { m.sent = append(m.sent, err.Error()) }
func (m *mockSender) SendMD(chatID int64, text string) { m.sent = append(m.sent, text) }
func (m *mockSender) SendWithKeyboard(chatID int64, text string, keyboard tgbotapi.InlineKeyboardMarkup) {
	m.sent = append(m.sent, text)
}
func (m *mockSender) EditKeyboard(chatID int64, messageID int, keyboard tgbotapi.InlineKeyboardMarkup) {
}
func (m *mockSender) EditText(chatID int64, messageID int, text string) {
	m.edited = append(m.edited, text)
}
func (m *mockSender) AnswerCallback(callbackID string, text string) {
	m.answered = append(m.answered, text)
}
func (m *mockSender) IsAdmin(chatID, userID int64) bool { return true }
