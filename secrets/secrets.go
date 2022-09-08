package secrets

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"

	"github.com/rs/zerolog"
)

// RDSCredentials Define the structure of the RDS generated credentials JSON
type RDSCredentials struct {
	Username             string `json:"username"`
	Password             string `json:"password"`
	Engine               string `json:"engine"`
	Port                 int    `json:"port"`
	DbInstanceIdentifier string `json:"dbInstanceIdentifier"`
	Host                 string `json:"host"`
}

// GetRDSCredentials Get RDS Credentials from AWS Secrets Manager
func GetRDSCredentials(secretName string, log *zerolog.Logger) (creds RDSCredentials) {
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

	err = json.Unmarshal([]byte(bytes), &creds)
	if err != nil {
		log.Panic().Msg(err.Error())
	}
	return creds
}
