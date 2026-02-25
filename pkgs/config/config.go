package config

import (
	"log/slog"
	"time"

	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv string `env:"APP_ENV" envDefault:"development"`

	DBDSN          string        `env:"DATABASE_URL,required"`
	DBMaxOpenConns int           `env:"DB_MAX_OPEN_CONNS" envDefault:"25"`
	DBConnLifetime time.Duration `env:"DB_CONN_LIFETIME" envDefault:"5m"`

	RedisURL string `env:"REDIS_URL,required"`
}

func Load() *Config {
	_ = godotenv.Load()

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		slog.Error("Unable to parse config", "Error", err)
	}

	return cfg
}
