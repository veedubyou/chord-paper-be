package env

import "os"

type Environment string

const (
	Production  Environment = "production"
	Development Environment = "development"
)

func Get() Environment {
	environment, ok := os.LookupEnv("ENVIRONMENT")
	if environment == "" || !ok {
		panic("No environment var is set")
	}

	switch environment {
	case "production":
		return Production
	case "development":
		return Development
	default:
		panic("Invalid environment is set")
	}
}
