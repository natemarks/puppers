package secrets

import (
	"context"
	"encoding/json"
	"errors"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/natemarks/postgr8/command"

	"github.com/rs/zerolog"
)

// GetRDSCredentials Get RDS Credentials from AWS Secrets Manager
func GetRDSCredentials(secretName string, log *zerolog.Logger) (connectionParams command.InstanceConnectionParams) {
	log.Info().Msg("getting credentials from secrets manager")

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Panic().Msg(err.Error())
	}

	client := *secretsmanager.NewFromConfig(cfg)

	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	}
	result, err := client.GetSecretValue(context.TODO(), input)
	if err != nil {
		log.Panic().Msg(err.Error())
	}
	bytes := []byte(*result.SecretString)

	err = json.Unmarshal([]byte(bytes), &connectionParams)
	if err != nil {
		log.Panic().Msg(err.Error())
	}
	log.Info().Msgf("found credentials for hostname: %s", connectionParams.Host)
	return connectionParams
}

// GetSecretFromEnvar get the RDS secret name from PUPPERS_SECRET_NAME env var
func GetSecretFromEnvar() string {
	// secretName "SecretA720EF05-2pmGVjf2abKX"
	secretName, set := os.LookupEnv("PUPPERS_SECRET_NAME")
	if !set {
		panic(errors.New("PUPPERS_SECRET_NAME is not set"))
	}
	return secretName
}
