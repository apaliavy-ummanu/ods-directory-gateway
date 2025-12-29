package elog

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func Init() {
	if os.Getenv("APP_ENV") == "local" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}) // good-looking logger in local env
	} else {
		zerolog.TimeFieldFormat = time.RFC3339
		log.Logger = log.Hook(SeverityHook{})
	}

	logLevel := os.Getenv("LOG_LEVEL")
	switch logLevel {
	case "DEBUG":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "INFO":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "WARN":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "ERROR":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		log.Panic().Msgf("Unrecognised LOG_LEVEL set: %s", logLevel)
	}
}

type SeverityHook struct{}

func (h SeverityHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	// if level == zerolog.ErrorLevel || level == zerolog.FatalLevel {
	//TODO: report to e.g. rollbar here
	// }
}
