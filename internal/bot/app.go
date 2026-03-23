package bot

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/dlc-01/GitNotify/internal/bot/callback"
	"github.com/dlc-01/GitNotify/internal/bot/commands"
	"github.com/dlc-01/GitNotify/internal/bot/core"
	internalkafka "github.com/dlc-01/GitNotify/internal/kafka"
	"github.com/dlc-01/GitNotify/internal/kafka/producer"
	"github.com/dlc-01/GitNotify/internal/repository"
)

type App struct {
	api     *tgbotapi.BotAPI
	handler *Handler
	log     *slog.Logger
}

func New(
	token string,
	repo repository.Repository,
	log *slog.Logger,
	prod producer.MultiProducer,
) (*App, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("init telegram bot: %w", err)
	}
	log.Info("bot authorized",
		slog.String("username", api.Self.UserName),
	)

	sender := core.NewSender(api, log)
	registry := core.NewRegistry()
	callbackRegistry := callback.NewRegistry()

	registry.Register(commands.NewSubscribeCommand(repo, sender, log, prod, internalkafka.TopicSubscriptionCreated))
	registry.Register(commands.NewUnsubscribeCommand(repo, sender, log, prod, internalkafka.TopicSubscriptionDeleted))
	registry.Register(commands.NewListCommand(repo, sender, log))
	registry.Register(commands.NewMuteCommand(repo, sender, log, prod, internalkafka.TopicSubscriptionMuted))
	registry.Register(commands.NewHelpCommand(sender, registry))
	registry.Register(commands.NewStartCommand(sender, registry))

	callbackRegistry.Register(callback.NewUnsubscribeHandler(repo, sender, log))
	callbackRegistry.Register(callback.NewMuteHandler(repo, sender, log))

	handler := NewHandler(api, repo, log, sender, registry, callbackRegistry)

	return &App{
		api:     api,
		handler: handler,
		log:     log,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	if err := a.handler.SetupCommands(); err != nil {
		return fmt.Errorf("setup commands: %w", err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30
	updates := a.api.GetUpdatesChan(u)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	a.log.Info("bot started")

	var wg sync.WaitGroup

	for {
		select {
		case update := <-updates:
			wg.Add(1)
			go func() {
				defer wg.Done()
				a.handler.Dispatch(ctx, update)
			}()
		case sig := <-sigCh:
			a.log.Info("shutting down",
				slog.String("signal", sig.String()),
			)
			a.api.StopReceivingUpdates()
			wg.Wait()
			return nil
		case <-ctx.Done():
			a.api.StopReceivingUpdates()
			wg.Wait()
			return ctx.Err()
		}
	}
}
