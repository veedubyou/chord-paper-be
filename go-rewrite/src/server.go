package main

import (
	"github.com/labstack/echo/v4/middleware"
	"net/url"
	"os"

	"github.com/labstack/echo/v4"
)

func main() {
	legacyHostStr, isSet := os.LookupEnv("LEGACY_BE_HOST")
	if !isSet {
		panic("Legacy backend host env var is not set")
	}

	runningEnvironment, isSet := os.LookupEnv("ENVIRONMENT")
	if !isSet {
		panic("Environment is not set")
	}

	portAddr := ""

	switch runningEnvironment {
	case "production":
		portAddr = ":5000"
	case "development":
		portAddr = ":5001"
	default:
		panicMsg := "Unrecognized environment: " + runningEnvironment
		panic(panicMsg)
	}

	legacyHostURL, err := url.Parse(legacyHostStr)
	if err != nil {
		panic(err)
	}

	e := echo.New()

	balancer := middleware.NewRandomBalancer([]*middleware.ProxyTarget{
		{
			Name: "Rust backend",
			URL:  legacyHostURL,
		},
	})

	e.Use(middleware.Logger())

	e.Use(middleware.ProxyWithConfig(middleware.ProxyConfig{
		Balancer: balancer,
		Skipper: func(c echo.Context) bool {
			return false
		},
	}))

	e.Logger.Fatal(e.Start(portAddr))
}
