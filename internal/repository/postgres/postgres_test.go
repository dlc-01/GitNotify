//go:build integration

package postgres

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
	"github.com/dlc-01/GitNotify/internal/repository"
)

var (
	testPool *pgxpool.Pool
	testRepo repository.Repository
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
	if err := goose.Up(sqlDB, "../../../migrations/bot"); err != nil {
		panic("failed to run migrations: " + err.Error())
	}

	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	testRepo = repository.NewLoggingRepository(New(testPool), log)

	os.Exit(m.Run())
}

func cleanDB(t *testing.T) {
	t.Helper()
	_, err := testPool.Exec(context.Background(), "TRUNCATE subscriptions, chats, users CASCADE")
	require.NoError(t, err)
}

func TestUpsertUser(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	user := &domain.User{
		UserID:   1,
		Username: "testuser",
	}

	err := testRepo.UpsertUser(ctx, user)
	require.NoError(t, err)

	err = testRepo.UpsertUser(ctx, &domain.User{
		UserID:   1,
		Username: "updateduser",
	})
	require.NoError(t, err)
}

func TestUpsertChat(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	err := testRepo.UpsertUser(ctx, &domain.User{UserID: 1, Username: "test"})
	require.NoError(t, err)

	chat := &domain.Chat{
		ChatID:   100,
		ChatType: domain.ChatPrivate,
	}

	err = testRepo.UpsertChat(ctx, chat)
	require.NoError(t, err)

	err = testRepo.UpsertChat(ctx, &domain.Chat{
		ChatID:   100,
		ChatType: domain.ChatGroup,
	})
	require.NoError(t, err)
}

func TestSubscribe_Success(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	require.NoError(t, testRepo.UpsertUser(ctx, &domain.User{UserID: 1, Username: "test"}))
	require.NoError(t, testRepo.UpsertChat(ctx, &domain.Chat{ChatID: 100, ChatType: domain.ChatPrivate}))

	sub, err := testRepo.Subscribe(ctx, 100, "https://github.com/golang/go")
	require.NoError(t, err)
	assert.Equal(t, int64(100), sub.ChatID)
	assert.Equal(t, "https://github.com/golang/go", sub.RepoURL)
	assert.Empty(t, sub.MutedEvents)
}

func TestSubscribe_AlreadyExists(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	require.NoError(t, testRepo.UpsertUser(ctx, &domain.User{UserID: 1, Username: "test"}))
	require.NoError(t, testRepo.UpsertChat(ctx, &domain.Chat{ChatID: 100, ChatType: domain.ChatPrivate}))

	_, err := testRepo.Subscribe(ctx, 100, "https://github.com/golang/go")
	require.NoError(t, err)

	_, err = testRepo.Subscribe(ctx, 100, "https://github.com/golang/go")
	require.Error(t, err)

	var repoErr *repository.Error
	require.ErrorAs(t, err, &repoErr)
	assert.ErrorIs(t, repoErr, repository.ErrAlreadyExists)
}

func TestUnsubscribe_Success(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	require.NoError(t, testRepo.UpsertUser(ctx, &domain.User{UserID: 1, Username: "test"}))
	require.NoError(t, testRepo.UpsertChat(ctx, &domain.Chat{ChatID: 100, ChatType: domain.ChatPrivate}))
	_, err := testRepo.Subscribe(ctx, 100, "https://github.com/golang/go")
	require.NoError(t, err)

	err = testRepo.Unsubscribe(ctx, 100, "https://github.com/golang/go")
	require.NoError(t, err)
}

func TestUnsubscribe_NotFound(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	err := testRepo.Unsubscribe(ctx, 100, "https://github.com/golang/go")
	require.Error(t, err)

	var repoErr *repository.Error
	require.ErrorAs(t, err, &repoErr)
	assert.ErrorIs(t, repoErr, repository.ErrNotFound)
}

func TestListSubscriptions_Empty(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	require.NoError(t, testRepo.UpsertUser(ctx, &domain.User{UserID: 1, Username: "test"}))
	require.NoError(t, testRepo.UpsertChat(ctx, &domain.Chat{ChatID: 100, ChatType: domain.ChatPrivate}))

	subs, err := testRepo.ListSubscriptions(ctx, 100)
	require.NoError(t, err)
	assert.Empty(t, subs)
}

func TestListSubscriptions_Multiple(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	require.NoError(t, testRepo.UpsertUser(ctx, &domain.User{UserID: 1, Username: "test"}))
	require.NoError(t, testRepo.UpsertChat(ctx, &domain.Chat{ChatID: 100, ChatType: domain.ChatPrivate}))

	_, err := testRepo.Subscribe(ctx, 100, "https://github.com/golang/go")
	require.NoError(t, err)
	_, err = testRepo.Subscribe(ctx, 100, "https://github.com/torvalds/linux")
	require.NoError(t, err)

	subs, err := testRepo.ListSubscriptions(ctx, 100)
	require.NoError(t, err)
	assert.Len(t, subs, 2)
}

func TestMuteEvent_Success(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	require.NoError(t, testRepo.UpsertUser(ctx, &domain.User{UserID: 1, Username: "test"}))
	require.NoError(t, testRepo.UpsertChat(ctx, &domain.Chat{ChatID: 100, ChatType: domain.ChatPrivate}))
	_, err := testRepo.Subscribe(ctx, 100, "https://github.com/golang/go")
	require.NoError(t, err)

	err = testRepo.MuteEvent(ctx, 100, "https://github.com/golang/go", domain.EventPush)
	require.NoError(t, err)

	subs, err := testRepo.ListSubscriptions(ctx, 100)
	require.NoError(t, err)
	require.Len(t, subs, 1)
	assert.Contains(t, subs[0].MutedEvents, domain.EventPush)
}

func TestMuteEvent_Idempotent(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	require.NoError(t, testRepo.UpsertUser(ctx, &domain.User{UserID: 1, Username: "test"}))
	require.NoError(t, testRepo.UpsertChat(ctx, &domain.Chat{ChatID: 100, ChatType: domain.ChatPrivate}))
	_, err := testRepo.Subscribe(ctx, 100, "https://github.com/golang/go")
	require.NoError(t, err)

	err = testRepo.MuteEvent(ctx, 100, "https://github.com/golang/go", domain.EventPush)
	require.NoError(t, err)
	err = testRepo.MuteEvent(ctx, 100, "https://github.com/golang/go", domain.EventPush)
	require.NoError(t, err)

	subs, err := testRepo.ListSubscriptions(ctx, 100)
	require.NoError(t, err)
	assert.Len(t, subs[0].MutedEvents, 1)
}

func TestMuteEvent_NotFound(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	err := testRepo.MuteEvent(ctx, 100, "https://github.com/golang/go", domain.EventPush)
	require.Error(t, err)

	var repoErr *repository.Error
	require.ErrorAs(t, err, &repoErr)
	assert.ErrorIs(t, repoErr, repository.ErrNotFound)
}

func TestCascadeDelete(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	require.NoError(t, testRepo.UpsertUser(ctx, &domain.User{UserID: 1, Username: "test"}))
	require.NoError(t, testRepo.UpsertChat(ctx, &domain.Chat{ChatID: 100, ChatType: domain.ChatPrivate}))
	_, err := testRepo.Subscribe(ctx, 100, "https://github.com/golang/go")
	require.NoError(t, err)

	_, err = testPool.Exec(ctx, "DELETE FROM chats WHERE chat_id = 100")
	require.NoError(t, err)

	subs, err := testRepo.ListSubscriptions(ctx, 100)
	require.NoError(t, err)
	assert.Empty(t, subs)
}
