package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dlc-01/GitNotify/internal/config"
	internalkafka "github.com/dlc-01/GitNotify/internal/kafka"
	"github.com/dlc-01/GitNotify/internal/kafka/consumer"
	"github.com/dlc-01/GitNotify/internal/kafka/producer"
	"github.com/dlc-01/GitNotify/internal/poller"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	if err := run(log); err != nil {
		log.Error("fatal error", slog.String("err", err.Error()))
		os.Exit(1)
	}
}

func run(log *slog.Logger) error {
	configPath := flag.String("config", "", "path to config file (optional)")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		var cfgErr *config.Error
		if errors.As(err, &cfgErr) {
			switch {
			case errors.Is(cfgErr, config.ErrNotFound):
				return fmt.Errorf("config file not found at %s", cfgErr.Path)
			case errors.Is(cfgErr, config.ErrInvalid):
				return fmt.Errorf("config file is invalid at %s", cfgErr.Path)
			}
		}
		return err
	}

	level := slog.LevelInfo
	if cfg.Debug {
		level = slog.LevelDebug
	}
	log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))

	prod := producer.NewLoggingMulti(
		producer.NewMulti(
			cfg.Kafka.Brokers,
			internalkafka.TopicEventPush,
			internalkafka.TopicEventPR,
			internalkafka.TopicEventIssue,
			internalkafka.TopicEventPipeline,
			internalkafka.TopicEventAnswer,
			internalkafka.TopicEventPost,
			internalkafka.TopicEventVideo,
		),
		log,
	)
	defer prod.Close()

	interval := 5 * time.Minute
	if cfg.Poller.Interval > 0 {
		interval = cfg.Poller.Interval
	}

	scheduler := poller.NewScheduler(
		prod,
		interval,
		log,
		poller.NewGitHubPoller(cfg.Poller.GitHubToken),
		poller.NewGitLabPoller(cfg.Poller.GitLabToken),
		poller.NewStackOverflowPoller(),
		poller.NewRedditPoller(),
		poller.NewYouTubePoller(cfg.Poller.YouTubeAPIKey),
	)

	c := consumer.New(cfg.Kafka.Brokers, "poller-group", log)

	c.Subscribe(internalkafka.TopicSubscriptionCreated, func(ctx context.Context, data []byte) error {
		var msg internalkafka.SubscriptionCreatedMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Error("unmarshal subscription created",
				slog.String("err", err.Error()),
			)
			return err
		}
		scheduler.Watch(msg.RepoURL)
		return nil
	})

	c.Subscribe(internalkafka.TopicSubscriptionDeleted, func(ctx context.Context, data []byte) error {
		var msg internalkafka.SubscriptionDeletedMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Error("unmarshal subscription deleted",
				slog.String("err", err.Error()),
			)
			return err
		}
		scheduler.Unwatch(msg.RepoURL)
		return nil
	})

	log.Info("poller started",
		slog.Duration("interval", interval),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		log.Info("received signal, shutting down")
		cancel()
	}()

	go func() {
		if err := scheduler.Run(ctx); err != nil {
			log.Error("scheduler error", slog.String("err", err.Error()))
		}
	}()

	<-ctx.Done()
	return c.Close()
}
