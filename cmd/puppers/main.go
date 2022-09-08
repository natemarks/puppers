package main

// Package main checks access to an RDS instance and writes json logs to puppers.log
import (
	"os"

	"github.com/natemarks/puppers/secrets"

	"github.com/natemarks/puppers"
	"github.com/rs/zerolog"

	"github.com/natemarks/ec2metadata"
)

const (
	secretName = "SecretA720EF05-2pmGVjf2abKX"
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
	instanceID, err := ec2metadata.GetV2("instance-id")
	if err == nil {
		log = log.With().Str("instance-id", instanceID).Logger()
	}
	creds := secrets.GetRDSCredentials(secretName, &log)
	log.Info().Msgf("found credentials for hostname: %s", creds.Host)
}
