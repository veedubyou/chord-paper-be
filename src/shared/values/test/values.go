package test

import (
	"github.com/veedubyou/chord-paper-be/src/shared/config"
	"github.com/veedubyou/chord-paper-be/src/shared/values/dev"
)

// DynamoDB
const (
	DynamoAccessKeyID     = dev.DynamoAccessKeyID
	DynamoSecretAccessKey = dev.DynamoSecretAccessKey
	DynamoDBHost          = dev.DynamoDBHost
)

func DynamoConfig(region string) config.LocalDynamo {
	return config.LocalDynamo{
		AccessKeyID:     DynamoAccessKeyID,
		SecretAccessKey: DynamoSecretAccessKey,
		Region:          region,
		Host:            DynamoDBHost,
	}
}

// RabbitMQ
const (
	RabbitMQHost = dev.RabbitMQHost
)
