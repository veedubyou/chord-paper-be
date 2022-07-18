package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pkg/errors"
	"github.com/veedubyou/chord-paper-be/go-rewrite/src/lib/env"
	middleware2 "github.com/veedubyou/chord-paper-be/go-rewrite/src/lib/middleware"
	songgateway "github.com/veedubyou/chord-paper-be/go-rewrite/src/song/gateway"
	songstorage "github.com/veedubyou/chord-paper-be/go-rewrite/src/song/storage"
	songusecase "github.com/veedubyou/chord-paper-be/go-rewrite/src/song/usecase"
	trackgateway "github.com/veedubyou/chord-paper-be/go-rewrite/src/track/gateway"
	trackstorage "github.com/veedubyou/chord-paper-be/go-rewrite/src/track/storage"
	trackusecase "github.com/veedubyou/chord-paper-be/go-rewrite/src/track/usecase"
	usergateway "github.com/veedubyou/chord-paper-be/go-rewrite/src/user/gateway"
	userusecase "github.com/veedubyou/chord-paper-be/go-rewrite/src/user/usecase"
	"github.com/veedubyou/chord-paper-be/go-rewrite/src/user/userstorage"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
)

type HTTPMethod string

const (
	GET  HTTPMethod = "GET"
	POST HTTPMethod = "POST"
)

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	corsMiddleware := makeCorsMiddleware()

	handleRoute := func(method HTTPMethod, path string, handlerFunc echo.HandlerFunc) {
		switch method {
		case GET:
			e.GET(path, handlerFunc, corsMiddleware, middleware2.ProxyMarkerOn)
		case POST:
			e.POST(path, handlerFunc, corsMiddleware, middleware2.ProxyMarkerOn)
		default:
			panic("unhandled http method!")
		}
	}

	dynamoDB := makeDynamoDB()
	songGateway := makeSongGateway(dynamoDB)
	trackGateway := makeTrackGateway(dynamoDB)
	userGateway := makeUserGateway(dynamoDB)

	handleRoute(GET, "/songs/:id", func(c echo.Context) error {
		songID := c.Param("id")
		return songGateway.GetSong(c, songID)
	})

	handleRoute(GET, "/songs/:id/tracklist", func(c echo.Context) error {
		songID := c.Param("id")
		return trackGateway.GetTrackList(c, songID)
	})

	handleRoute(POST, "/login", userGateway.Login)

	e.Any("/*", proxyHandler, middleware2.ProxyMarkerOff, makeRustProxyMiddleware())

	e.Logger.Fatal(e.Start(":5000"))
}

func makeDynamoDB() *dynamo.DB {
	dbSession := session.Must(session.NewSession())

	config := aws.NewConfig().
		WithCredentials(credentials.NewEnvCredentials())

	switch env.Get() {
	case env.Production:
		config = config.WithRegion("us-east-2")
		return dynamo.New(dbSession, config)

	case env.Development:
		config = config.WithEndpoint("http://localhost:8000").
			WithRegion("localhost")
		return dynamo.New(dbSession, config)

	default:
		panic("unexpected environment")
	}
}

func makeSongGateway(dynamoDB *dynamo.DB) songgateway.Gateway {
	songDB := songstorage.NewDB(dynamoDB)
	songUsecase := songusecase.NewUsecase(songDB)
	return songgateway.NewGateway(songUsecase)
}

func makeTrackGateway(dynamoDB *dynamo.DB) trackgateway.Gateway {
	trackDB := trackstorage.NewDB(dynamoDB)
	trackUsecase := trackusecase.NewUsecase(trackDB)
	return trackgateway.NewGateway(trackUsecase)
}

func makeUserGateway(dynamoDB *dynamo.DB) usergateway.Gateway {
	userDB := userstorage.NewDB(dynamoDB)
	userUsecase := userusecase.NewUsecase(userDB)
	return usergateway.NewGateway(userUsecase)
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
	if err != nil {
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
