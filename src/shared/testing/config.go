package testing

import (
	server_app "github.com/veedubyou/chord-paper-be/src/server/application"
	"github.com/veedubyou/chord-paper-be/src/shared/config"
	"github.com/veedubyou/chord-paper-be/src/shared/values/local"
	"github.com/veedubyou/chord-paper-be/src/shared/values/test"
	worker_app "github.com/veedubyou/chord-paper-be/src/worker/application"
	"path"
)

func ServerConfig(dbRegion string) server_app.Config {
	return server_app.Config{
		DynamoConfig:       test.DynamoConfig(dbRegion),
		RabbitMQURL:        test.RabbitMQHost,
		RabbitMQQueueName:  RabbitMQQueueName,
		CORSAllowedOrigins: []string{"*"},
		UserValidator:      Validator{},
		Port:               ServerPort,
		Log:                false,
	}
}

func WorkerConfig(dbRegion string, cloudStorageConfig config.LocalCloudStorage) worker_app.Config {
	return worker_app.Config{
		DynamoConfig:       test.DynamoConfig(dbRegion),
		CloudStorageConfig: cloudStorageConfig,
		RabbitMQURL:        test.RabbitMQHost,
		RabbitMQQueueName:  RabbitMQQueueName,
		//TODO
		YoutubeDLBinPath:        "/home/linuxbrew/.linuxbrew/bin/youtube-dl",
		YoutubeDLWorkingDirPath: path.Join(local.ProjectRoot(), "/src/worker/wd/youtube-dl"),
		//TODO
		SpleeterBinPath:        "/home/linuxbrew/.linuxbrew/bin/spleeter",
		SpleeterWorkingDirPath: path.Join(local.ProjectRoot(), "/src/worker/wd/spleeter"),
	}
}
