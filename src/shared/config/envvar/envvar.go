package envvar

import (
	"fmt"
	"os"
)

const (
	AWS_ACCESS_KEY_ID                = "AWS_ACCESS_KEY_ID"
	AWS_SECRET_ACCESS_KEY            = "AWS_SECRET_ACCESS_KEY"
	RABBITMQ_URL                     = "RABBITMQ_URL"
	RABBITMQ_QUEUE_NAME              = "RABBITMQ_QUEUE_NAME"
	GOOGLE_CLOUD_KEY                 = "GOOGLE_CLOUD_KEY"
	GOOGLE_CLOUD_STORAGE_BUCKET_NAME = "GOOGLE_CLOUD_STORAGE_BUCKET_NAME"
	YOUTUBEDL_BIN_PATH               = "YOUTUBEDL_BIN_PATH"
	YOUTUBEDL_WORKING_DIR_PATH       = "YOUTUBEDL_WORKING_DIR_PATH"
	SPLEETER_BIN_PATH                = "SPLEETER_BIN_PATH"
	SPLEETER_WORKING_DIR_PATH        = "SPLEETER_WORKING_DIR_PATH"
)

func MustGet(key string) string {
	val, isSet := os.LookupEnv(key)
	if !isSet {
		panic(fmt.Sprintf("No env variable found for key %s", key))
	}

	if val == "" {
		panic(fmt.Sprintf("Env variable is empty for key %s", key))
	}

	return val
}
