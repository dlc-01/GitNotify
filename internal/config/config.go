package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Debug    bool           `mapstructure:"debug"`
	Bot      BotConfig      `mapstructure:"bot"`
	Postgres PostgresConfig `mapstructure:"postgres"`
	Kafka    KafkaConfig    `mapstructure:"kafka"`
	Webhook  WebhookConfig  `mapstructure:"webhook"`
	Poller   PollerConfig   `mapstructure:"poller"`
}

type BotConfig struct {
	Token string `mapstructure:"token"`
}

type PostgresConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
}

type KafkaConfig struct {
	Brokers []string `mapstructure:"brokers"`
}

type WebhookConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	GitHubSecret string `mapstructure:"githubsecret"`
	GitLabSecret string `mapstructure:"gitlabsecret"`
}

type PollerConfig struct {
	Interval      time.Duration `mapstructure:"interval"`
	GitHubToken   string        `mapstructure:"githubtoken"`
	GitLabToken   string        `mapstructure:"gitlabtoken"`
	YouTubeAPIKey string        `mapstructure:"youtubeapikey"`
}

type Options struct {
	ConfigFile string
	EnvFile    string
}

func Load(opts Options) (*Config, error) {
	if opts.EnvFile != "" {
		if err := godotenv.Load(opts.EnvFile); err != nil {
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("load env file: %w", err)
			}
		}
	}

	v := viper.New()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	envBindings := map[string]string{
		"debug":                "DEBUG",
		"bot.token":            "BOT_TOKEN",
		"postgres.host":        "POSTGRES_HOST",
		"postgres.port":        "POSTGRES_PORT",
		"postgres.user":        "POSTGRES_USER",
		"postgres.password":    "POSTGRES_PASSWORD",
		"postgres.dbname":      "POSTGRES_DBNAME",
		"kafka.brokers":        "KAFKA_BROKERS",
		"webhook.host":         "WEBHOOK_HOST",
		"webhook.port":         "WEBHOOK_PORT",
		"webhook.githubsecret": "WEBHOOK_GITHUBSECRET",
		"webhook.gitlabsecret": "WEBHOOK_GITLABSECRET",
		"poller.interval":      "POLLER_INTERVAL",
		"poller.githubtoken":   "POLLER_GITHUBTOKEN",
		"poller.gitlabtoken":   "POLLER_GITLABTOKEN",
		"poller.youtubeapikey": "POLLER_YOUTUBEAPIKEY",
	}

	for key, envName := range envBindings {
		if err := v.BindEnv(key, envName); err != nil {
			return nil, fmt.Errorf("bind env %s: %w", envName, err)
		}
	}

	if opts.ConfigFile != "" {
		v.SetConfigFile(opts.ConfigFile)
		v.SetDefault("debug", false)
		v.SetDefault("postgres.port", 5432)
		v.SetDefault("webhook.port", 8080)
		v.SetDefault("kafka.brokers", []string{"localhost:9092"})
		v.SetDefault("poller.interval", 5*time.Minute)

		if err := v.ReadInConfig(); err != nil {
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("config file not found: %w", err)
			}
			return nil, fmt.Errorf("read config file: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if len(cfg.Kafka.Brokers) == 1 && strings.Contains(cfg.Kafka.Brokers[0], ",") {
		cfg.Kafka.Brokers = strings.Split(cfg.Kafka.Brokers[0], ",")
	}

	return &cfg, nil
}
