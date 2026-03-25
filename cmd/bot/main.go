package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/dlc-01/GitNotify/internal/bot"
	"github.com/dlc-01/GitNotify/internal/config"
	internalkafka "github.com/dlc-01/GitNotify/internal/kafka"
	"github.com/dlc-01/GitNotify/internal/kafka/producer"
	"github.com/dlc-01/GitNotify/internal/repository"
	"github.com/dlc-01/GitNotify/internal/repository/postgres"
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
	configFile := flag.String("config", "", "path to config yaml")
	envFile := flag.String("env-file", "", "path to .env file")
	flag.Parse()

	cfg, err := config.Load(config.Options{
		ConfigFile: *configFile,
		EnvFile:    *envFile,
	})
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

	fmt.Printf("postgres config: %+v\n", cfg.Postgres)

	level := slog.LevelInfo
	if cfg.Debug {
		level = slog.LevelDebug
	}
	log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))

	initCtx, initCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer initCancel()

	pool, err := postgres.NewPool(initCtx, &cfg.Postgres)
	if err != nil {
		var pgErr *postgres.Error
		if errors.As(err, &pgErr) {
			switch {
			case errors.Is(pgErr, postgres.ErrConnect):
				return fmt.Errorf("cannot connect to postgres: %w", err)
			case errors.Is(pgErr, postgres.ErrPing):
				return fmt.Errorf("postgres is unreachable: %w", err)
			}
		}
		return err
	}
	defer pool.Close()
	log.Info("connected to postgres")

	sqlDB := stdlib.OpenDBFromPool(pool)
	defer sqlDB.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("goose set dialect: %w", err)
	}
	if err := goose.Up(sqlDB, "migrations/bot"); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}
	log.Info("migrations applied")

	repo := repository.NewLoggingRepository(
		postgres.New(pool),
		log,
	)

	prod := producer.NewLoggingMulti(
		producer.NewMulti(
			cfg.Kafka.Brokers,
			internalkafka.TopicSubscriptionCreated,
			internalkafka.TopicSubscriptionDeleted,
			internalkafka.TopicSubscriptionMuted,
			internalkafka.TopicSubscriptionUnmuted,
		),
		log,
	)
	defer prod.Close()

	log.Info("kafka producers initialized")

	app, err := bot.New(cfg.Bot.Token, repo, log, prod)
	if err != nil {
		return fmt.Errorf("init bot: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	return app.Run(ctx)
}
