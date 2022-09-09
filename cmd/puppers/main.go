package main

// Package main checks access to an RDS instance and writes json logs to puppers.log
import (
	"errors"
	"os"

	"github.com/natemarks/postgr8/command"

	"github.com/natemarks/puppers/secrets"

	"github.com/natemarks/puppers"
	"github.com/rs/zerolog"

	"github.com/natemarks/ec2metadata"
)

func getSecretFromEnvar() string {
	// secretName "SecretA720EF05-2pmGVjf2abKX"
	secretName, set := os.LookupEnv("PUPPERS_SECRET_NAME")
	if !set {
		panic(errors.New("PUPPERS_SECRET_NAME is not set"))
	}
	return secretName
}

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
	instanceID, err := ec2metadata.GetAWSEc2Metadata("instance-id")
	if err == nil {
		log = log.With().Str("instance-id", instanceID).Logger()
	}
	creds := secrets.GetRDSCredentials(getSecretFromEnvar(), &log)
	log.Info().Msgf("found credentials for hostname: %s", creds.Host)
	// Check connectivity to database instance
	if !command.TCPOk(creds, 30) {
		log.Panic().Msgf("TCP Connection Failure: %s:%d", creds.Host, creds.Port)
	}
	log.Info().Msgf("TCP Connection Success: %s:%d", creds.Host, creds.Port)

	// Make sure the credentials are valid
	validCreds, err := command.ValidCredentials(creds)
	if !validCreds {
		log.Panic().Msg(err.Error())
	}
	log.Info().Msg("database credentials are valid")
	conn, err := command.NewInstanceConn(creds)
	if err != nil {
		log.Panic().Msg(err.Error())
	}
	dbNames, err := command.ListDatabases(conn)
	if err != nil {
		log.Panic().Msg(err.Error())
	}
	log.Info().Msgf("Found databases in instance: %d", len(dbNames))
}
