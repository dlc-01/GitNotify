package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load(Options{ConfigFile: "/nonexistent/path/config.yaml"})
	require.Error(t, err)
}

func TestLoad_InvalidYaml(t *testing.T) {
	f, err := os.CreateTemp("", "config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(f.Name())

	_, err = f.WriteString("invalid: yaml: content: [")
	require.NoError(t, err)
	f.Close()

	_, err = Load(Options{ConfigFile: f.Name()})
	require.Error(t, err)
}

func TestLoad_ValidYaml(t *testing.T) {
	f, err := os.CreateTemp("", "config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(f.Name())

	_, err = f.WriteString(`
debug: true
bot:
  token: test-token
postgres:
  host: localhost
  port: 5432
  user: testuser
  password: testpass
  dbname: testdb
kafka:
  brokers:
    - localhost:9092
    - localhost:9093
`)
	require.NoError(t, err)
	f.Close()

	cfg, err := Load(Options{ConfigFile: f.Name()})
	require.NoError(t, err)

	assert.True(t, cfg.Debug)
	assert.Equal(t, "test-token", cfg.Bot.Token)
	assert.Equal(t, "localhost", cfg.Postgres.Host)
	assert.Equal(t, 5432, cfg.Postgres.Port)
	assert.Equal(t, "testuser", cfg.Postgres.User)
	assert.Equal(t, "testpass", cfg.Postgres.Password)
	assert.Equal(t, "testdb", cfg.Postgres.DBName)
	assert.Equal(t, []string{"localhost:9092", "localhost:9093"}, cfg.Kafka.Brokers)
}

func TestLoad_EmptyPath_EnvOnly(t *testing.T) {
	os.Setenv("BOT_TOKEN", "env-token")
	os.Setenv("POSTGRES_HOST", "env-host")
	os.Setenv("POSTGRES_PORT", "5433")
	os.Setenv("POSTGRES_USER", "env-user")
	os.Setenv("POSTGRES_PASSWORD", "env-pass")
	os.Setenv("POSTGRES_DBNAME", "env-db")
	os.Setenv("KAFKA_BROKERS", "kafka1:9092,kafka2:9092")
	os.Setenv("DEBUG", "true")
	defer func() {
		os.Unsetenv("BOT_TOKEN")
		os.Unsetenv("POSTGRES_HOST")
		os.Unsetenv("POSTGRES_PORT")
		os.Unsetenv("POSTGRES_USER")
		os.Unsetenv("POSTGRES_PASSWORD")
		os.Unsetenv("POSTGRES_DBNAME")
		os.Unsetenv("KAFKA_BROKERS")
		os.Unsetenv("DEBUG")
	}()

	cfg, err := Load(Options{})
	require.NoError(t, err)

	assert.True(t, cfg.Debug)
	assert.Equal(t, "env-token", cfg.Bot.Token)
	assert.Equal(t, "env-host", cfg.Postgres.Host)
	assert.Equal(t, "env-user", cfg.Postgres.User)
	assert.Equal(t, []string{"kafka1:9092", "kafka2:9092"}, cfg.Kafka.Brokers)
}

func TestLoad_EnvOverridesYaml(t *testing.T) {
	f, err := os.CreateTemp("", "config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(f.Name())

	_, err = f.WriteString(`
bot:
  token: yaml-token
postgres:
  host: yaml-host
  port: 5432
  user: yaml-user
  password: yaml-pass
  dbname: yaml-db
`)
	require.NoError(t, err)
	f.Close()

	os.Setenv("BOT_TOKEN", "env-token")
	defer os.Unsetenv("BOT_TOKEN")

	cfg, err := Load(Options{ConfigFile: f.Name()})
	require.NoError(t, err)

	assert.Equal(t, "env-token", cfg.Bot.Token)
	assert.Equal(t, "yaml-host", cfg.Postgres.Host)
}

func TestLoad_Defaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.yaml")
	require.NoError(t, os.WriteFile(path, []byte("debug: false\n"), 0644))

	cfg, err := Load(Options{ConfigFile: path})
	require.NoError(t, err)

	assert.Equal(t, 5432, cfg.Postgres.Port)
	assert.Equal(t, []string{"localhost:9092"}, cfg.Kafka.Brokers)
}

func TestLoad_KafkaBrokersCommaSeparated(t *testing.T) {
	os.Setenv("KAFKA_BROKERS", "kafka1:9092,kafka2:9092,kafka3:9092")
	defer os.Unsetenv("KAFKA_BROKERS")

	cfg, err := Load(Options{})
	require.NoError(t, err)

	assert.Equal(t, []string{"kafka1:9092", "kafka2:9092", "kafka3:9092"}, cfg.Kafka.Brokers)
}
