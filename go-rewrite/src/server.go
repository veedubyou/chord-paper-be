package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pkg/errors"
	"github.com/rabbitmq/amqp091-go"
	dynamolib "github.com/veedubyou/chord-paper-be/go-rewrite/src/lib/dynamo"
	"github.com/veedubyou/chord-paper-be/go-rewrite/src/lib/env"
	middleware2 "github.com/veedubyou/chord-paper-be/go-rewrite/src/lib/middleware"
	"github.com/veedubyou/chord-paper-be/go-rewrite/src/lib/rabbitmq"
	songgateway "github.com/veedubyou/chord-paper-be/go-rewrite/src/song/gateway"
	songstorage "github.com/veedubyou/chord-paper-be/go-rewrite/src/song/storage"
	songusecase "github.com/veedubyou/chord-paper-be/go-rewrite/src/song/usecase"
	trackgateway "github.com/veedubyou/chord-paper-be/go-rewrite/src/track/gateway"
	trackstorage "github.com/veedubyou/chord-paper-be/go-rewrite/src/track/storage"
	trackusecase "github.com/veedubyou/chord-paper-be/go-rewrite/src/track/usecase"
	usergateway "github.com/veedubyou/chord-paper-be/go-rewrite/src/user/gateway"
	"github.com/veedubyou/chord-paper-be/go-rewrite/src/user/google_id"
	"github.com/veedubyou/chord-paper-be/go-rewrite/src/user/storage"
	userusecase "github.com/veedubyou/chord-paper-be/go-rewrite/src/user/usecase"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
)

type HTTPMethod string

const (
	GET    HTTPMethod = "GET"
	POST   HTTPMethod = "POST"
	PUT    HTTPMethod = "PUT"
	DELETE HTTPMethod = "DELETE"

	// should get injected as an env var, but YAGNI for now as it's not a secret
	// and there's no other case to reflect this need
	googleClientID = "650853277550-ta69qbfcvdl6tb5ogtnh2d07ae9rcdlf.apps.googleusercontent.com"
)

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	corsMiddleware := makeCorsMiddleware()

	handleRoute := func(method HTTPMethod, path string, handlerFunc echo.HandlerFunc) {
		params := func() (string, echo.HandlerFunc, echo.MiddlewareFunc, echo.MiddlewareFunc) {
			return path, handlerFunc, corsMiddleware, middleware2.ProxyMarkerOn
		}

		switch method {
		case GET:
			e.GET(params())
		case POST:
			e.POST(params())
		case PUT:
			e.PUT(params())
		case DELETE:
			e.DELETE(params())
		default:
			panic("unhandled http method!")
		}
	}

	dynamoDB := makeDynamoDB()
	userUsecase := makeUserUsecase(dynamoDB)
	songGateway := makeSongGateway(dynamoDB, userUsecase)
	trackGateway := makeTrackGateway(dynamoDB)
	userGateway := makeUserGateway(userUsecase)

	handleRoute(POST, "/login", userGateway.Login)

	handleRoute(GET, "/songs/:id", func(c echo.Context) error {
		songID := c.Param("id")
		return songGateway.GetSong(c, songID)
	})

	handleRoute(POST, "/songs", songGateway.CreateSong)

	handleRoute(PUT, "/songs/:id", func(c echo.Context) error {
		songID := c.Param("id")
		return songGateway.UpdateSong(c, songID)
	})

	handleRoute(DELETE, "/songs/:id", func(c echo.Context) error {
		songID := c.Param("id")
		return songGateway.DeleteSong(c, songID)
	})

	handleRoute(GET, "/songs/:id/tracklist", func(c echo.Context) error {
		songID := c.Param("id")
		return trackGateway.GetTrackList(c, songID)
	})

	handleRoute(GET, "/users/:id/songs", func(c echo.Context) error {
		userID := c.Param("id")
		return songGateway.GetSongSummariesForUser(c, userID)
	})

	e.Any("/*", proxyHandler, middleware2.ProxyMarkerOff, makeRustProxyMiddleware())

	e.Logger.Fatal(e.Start(":5000"))
}

func makeRabbitMQPublisher() rabbitmq.Publisher {
	//TODO env var it
	conn, err := amqp091.Dial("localhost:5672")
	if err != nil {
		panic(errors.Wrap(err, "Failed to dial rabbitMQ url"))
	}

	//TODO env var it
	publisher, err := rabbitmq.NewPublisher(conn, "chord-paper-tracks-dev")
	if err != nil {
		panic(errors.Wrap(err, "Failed to create rabbitMQ publisher"))
	}

	return publisher
}

func makeDynamoDB() dynamolib.DynamoDBWrapper {
	dbSession := session.Must(session.NewSession())

	config := aws.NewConfig().
		WithCredentials(credentials.NewEnvCredentials())

	switch env.Get() {
	case env.Production:
		config = config.WithRegion("us-east-2")

	case env.Development:
		config = config.WithEndpoint("http://localhost:8000").
			WithRegion("localhost")

	default:
		panic("unexpected environment")
	}

	db := dynamo.New(dbSession, config)
	return dynamolib.NewDynamoDBWrapper(db)
}

func makeSongGateway(dynamoDB dynamolib.DynamoDBWrapper, userUsecase userusecase.Usecase) songgateway.Gateway {
	songDB := songstorage.NewDB(dynamoDB)
	songUsecase := songusecase.NewUsecase(songDB, userUsecase)
	return songgateway.NewGateway(songUsecase)
}

func makeTrackGateway(dynamoDB dynamolib.DynamoDBWrapper) trackgateway.Gateway {
	trackDB := trackstorage.NewDB(dynamoDB)
	trackUsecase := trackusecase.NewUsecase(trackDB)
	return trackgateway.NewGateway(trackUsecase)
}

func makeUserUsecase(dynamoDB dynamolib.DynamoDBWrapper) userusecase.Usecase {
	userDB := userstorage.NewDB(dynamoDB)
	validator := google_id.GoogleValidator{ClientID: googleClientID}
	return userusecase.NewUsecase(userDB, validator)
}

func makeUserGateway(userUsecase userusecase.Usecase) usergateway.Gateway {
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
