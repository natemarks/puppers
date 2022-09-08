package main

// Package main runs on a loop creating JSON log messages and pushing them to cloudwatch
// I stole the logic entirely from https://github.com/mathisve/golang-cloudwatch-logs-example
import (
	"os"

	"github.com/natemarks/puppers"
	"github.com/rs/zerolog"

	nmec2 "github.com/natemarks/ec2metadata"
)

func main() {
	logFile, err := os.OpenFile("puppers.log",
		os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	defer logFile.Close()
	log := zerolog.New(logFile).With().Timestamp().Logger()
	log = log.With().Str("version", puppers.Version).Logger()
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Info().Msg("puppers is starting")
	instanceID, err := nmec2.GetV2("instance-id")
	if err == nil {
		log = log.With().Str("instance-id", instanceID).Logger()
	}
}
