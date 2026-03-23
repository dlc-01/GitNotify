package notifier

import (
	"context"
	"log/slog"

	internalkafka "github.com/dlc-01/GitNotify/internal/kafka"
	"github.com/dlc-01/GitNotify/internal/kafka/consumer"
)

type App struct {
	consumer *consumer.Consumer
	handler  *Handler
	log      *slog.Logger
}

func New(
	brokers []string,
	groupID string,
	handler *Handler,
	log *slog.Logger,
) *App {
	c := consumer.New(brokers, groupID, log)
	return &App{
		consumer: c,
		handler:  handler,
		log:      log,
	}
}

func (a *App) Run(ctx context.Context) error {
	a.log.Info("notifier started")

	a.consumer.Subscribe(internalkafka.TopicSubscriptionCreated, a.handler.HandleSubscriptionCreated)
	a.consumer.Subscribe(internalkafka.TopicSubscriptionDeleted, a.handler.HandleSubscriptionDeleted)
	a.consumer.Subscribe(internalkafka.TopicSubscriptionMuted, a.handler.HandleSubscriptionMuted)

	a.consumer.Subscribe(internalkafka.TopicEventPush, a.handler.HandleEvent)
	a.consumer.Subscribe(internalkafka.TopicEventPR, a.handler.HandleEvent)
	a.consumer.Subscribe(internalkafka.TopicEventIssue, a.handler.HandleEvent)
	a.consumer.Subscribe(internalkafka.TopicEventPipeline, a.handler.HandleEvent)
	a.consumer.Subscribe(internalkafka.TopicEventAnswer, a.handler.HandleEvent)
	a.consumer.Subscribe(internalkafka.TopicEventPost, a.handler.HandleEvent)
	a.consumer.Subscribe(internalkafka.TopicEventVideo, a.handler.HandleEvent)

	<-ctx.Done()

	a.log.Info("shutting down notifier")

	if err := a.consumer.Close(); err != nil {
		return err
	}

	a.log.Info("notifier stopped")
	return nil
}
