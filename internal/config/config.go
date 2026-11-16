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

func LoadConfig() (*Config, error) {
	var cfg Config

	if _, err := os.Stat(".env"); err == nil {
		if err := cleanenv.ReadConfig(".env", &cfg); err != nil {
			return nil, fmt.Errorf("Err reading env-file: %w\n", err)
		}
		return &cfg, nil
	}

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("Err reading env variables %w\n", err)
	}

	return &cfg, nil

}
