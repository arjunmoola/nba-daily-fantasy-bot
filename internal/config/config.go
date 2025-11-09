package config

import (
	"os"
)

type Config struct {
	DatabaseURL string
	Token string
	ServerUrl string
}

func New() *Config {
	cfg := &Config{}
	cfg.Token = os.Getenv("TOKEN")
	cfg.DatabaseURL = os.Getenv("DATABASE_URL")
	cfg.ServerUrl = os.Getenv("URL")
	return cfg
}
