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
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/dlc-01/GitNotify/internal/config"
	"github.com/dlc-01/GitNotify/internal/notifier"
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

	level := slog.LevelInfo
	if cfg.Debug {
		level = slog.LevelDebug
	}
	log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))

	initCtx, initCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer initCancel()

	pool, err := pgxpool.New(initCtx, fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.Postgres.User,
		cfg.Postgres.Password,
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		cfg.Postgres.DBName,
	))
	if err != nil {
		return fmt.Errorf("connect to postgres: %w", err)
	}
	defer pool.Close()

	if err := pool.Ping(initCtx); err != nil {
		return fmt.Errorf("ping postgres: %w", err)
	}
	log.Info("connected to postgres")

	sqlDB := stdlib.OpenDBFromPool(pool)
	defer sqlDB.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("goose set dialect: %w", err)
	}
	if err := goose.Up(sqlDB, "migrations/notifier"); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}
	log.Info("migrations applied")

	repo := notifier.NewPostgresRepository(pool, log)

	sender, err := notifier.NewSender(cfg.Bot.Token, log)
	if err != nil {
		return fmt.Errorf("init sender: %w", err)
	}

	handler := notifier.NewHandler(repo, sender, log)

	app := notifier.New(
		cfg.Kafka.Brokers,
		"notifier-group",
		handler,
		log,
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

	return app.Run(ctx)
}
