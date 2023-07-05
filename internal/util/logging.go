package util

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/internal"
	"os"
	"time"
)

func InitLogging() {
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		NoColor:    true,
		TimeFormat: time.RFC3339,
	})
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	log.Info().Msgf("Started dyndns version %s, commit %s", internal.BuildVersion, internal.CommitHash)
}
