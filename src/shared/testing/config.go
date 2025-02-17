package testing

import (
	server_app "github.com/veedubyou/chord-paper-be/src/server/application"
	"github.com/veedubyou/chord-paper-be/src/shared/config"
	"github.com/veedubyou/chord-paper-be/src/shared/config/dev"
	"github.com/veedubyou/chord-paper-be/src/shared/config/envvar"
	"github.com/veedubyou/chord-paper-be/src/shared/config/local"
	worker_app "github.com/veedubyou/chord-paper-be/src/worker/application"
	"path"
)

func ServerConfig(dbRegion string) server_app.Config {
	return server_app.Config{
		DynamoConfig:       DynamoConfig(dbRegion),
		RabbitMQURL:        RabbitMQHost,
		RabbitMQQueueName:  RabbitMQQueueName,
		CORSAllowedOrigins: []string{"*"},
		UserValidator:      Validator{},
		Port:               ServerPort,
		Log:                false,
	}
}

func WorkerConfig(dbRegion string, cloudStorageConfig config.LocalCloudStorage) worker_app.Config {
	return worker_app.Config{
		DynamoConfig:            DynamoConfig(dbRegion),
		CloudStorageConfig:      cloudStorageConfig,
		RabbitMQURL:             RabbitMQHost,
		RabbitMQQueueName:       RabbitMQQueueName,
		YoutubeDLBinPath:        "/not-a-real-path-until-we-need-one",
		YoutubeDLWorkingDirPath: path.Join(local.ProjectRoot(), "/src/worker/wd/youtube-dl"),
		SpleeterBinPath:         envvar.MustGet(envvar.SPLEETER_BIN_PATH),
		SpleeterWorkingDirPath:  path.Join(local.ProjectRoot(), "/src/worker/wd/spleeter"),
		DemucsBinPath:           envvar.MustGet(envvar.DEMUCS_BIN_PATH),
		DemucsWorkingDirPath:    path.Join(local.ProjectRoot(), "/src/worker/wd/demucs"),
	}
}

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
	RabbitMQHost      = dev.RabbitMQHost
	RabbitMQQueueName = "chord-paper-tracks-test"
)

// Server
const (
	ServerPort = ":5010"
)
