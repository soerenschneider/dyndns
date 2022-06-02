package util

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/internal"
	"golang.org/x/crypto/ssh/terminal"
	"os"
	"time"
)

func InitLogging() {
	if terminal.IsTerminal(int(os.Stderr.Fd())) {
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stderr,
			NoColor:    true,
			TimeFormat: time.RFC3339,
		})
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	}
	zerolog.TimestampFieldName = "timestampSeconds"

	log.Info().Msgf("Started dyndns version %s, commit %s", internal.BuildVersion, internal.CommitHash)
}
