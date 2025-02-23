package main

import (
	"github.com/veedubyou/chord-paper-be/src/shared/config"
	"github.com/veedubyou/chord-paper-be/src/shared/config/dev"
	"github.com/veedubyou/chord-paper-be/src/shared/config/envvar"
	"github.com/veedubyou/chord-paper-be/src/shared/config/local"
	"github.com/veedubyou/chord-paper-be/src/shared/config/prod"
	"github.com/veedubyou/chord-paper-be/src/shared/lib/env"
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
				StorageHost: prod.GOOGLE_STORAGE_HOST,
				SecretKey:   envvar.MustGet(envvar.GOOGLE_CLOUD_KEY),
				BucketName:  envvar.MustGet(envvar.GOOGLE_CLOUD_STORAGE_BUCKET_NAME),
			},
			RabbitMQURL:             envvar.MustGet(envvar.RABBITMQ_URL),
			RabbitMQQueueName:       envvar.MustGet(envvar.RABBITMQ_QUEUE_NAME),
			YoutubeDLBinPath:        config.YoutubeDLPath(),
			YoutubeDLWorkingDirPath: envvar.MustGet(envvar.YOUTUBEDL_WORKING_DIR_PATH),
			SpleeterBinPath:         config.SpleeterPath(),
			SpleeterWorkingDirPath:  envvar.MustGet(envvar.SPLEETER_WORKING_DIR_PATH),
			DemucsBinPath:           config.DemucsPath(),
			DemucsWorkingDirPath:    envvar.MustGet(envvar.DEMUCS_WORKING_DIR_PATH),
		}

	case env.Development:
		appConfig = application.Config{
			DynamoConfig: dev.DynamoConfig,
			// using prod GCS creds for now because the local fake GCS doesn't persist
			CloudStorageConfig: config.ProdCloudStorage{
				StorageHost: prod.GOOGLE_STORAGE_HOST,
				SecretKey:   envvar.MustGet(envvar.GOOGLE_CLOUD_KEY),
				BucketName:  envvar.MustGet(envvar.GOOGLE_CLOUD_STORAGE_BUCKET_NAME),
			},
			RabbitMQURL:             dev.RabbitMQHost,
			RabbitMQQueueName:       dev.RabbitMQQueueName,
			YoutubeDLBinPath:        config.YoutubeDLPath(),
			YoutubeDLWorkingDirPath: path.Join(local.ProjectRoot(), "/src/worker/wd/youtube-dl"),
			SpleeterBinPath:         config.SpleeterPath(),
			SpleeterWorkingDirPath:  path.Join(local.ProjectRoot(), "/src/worker/wd/spleeter"),
			DemucsBinPath:           config.DemucsPath(),
			DemucsWorkingDirPath:    path.Join(local.ProjectRoot(), "/src/worker/wd/demucs"),
		}
	default:
		panic("Unexpected environment")
	}

	app := application.NewApp(appConfig)
	if err := app.Start(); err != nil {
		panic(err)
	}
}
