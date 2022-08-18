package dev

import "github.com/veedubyou/chord-paper-be/src/shared/config"

// DynamoDB
const (
	DynamoAccessKeyID     = "local"
	DynamoSecretAccessKey = "local"
	DynamoDBHost          = "http://localhost:8000"
	DynamoDBRegion        = "localhost"
)

var DynamoConfig = config.LocalDynamo{
	AccessKeyID:     DynamoAccessKeyID,
	SecretAccessKey: DynamoSecretAccessKey,
	Region:          DynamoDBRegion,
	Host:            DynamoDBHost,
}

// RabbitMQ
const (
	RabbitMQHost      = "amqp://localhost:5672"
	RabbitMQQueueName = "chord-paper-tracks-dev"
)
