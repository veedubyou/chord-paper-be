package envvar

import (
	"fmt"
	"os"
)

const (
	RABBITMQ_URL                     = "RABBITMQ_URL"
	RABBITMQ_QUEUE_NAME              = "RABBITMQ_QUEUE_NAME"
	GOOGLE_CLOUD_KEY                 = "GOOGLE_CLOUD_KEY"
	GOOGLE_CLOUD_STORAGE_BUCKET_NAME = "GOOGLE_CLOUD_STORAGE_BUCKET_NAME"
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
