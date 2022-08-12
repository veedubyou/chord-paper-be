package env

import (
	"github.com/veedubyou/chord-paper-be/shared/values/envvars"
)

type Environment string

const (
	Production  Environment = "production"
	Development Environment = "development"
	Test        Environment = "test"
)

func Get() Environment {
	environment := envvars.MustGet("ENVIRONMENT")

	switch environment {
	case "production":
		return Production
	case "development":
		return Development
	case "test":
		return Test
	default:
		panic("Invalid environment is set")
	}
}
