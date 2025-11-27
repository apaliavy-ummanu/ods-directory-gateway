package config

import (
	"os"
)

type Config struct {
	HTTPPort            string `env:"PORT" envDefault:"8080"`
	APIKey              string `env:"API_KEY"`
	OdsFhirAPIServerURL string `env:"ODS_FHIR_API_SERVER_URL"`
}

// todo properly load env variables
func NewConfigFromEnv() (Config, error) {
	return Config{
		HTTPPort:            os.Getenv("PORT"),
		APIKey:              os.Getenv("API_KEY"),
		OdsFhirAPIServerURL: "https://uat.directory.spineservices.nhs.uk/STU3",
	}, nil
}
