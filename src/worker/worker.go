package main

import (
	"github.com/veedubyou/chord-paper-be/src/shared/config"
	"github.com/veedubyou/chord-paper-be/src/shared/lib/env"
	"github.com/veedubyou/chord-paper-be/src/shared/values/dev"
	"github.com/veedubyou/chord-paper-be/src/shared/values/envvar"
	"github.com/veedubyou/chord-paper-be/src/shared/values/local"
	"github.com/veedubyou/chord-paper-be/src/shared/values/prod"
	"github.com/veedubyou/chord-paper-be/src/worker/application"
	"path"
)

func main() {
	var appConfig application.Config

	switch env.Get() {
	case env.Production:
		appConfig = application.Config{
			DynamoConfig: config.ProdDynamo{
				AccessKeyID:     envvar.MustGet(envvar.AWS_ACCESS_KEY_ID),
				SecretAccessKey: envvar.MustGet(envvar.AWS_SECRET_ACCESS_KEY),
				Region:          prod.DynamoDBRegion,
			},
			CloudStorageConfig: config.ProdCloudStorage{
				StorageHost: config.GOOGLE_STORAGE_HOST,
				SecretKey:   envvar.MustGet(envvar.GOOGLE_CLOUD_KEY),
				BucketName:  envvar.MustGet(envvar.GOOGLE_CLOUD_STORAGE_BUCKET_NAME),
			},
			RabbitMQURL:             envvar.MustGet(envvar.RABBITMQ_URL),
			RabbitMQQueueName:       envvar.MustGet(envvar.RABBITMQ_QUEUE_NAME),
			YoutubeDLBinPath:        envvar.MustGet("YOUTUBEDL_BIN_PATH"),
			YoutubeDLWorkingDirPath: envvar.MustGet("YOUTUBEDL_WORKING_DIR_PATH"),
			SpleeterBinPath:         envvar.MustGet("SPLEETER_BIN_PATH"),
			SpleeterWorkingDirPath:  envvar.MustGet("SPLEETER_WORKING_DIR_PATH"),
		}

	case env.Development:
		appConfig = application.Config{
			DynamoConfig: dev.DynamoConfig,
			CloudStorageConfig: config.LocalCloudStorage{
				//TODO
				StorageHost:  "",
				HostEndpoint: "",
				BucketName:   envvar.MustGet(envvar.GOOGLE_CLOUD_STORAGE_BUCKET_NAME),
			},
			RabbitMQURL:             dev.RabbitMQHost,
			RabbitMQQueueName:       dev.RabbitMQQueueName,
			YoutubeDLBinPath:        envvar.MustGet("YOUTUBEDL_BIN_PATH"),
			YoutubeDLWorkingDirPath: path.Join(local.ProjectRoot(), "/src/worker/wd/youtube-dl"),
			SpleeterBinPath:         envvar.MustGet("SPLEETER_BIN_PATH"),
			SpleeterWorkingDirPath:  path.Join(local.ProjectRoot(), "/src/worker/wd/spleeter"),
		}
	default:
		panic("Unexpected environment")
	}

	app := application.NewApp(appConfig)
	if err := app.Start(); err != nil {
		panic(err)
	}
}
