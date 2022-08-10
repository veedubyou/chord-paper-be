package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/cockroachdb/errors"
	"github.com/guregu/dynamo"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rabbitmq/amqp091-go"
	dynamolib "github.com/veedubyou/chord-paper-be/server/src/lib/dynamo"
	"github.com/veedubyou/chord-paper-be/server/src/lib/env"
	"github.com/veedubyou/chord-paper-be/server/src/lib/rabbitmq"
	songgateway "github.com/veedubyou/chord-paper-be/server/src/song/gateway"
	songstorage "github.com/veedubyou/chord-paper-be/server/src/song/storage"
	songusecase "github.com/veedubyou/chord-paper-be/server/src/song/usecase"
	trackgateway "github.com/veedubyou/chord-paper-be/server/src/track/gateway"
	trackstorage "github.com/veedubyou/chord-paper-be/server/src/track/storage"
	trackusecase "github.com/veedubyou/chord-paper-be/server/src/track/usecase"
	usergateway "github.com/veedubyou/chord-paper-be/server/src/user/gateway"
	"github.com/veedubyou/chord-paper-be/server/src/user/google_id"
	"github.com/veedubyou/chord-paper-be/server/src/user/storage"
	userusecase "github.com/veedubyou/chord-paper-be/server/src/user/usecase"
	"os"
	"strings"
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
		params := func() (string, echo.HandlerFunc, echo.MiddlewareFunc) {
			return path, handlerFunc, corsMiddleware
		}

		e.OPTIONS(params())

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
	rabbitmqPublisher := makeRabbitMQPublisherForEnv()
	userUsecase := makeUserUsecase(dynamoDB)
	songUsecase := makeSongUsecase(dynamoDB, userUsecase)

	userGateway := makeUserGateway(userUsecase)
	songGateway := makeSongGateway(songUsecase)
	trackGateway := makeTrackGateway(dynamoDB, songUsecase, rabbitmqPublisher)

	// login route
	handleRoute(POST, "/login", userGateway.Login)

	// song routes
	handleRoute(POST, "/songs", songGateway.CreateSong)
	handleRoute(GET, "/songs/:id", func(c echo.Context) error {
		songID := c.Param("id")
		return songGateway.GetSong(c, songID)
	})
	handleRoute(PUT, "/songs/:id", func(c echo.Context) error {
		songID := c.Param("id")
		return songGateway.UpdateSong(c, songID)
	})
	handleRoute(DELETE, "/songs/:id", func(c echo.Context) error {
		songID := c.Param("id")
		return songGateway.DeleteSong(c, songID)
	})
	handleRoute(GET, "/users/:id/songs", func(c echo.Context) error {
		userID := c.Param("id")
		return songGateway.GetSongSummariesForUser(c, userID)
	})

	// tracklist routes
	handleRoute(GET, "/songs/:id/tracklist", func(c echo.Context) error {
		songID := c.Param("id")
		return trackGateway.GetTrackList(c, songID)
	})
	handleRoute(PUT, "/songs/:id/tracklist", func(c echo.Context) error {
		songID := c.Param("id")
		return trackGateway.SetTrackList(c, songID)
	})

	e.Logger.Fatal(e.Start(":5000"))
}

func makeRabbitMQPublisherForEnv() rabbitmq.Publisher {
	switch env.Get() {
	case env.Production:
		queueName, isSet := os.LookupEnv("RABBITMQ_QUEUE")
		if !isSet {
			panic("RABBITMQ_QUEUE is not set")
		}

		hostURL, isSet := os.LookupEnv("RABBITMQ_URL")
		if !isSet {
			panic("RABBITMQ_URL is not set")
		}

		return makeRabbitMQPublisher(hostURL, queueName)

	case env.Development:
		return makeRabbitMQPublisher("amqp://localhost:5672", "chord-paper-tracks-dev")

	default:
		panic("unexpected environment")
	}
}

func makeRabbitMQPublisher(hostURL string, queueName string) rabbitmq.Publisher {
	conn, err := amqp091.Dial(hostURL)
	if err != nil {
		panic(errors.Wrap(err, "Failed to dial rabbitMQ url"))
	}

	publisher, err := rabbitmq.NewPublisher(conn, queueName)
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

func makeSongUsecase(dynamoDB dynamolib.DynamoDBWrapper, userUsecase userusecase.Usecase) songusecase.Usecase {
	songDB := songstorage.NewDB(dynamoDB)
	return songusecase.NewUsecase(songDB, userUsecase)
}

func makeSongGateway(songUsecase songusecase.Usecase) songgateway.Gateway {
	return songgateway.NewGateway(songUsecase)
}

func makeTrackGateway(dynamoDB dynamolib.DynamoDBWrapper, songUsecase songusecase.Usecase, publisher rabbitmq.Publisher) trackgateway.Gateway {
	trackDB := trackstorage.NewDB(dynamoDB)
	trackUsecase := trackusecase.NewUsecase(trackDB, songUsecase, publisher)
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
		allowedOrigins = []string{"*"}
	default:
		panic("Unexpected environment")
	}

	return middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: allowedOrigins,
		AllowHeaders: []string{echo.HeaderContentType, echo.HeaderAuthorization},
	})
}
