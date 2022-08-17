package main

import (
	"github.com/veedubyou/chord-paper-be/src/server/application"
	"github.com/veedubyou/chord-paper-be/src/server/google_id"
	"github.com/veedubyou/chord-paper-be/src/shared/config"
	"github.com/veedubyou/chord-paper-be/src/shared/lib/env"
	"github.com/veedubyou/chord-paper-be/src/shared/values/dev"
	"github.com/veedubyou/chord-paper-be/src/shared/values/envvar"
	"github.com/veedubyou/chord-paper-be/src/shared/values/prod"
	"strings"
)

const (
	// should get injected as an env var, but YAGNI for now as it's not a secret
	// and there's no other case to reflect this need
	googleClientID = "650853277550-ta69qbfcvdl6tb5ogtnh2d07ae9rcdlf.apps.googleusercontent.com"
)

func main() {
	var appConfig application.Config

	switch env.Get() {
	case env.Production:
		commaSeparatedOrigins := envvar.MustGet("ALLOWED_FE_ORIGINS")
		allowedOrigins := strings.Split(commaSeparatedOrigins, ",")

		appConfig = application.Config{
			DynamoConfig: config.ProdDynamo{
				AccessKeyID:     envvar.MustGet(envvar.AWS_ACCESS_KEY_ID),
				SecretAccessKey: envvar.MustGet(envvar.AWS_SECRET_ACCESS_KEY),
				Region:          prod.DynamoDBRegion,
			},
			RabbitMQURL:        envvar.MustGet(envvar.RABBITMQ_URL),
			RabbitMQQueueName:  envvar.MustGet(envvar.RABBITMQ_QUEUE_NAME),
			CORSAllowedOrigins: allowedOrigins,
			UserValidator:      google_id.GoogleValidator{ClientID: googleClientID},
			Port:               ":5000",
			Log:                true,
		}
	case env.Development:
		appConfig = application.Config{
			DynamoConfig:       dev.DynamoConfig,
			RabbitMQURL:        dev.RabbitMQHost,
			RabbitMQQueueName:  dev.RabbitMQQueueName,
			CORSAllowedOrigins: []string{"*"},
			UserValidator:      google_id.GoogleValidator{ClientID: googleClientID},
			Port:               ":5000",
			Log:                true,
		}

	default:
		panic("Unexpected environment")
	}

	app := application.NewApp(appConfig)
	if err := app.Start(); err != nil {
		panic(err)
	}
}
