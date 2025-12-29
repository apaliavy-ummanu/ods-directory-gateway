package config

import (
	"time"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	Account        string        `env:"ACCOUNT"`
	HTTPPort       string        `env:"PORT" envDefault:"8080"`
	APIKey         string        `env:"API_KEY"`
	RequestTimeout time.Duration `env:"REQUEST_TIMEOUT" envDefault:"30s"`
	LogLevel       string        `env:"LOG_LEVEL" envDefault:"INFO"`
	ODSConfig      ODSConfig
}

type ODSConfig struct {
	ServerURL string `env:"ODS_FHIR_API_SERVER_URL"`
}

func NewConfigFromEnv() (Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}
