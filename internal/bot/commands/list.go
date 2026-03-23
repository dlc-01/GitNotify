package commands

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/dlc-01/GitNotify/internal/bot/core"
	"github.com/dlc-01/GitNotify/internal/domain"
	"github.com/dlc-01/GitNotify/internal/repository"
)

type ListCommand struct {
	repo   repository.Repository
	sender core.Senderer
	log    *slog.Logger
}

func NewListCommand(repo repository.Repository, sender core.Senderer, log *slog.Logger) *ListCommand {
	return &ListCommand{repo: repo, sender: sender, log: log}
}

func (c *ListCommand) Name() string        { return "list" }
func (c *ListCommand) Description() string { return "List your subscriptions" }
func (c *ListCommand) Usage() string       { return "/list" }

func (c *ListCommand) Execute(ctx context.Context, chatID int64, args string) {
	subs, err := c.repo.ListSubscriptions(ctx, chatID)
	if err != nil {
		c.log.Error("list subscriptions",
			slog.Group("chat",
				slog.Int64("id", chatID)),
			slog.String("err", err.Error()),
		)
		c.sender.SendErr(chatID, core.Wrap("Execute", core.ErrInternal))
		return
	}

	if len(subs) == 0 {
		c.sender.Send(chatID, "You have no subscriptions. Add one with /subscribe")
		return
	}

	for _, sub := range subs {
		c.sender.SendWithKeyboard(chatID, formatSubscription(sub), buildKeyboard(sub))
	}
}

func formatSubscription(sub *domain.Subscription) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📦 %s", sub.RepoURL))
	if len(sub.MutedEvents) > 0 {
		muted := make([]string, len(sub.MutedEvents))
		for i, e := range sub.MutedEvents {
			muted[i] = string(e)
		}
		sb.WriteString(fmt.Sprintf("\n🔕 muted: %s", strings.Join(muted, ", ")))
	}
	return sb.String()
}

func buildKeyboard(sub *domain.Subscription) tgbotapi.InlineKeyboardMarkup {
	muteButtons := make([]tgbotapi.InlineKeyboardButton, 0, len(domain.AllEventTypes))
	for _, event := range domain.AllEventTypes {
		if !sub.IsEventMuted(event) {
			muteButtons = append(muteButtons, tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("🔕 %s", string(event)),
				fmt.Sprintf("mute:%s:%s", sub.RepoURL, string(event)),
			))
		}
	}

	rows := [][]tgbotapi.InlineKeyboardButton{
		{
			tgbotapi.NewInlineKeyboardButtonData(
				"❌ unsubscribe",
				fmt.Sprintf("unsubscribe:%s", sub.RepoURL),
			),
		},
	}

	if len(muteButtons) > 0 {
		rows = append(rows, muteButtons)
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}
