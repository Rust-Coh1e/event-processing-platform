package config

import (
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Addr         string        `env:"HTTP_ADDR" env-default:":8080"`
	ShutdownTOut time.Duration `env:"SHUTDOWN_TIMEOUT" env-default:"5s"`
}

// LoadConfig загружает конфигурацию из указанного файла (если он существует) и переменных окружения.
func LoadConfig(envPath string) (*Config, error) {
	var cfg Config

	if envPath != "" {
		if _, err := os.Stat(envPath); err == nil {
			if err := cleanenv.ReadConfig(envPath, &cfg); err != nil {
				return nil, fmt.Errorf("error reading env-file %q: %w", envPath, err)
			}
			return &cfg, nil
		}
	}

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("error reading env variables: %w", err)
	}

	return &cfg, nil
}
