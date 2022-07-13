package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pkg/errors"
	"github.com/veedubyou/chord-paper-be/go-rewrite/src/lib/env"
	z "github.com/veedubyou/chord-paper-be/go-rewrite/src/lib/errors/errlog"
	middleware2 "github.com/veedubyou/chord-paper-be/go-rewrite/src/lib/middleware"
	songateway "github.com/veedubyou/chord-paper-be/go-rewrite/src/song/gateway"
	songstorage "github.com/veedubyou/chord-paper-be/go-rewrite/src/song/storage"
	songusecase "github.com/veedubyou/chord-paper-be/go-rewrite/src/song/usecase"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()
	e.Use(middleware.Logger())

	songGateway := makeSongGateway(makeDynamoDB())
	e.GET("/songs/:id", func(c echo.Context) error {
		songID := c.Param("id")
		return songGateway.GetSong(c, songID)
	}, makeCorsMiddleware(), middleware2.ProxyMarkerOn)

	proxyMiddleware := makeRustProxyMiddleware()

	// specifically making calling this route out so that it's not caught by the get songs router
	e.GET("/songs/:id/tracklist", proxyHandler, middleware2.ProxyMarkerOff, proxyMiddleware)
	e.Any("/*", proxyHandler, middleware2.ProxyMarkerOff, proxyMiddleware)

	e.Logger.Fatal(e.Start(":5000"))
}

func makeDynamoDB() *dynamo.DB {
	dbSession := session.Must(session.NewSession())

	config := aws.NewConfig().
		WithRegion("us-east-2").
		WithCredentials(credentials.NewEnvCredentials())

	switch env.Get() {
	case env.Production:
		return dynamo.New(dbSession, config)

	case env.Development:
		config = config.WithEndpoint("http://localhost:8000")
		return dynamo.New(dbSession, config)

	default:
		panic("unexpected environment")
	}
}

func makeSongGateway(dynamoDB *dynamo.DB) songateway.Gateway {
	songDB := songstorage.NewDB(dynamoDB)
	songUsecase := songusecase.NewUsecase(songDB)
	return songateway.NewGateway(songUsecase)
}

func proxyHandler(c echo.Context) error {
	return c.String(http.StatusInternalServerError, "Proxy handler, this should never be seen")
}

func makeRustProxyMiddleware() echo.MiddlewareFunc {
	legacyHostStr, isSet := os.LookupEnv("LEGACY_BE_HOST")
	if !isSet {
		panic("Legacy backend host env var is not set")
	}

	legacyHostURL, err := url.Parse(legacyHostStr)
	if z.Err(err) {
		err = errors.Wrap(err, "Failed to parse legacy host URL")
		panic(err)
	}

	balancer := middleware.NewRandomBalancer([]*middleware.ProxyTarget{
		{
			Name: "Rust backend",
			URL:  legacyHostURL,
		},
	})

	return middleware.ProxyWithConfig(middleware.ProxyConfig{
		Balancer: balancer,
	})
}

func makeCorsMiddleware() echo.MiddlewareFunc {
	allowedOrigins := []string{}

	switch env.Get() {
	case env.Production:
		commaSeparatedOrigins, ok := os.LookupEnv("ALLOWED_FE_ORIGINS")
		if !ok {
			panic("ALLOWED_FE_ORIGINS not set")
		}

		allowedOrigins = strings.Split(commaSeparatedOrigins, ",")
	case env.Development:
		allowedOrigins = []string{"http://localhost:3000"}
	default:
		panic("Unexpected environment")
	}

	return middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: allowedOrigins,
		AllowHeaders: []string{echo.HeaderContentType, echo.HeaderAuthorization},
	})
}
