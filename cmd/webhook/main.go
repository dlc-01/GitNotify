package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/dlc-01/GitNotify/internal/config"
	internalkafka "github.com/dlc-01/GitNotify/internal/kafka"
	"github.com/dlc-01/GitNotify/internal/kafka/producer"
	"github.com/dlc-01/GitNotify/internal/webhook"
	"github.com/dlc-01/GitNotify/internal/webhook/parser"
	"github.com/dlc-01/GitNotify/internal/webhook/validator"
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

	log.Info("kafka producers initialized")

	handler := webhook.NewHandler(prod, log)

	handler.RegisterValidator(validator.NewGitHubValidator(cfg.Webhook.GitHubSecret))
	handler.RegisterValidator(validator.NewGitLabValidator(cfg.Webhook.GitLabSecret))

	handler.RegisterParser(parser.NewGitHubParser())
	handler.RegisterParser(parser.NewGitLabParser())

	srv := webhook.NewServer(webhook.DefaultServerConfig(), handler, log)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		cancel()
	}()

	return srv.Run(ctx)
}
