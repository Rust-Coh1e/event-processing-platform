package config

import (
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Port string `env:"HTTP_ADDR" env-default:"8080"`
}

func LoadConfig() *Config {
	var cfg Config

	if _, err := os.Stat(".env"); err == nil {
		if err := cleanenv.ReadConfig(".env", &cfg); err != nil {
			log.Fatalf("Err reading env-file\n")
		}
		return &cfg
	}

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("Err reading env variables\n")
	}

	return &cfg

}
