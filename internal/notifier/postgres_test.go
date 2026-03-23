//go:build integration

package notifier

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/dlc-01/GitNotify/internal/domain"
)

var (
	testPool *pgxpool.Pool
	testRepo Repository
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:16-alpine"),
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		panic("failed to start postgres container: " + err.Error())
	}
	defer container.Terminate(ctx)

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		panic("failed to get connection string: " + err.Error())
	}

	testPool, err = pgxpool.New(ctx, dsn)
	if err != nil {
		panic("failed to create pool: " + err.Error())
	}
	defer testPool.Close()

	sqlDB := stdlib.OpenDBFromPool(testPool)
	defer sqlDB.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		panic("failed to set dialect: " + err.Error())
	}
	if err := goose.Up(sqlDB, "../../migrations/notifier"); err != nil {
		panic("failed to run migrations: " + err.Error())
	}

	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	testRepo = NewPostgresRepository(testPool, log)

	os.Exit(m.Run())
}

func cleanDB(t *testing.T) {
	t.Helper()
	_, err := testPool.Exec(context.Background(), "TRUNCATE subscriptions")
	require.NoError(t, err)
}

func TestPostgresRepository_Subscribe_Success(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	err := testRepo.Subscribe(ctx, 123, "https://github.com/golang/go")
	require.NoError(t, err)
}

func TestPostgresRepository_Subscribe_Idempotent(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	err := testRepo.Subscribe(ctx, 123, "https://github.com/golang/go")
	require.NoError(t, err)

	err = testRepo.Subscribe(ctx, 123, "https://github.com/golang/go")
	require.NoError(t, err)
}

func TestPostgresRepository_Unsubscribe_Success(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	require.NoError(t, testRepo.Subscribe(ctx, 123, "https://github.com/golang/go"))

	err := testRepo.Unsubscribe(ctx, 123, "https://github.com/golang/go")
	require.NoError(t, err)
}

func TestPostgresRepository_Unsubscribe_NotFound(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	err := testRepo.Unsubscribe(ctx, 123, "https://github.com/golang/go")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestPostgresRepository_MuteEvent_Success(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	require.NoError(t, testRepo.Subscribe(ctx, 123, "https://github.com/golang/go"))

	err := testRepo.MuteEvent(ctx, 123, "https://github.com/golang/go", domain.EventPush)
	require.NoError(t, err)
}

func TestPostgresRepository_MuteEvent_Idempotent(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	require.NoError(t, testRepo.Subscribe(ctx, 123, "https://github.com/golang/go"))

	err := testRepo.MuteEvent(ctx, 123, "https://github.com/golang/go", domain.EventPush)
	require.NoError(t, err)

	err = testRepo.MuteEvent(ctx, 123, "https://github.com/golang/go", domain.EventPush)
	require.NoError(t, err)
}

func TestPostgresRepository_GetSubscribersByRepo_Success(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	require.NoError(t, testRepo.Subscribe(ctx, 123, "https://github.com/golang/go"))
	require.NoError(t, testRepo.Subscribe(ctx, 456, "https://github.com/golang/go"))

	chatIDs, err := testRepo.GetSubscribersByRepo(ctx, "https://github.com/golang/go", domain.EventPush)
	require.NoError(t, err)
	assert.Len(t, chatIDs, 2)
	assert.Contains(t, chatIDs, int64(123))
	assert.Contains(t, chatIDs, int64(456))
}

func TestPostgresRepository_GetSubscribersByRepo_FiltersMuted(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	require.NoError(t, testRepo.Subscribe(ctx, 123, "https://github.com/golang/go"))
	require.NoError(t, testRepo.Subscribe(ctx, 456, "https://github.com/golang/go"))

	require.NoError(t, testRepo.MuteEvent(ctx, 123, "https://github.com/golang/go", domain.EventPush))

	chatIDs, err := testRepo.GetSubscribersByRepo(ctx, "https://github.com/golang/go", domain.EventPush)
	require.NoError(t, err)
	assert.Len(t, chatIDs, 1)
	assert.Contains(t, chatIDs, int64(456))
	assert.NotContains(t, chatIDs, int64(123))
}

func TestPostgresRepository_GetSubscribersByRepo_Empty(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	chatIDs, err := testRepo.GetSubscribersByRepo(ctx, "https://github.com/golang/go", domain.EventPush)
	require.NoError(t, err)
	assert.Empty(t, chatIDs)
}

func TestPostgresRepository_GetSubscribersByRepo_DifferentEvents(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	require.NoError(t, testRepo.Subscribe(ctx, 123, "https://github.com/golang/go"))
	require.NoError(t, testRepo.MuteEvent(ctx, 123, "https://github.com/golang/go", domain.EventPush))

	chatIDs, err := testRepo.GetSubscribersByRepo(ctx, "https://github.com/golang/go", domain.EventPR)
	require.NoError(t, err)
	assert.Len(t, chatIDs, 1)

	chatIDs, err = testRepo.GetSubscribersByRepo(ctx, "https://github.com/golang/go", domain.EventPush)
	require.NoError(t, err)
	assert.Empty(t, chatIDs)
}
