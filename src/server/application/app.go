package application

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/cockroachdb/errors"
	"github.com/guregu/dynamo"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rabbitmq/amqp091-go"
	"github.com/veedubyou/chord-paper-be/src/server/google_id"
	"github.com/veedubyou/chord-paper-be/src/server/internal/song/gateway"
	"github.com/veedubyou/chord-paper-be/src/server/internal/song/storage"
	"github.com/veedubyou/chord-paper-be/src/server/internal/song/usecase"
	"github.com/veedubyou/chord-paper-be/src/server/internal/track/gateway"
	"github.com/veedubyou/chord-paper-be/src/server/internal/track/usecase"
	"github.com/veedubyou/chord-paper-be/src/server/internal/user/gateway"
	userstorage "github.com/veedubyou/chord-paper-be/src/server/internal/user/storage"
	"github.com/veedubyou/chord-paper-be/src/server/internal/user/usecase"
	"github.com/veedubyou/chord-paper-be/src/shared/config"
	"github.com/veedubyou/chord-paper-be/src/shared/lib/dynamo"
	"github.com/veedubyou/chord-paper-be/src/shared/lib/rabbitmq"
	"github.com/veedubyou/chord-paper-be/src/shared/track/storage"
	"net/http"
)

type HTTPMethod string

const (
	GET    HTTPMethod = "GET"
	POST   HTTPMethod = "POST"
	PUT    HTTPMethod = "PUT"
	DELETE HTTPMethod = "DELETE"
)

type App struct {
	echo *echo.Echo
	port string
}

type Config struct {
	DynamoConfig       config.Dynamo
	RabbitMQURL        string
	RabbitMQQueueName  string
	CORSAllowedOrigins []string
	UserValidator      google_id.Validator
	Port               string
	Log                bool
}

func NewApp(config Config) App {
	e := echo.New()

	if config.Log {
		e.Use(middleware.Logger())
	}

	corsMiddleware := makeCorsMiddleware(config)

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

	dynamoDB := makeDynamoDB(config.DynamoConfig)
	rabbitmqPublisher := makeRabbitMQPublisher(config)
	userUsecase := makeUserUsecase(config, dynamoDB)
	songUsecase := makeSongUsecase(dynamoDB, userUsecase)

	userGateway := makeUserGateway(userUsecase)
	songGateway := makeSongGateway(songUsecase)
	trackGateway := makeTrackGateway(dynamoDB, songUsecase, rabbitmqPublisher)

	// health check
	handleRoute(GET, "/health-check", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

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

	return App{
		echo: e,
		port: config.Port,
	}
}

func (a *App) Start() error {
	err := a.echo.Start(a.port)
	if err != nil && err != http.ErrServerClosed {
		return errors.Wrap(err, "Couldn't start echo server")
	}

	return nil
}

func (a *App) Stop() error {
	err := a.echo.Close()
	if err != nil {
		return errors.Wrap(err, "Failed to stop echo server")
	}

	return nil
}

func makeRabbitMQPublisher(config Config) rabbitmq.QueuePublisher {
	conn, err := amqp091.Dial(config.RabbitMQURL)
	if err != nil {
		panic(errors.Wrap(err, "Failed to dial rabbitMQ url"))
	}

	publisher, err := rabbitmq.NewQueuePublisher(conn, config.RabbitMQQueueName)
	if err != nil {
		panic(errors.Wrap(err, "Failed to create rabbitMQ publisher"))
	}

	return publisher
}

func makeDynamoDB(dynamoConfig config.Dynamo) dynamolib.DynamoDBWrapper {
	dbSession := session.Must(session.NewSession())

	var dbConfig *aws.Config

	switch t := dynamoConfig.(type) {
	case config.ProdDynamo:
		dbConfig = aws.NewConfig().
			WithCredentials(credentials.NewStaticCredentials(
				t.AccessKeyID,
				t.SecretAccessKey,
				"",
			)).
			WithRegion(t.Region)

	case config.LocalDynamo:
		dbConfig = aws.NewConfig().
			WithCredentials(credentials.NewStaticCredentials(
				t.AccessKeyID,
				t.SecretAccessKey,
				"",
			)).
			WithRegion(t.Region).
			WithEndpoint(t.Host)

	default:
		panic("Unexpected dynamo config type")
	}

	db := dynamo.New(dbSession, dbConfig)
	return dynamolib.NewDynamoDBWrapper(db)
}

func makeSongUsecase(dynamoDB dynamolib.DynamoDBWrapper, userUsecase userusecase.Usecase) songusecase.Usecase {
	songDB := songstorage.NewDB(dynamoDB)
	return songusecase.NewUsecase(songDB, userUsecase)
}

func makeSongGateway(songUsecase songusecase.Usecase) songgateway.Gateway {
	return songgateway.NewGateway(songUsecase)
}

func makeTrackGateway(dynamoDB dynamolib.DynamoDBWrapper, songUsecase songusecase.Usecase, publisher rabbitmq.QueuePublisher) trackgateway.Gateway {
	trackDB := trackstorage.NewDB(dynamoDB)
	trackUsecase := trackusecase.NewUsecase(trackDB, songUsecase, publisher)
	return trackgateway.NewGateway(trackUsecase)
}

func makeUserUsecase(config Config, dynamoDB dynamolib.DynamoDBWrapper) userusecase.Usecase {
	userDB := userstorage.NewDB(dynamoDB)
	return userusecase.NewUsecase(userDB, config.UserValidator)
}

func makeUserGateway(userUsecase userusecase.Usecase) usergateway.Gateway {
	return usergateway.NewGateway(userUsecase)
}

func makeCorsMiddleware(config Config) echo.MiddlewareFunc {
	return middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: config.CORSAllowedOrigins,
		AllowHeaders: []string{echo.HeaderContentType, echo.HeaderAuthorization},
	})
}
