package postgres

import (
	"context"
	"fmt"

	"github.com/dlc-01/GitNotify/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, cfg *config.PostgresConfig) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName,
	)

	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, wrap("NewPool", ErrConnect)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, wrap("NewPool", ErrConnect)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, wrap("NewPool", ErrPing)
	}

	return pool, nil
}
