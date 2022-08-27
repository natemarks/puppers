package main

import (
	"os"
	"time"

	"github.com/natemarks/puppers/version"
	"github.com/rs/zerolog"
)

func getLogger() (logMe *zerolog.Logger) {
	logFile, err := os.OpenFile("puppers.log",
		os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	logger := zerolog.New(logFile).With().Timestamp().Logger()
	logger = logger.With().Str("version", version.Version).Logger()
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	return &logger
}

func main() {
	logger := getLogger()
	logger.Info().Msg("Puppers Starting")
	for {
		logger.Info().Msg("I'm alive!!")
		time.Sleep(4 * time.Second)
	}
}
