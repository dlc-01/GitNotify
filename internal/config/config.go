package config

import (
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Debug    bool
	Bot      BotConfig
	Postgres PostgresConfig
	Kafka    KafkaConfig
}

type BotConfig struct {
	Token string
}

type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

type KafkaConfig struct {
	Brokers []string
}

func Load(path string) (*Config, error) {
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	viper.SetDefault("kafka.brokers", []string{"localhost:9092"})
	viper.SetDefault("postgres.port", 5432)
	viper.SetDefault("debug", false)

	if path != "" {
		viper.SetConfigFile(path)
		if err := viper.ReadInConfig(); err != nil {
			if os.IsNotExist(err) {
				return nil, wrap(path, ErrNotFound)
			}
			return nil, wrap(path, ErrInvalid)
		}
	}

	cfg := &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, wrap(path, ErrUnmarshal)
	}

	if len(cfg.Kafka.Brokers) == 1 && strings.Contains(cfg.Kafka.Brokers[0], ",") {
		cfg.Kafka.Brokers = strings.Split(cfg.Kafka.Brokers[0], ",")
	}

	return cfg, nil
}
