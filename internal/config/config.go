package config

import (
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Bot      BotConfig
	Postgres PostgresConfig
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

func Load(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if os.IsNotExist(err) {
			return nil, wrap(path, ErrNotFound)
		}
		return nil, wrap(path, ErrInvalid)
	}

	cfg := &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, wrap(path, ErrUnmarshal)
	}
	return cfg, nil
}
