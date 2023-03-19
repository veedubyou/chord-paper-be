package application

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
	"github.com/rabbitmq/amqp091-go"
	"github.com/veedubyou/chord-paper-be/src/shared/config"
	dynamolib "github.com/veedubyou/chord-paper-be/src/shared/lib/dynamo"
	"github.com/veedubyou/chord-paper-be/src/shared/lib/rabbitmq"
	trackentity "github.com/veedubyou/chord-paper-be/src/shared/track/entity"
	trackstorage "github.com/veedubyou/chord-paper-be/src/shared/track/storage"
	filestore "github.com/veedubyou/chord-paper-be/src/worker/internal/application/cloud_storage/store"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/application/executor"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/application/jobs/job_router"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/application/jobs/save_stems_to_db"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/application/jobs/split"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/application/jobs/split/splitter"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/application/jobs/split/splitter/file_splitter"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/application/jobs/start"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/application/jobs/transfer"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/application/jobs/transfer/download"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/application/worker"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/lib/cerr"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/lib/storagepath"
	"google.golang.org/api/option"
	"os"
)

func must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}

	return t
}

type App struct {
	worker worker.QueueWorker
}

type Config struct {
	RabbitMQURL        string
	RabbitMQQueueName  string
	DynamoConfig       config.Dynamo
	CloudStorageConfig config.CloudStorage

	YoutubeDLBinPath        string
	YoutubeDLWorkingDirPath string
	SpleeterBinPath         string
	SpleeterWorkingDirPath  string
}

func NewApp(config Config) App {
	consumerConn := must(amqp091.Dial(config.RabbitMQURL))

	return App{
		worker: newWorker(config, consumerConn),
	}
}

func (a *App) Start() error {
	err := a.worker.Start()
	if err != nil {
		return cerr.Wrap(err).Error("Failed to start worker")
	}

	return nil
}

func (a *App) Stop() {
	a.worker.Stop()
}

func newWorker(config Config, consumerConn *amqp091.Connection) worker.QueueWorker {
	publisher := newPublisher(config)

	trackStore := trackstorage.NewDB(newDynamoDB(config.DynamoConfig))
	queueWorker := must(worker.NewQueueWorkerFromConnection(
		consumerConn,
		config.RabbitMQQueueName,
		newJobRouter(config, trackStore, publisher)))

	return queueWorker
}

func newPublisher(config Config) *rabbitmq.QueuePublisher {
	return rabbitmq.NewQueuePublisher(config.RabbitMQURL, config.RabbitMQQueueName)
}

func newDynamoDB(dynamoConfig config.Dynamo) dynamolib.DynamoDBWrapper {
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

	return dynamolib.DynamoDBWrapper{
		DB: dynamo.New(dbSession, dbConfig),
	}
}

func newGoogleFileStore(cloudStorageConfig config.CloudStorage) filestore.GoogleFileStore {
	switch t := cloudStorageConfig.(type) {
	case config.ProdCloudStorage:
		return must(filestore.NewGoogleFileStore(
			t.StorageHost,
			option.WithCredentialsJSON([]byte(t.SecretKey)),
		))

	case config.LocalCloudStorage:
		return must(filestore.NewGoogleFileStore(
			t.StorageHost,
			option.WithEndpoint(t.HostEndpoint),
			option.WithAPIKey("fake_api_key"),
		))

	default:
		panic("Unrecognized cloud storage config")
	}
}

func newJobRouter(config Config, trackStore trackentity.Store, publisher rabbitmq.Publisher) job_router.JobRouter {
	pathGenerator := storagepath.Generator{
		Host:   config.CloudStorageConfig.GetStorageHost(),
		Bucket: config.CloudStorageConfig.GetBucket(),
	}

	return job_router.NewJobRouter(
		trackStore,
		publisher,
		newStartJobHandler(trackStore),
		newDownloadJobHandler(config, pathGenerator),
		newSplitJobHandler(config, pathGenerator),
		newSaveToDBJobHandler(trackStore))
}

func newStartJobHandler(trackStore trackentity.Store) start.JobHandler {
	return start.NewJobHandler(trackStore)
}

func newDownloadJobHandler(config Config, pathGenerator storagepath.Generator) transfer.JobHandler {
	if err := os.MkdirAll(config.YoutubeDLWorkingDirPath, os.ModePerm); err != nil {
		panic(err)
	}

	youtubedler := download.NewYoutubeDLer(config.YoutubeDLBinPath, executor.BinaryFileExecutor{})
	genericdler := download.NewGenericDLer()

	selectdler := download.NewSelectDLer(youtubedler, genericdler)

	trackStore := trackstorage.NewDB(newDynamoDB(config.DynamoConfig))
	trackDownloader := must(transfer.NewTrackTransferrer(
		selectdler,
		trackStore,
		newGoogleFileStore(config.CloudStorageConfig),
		pathGenerator,
		config.YoutubeDLWorkingDirPath,
	))

	return transfer.NewJobHandler(trackDownloader)
}

func newSplitJobHandler(config Config, pathGenerator storagepath.Generator) split.JobHandler {
	if err := os.MkdirAll(config.SpleeterWorkingDirPath, os.ModePerm); err != nil {
		panic(err)
	}

	localUsecase := must(file_splitter.NewLocalFileSplitter(
		config.SpleeterWorkingDirPath,
		config.SpleeterBinPath,
		executor.BinaryFileExecutor{},
	))

	googleFileStore := newGoogleFileStore(config.CloudStorageConfig)
	remoteUsecase := must(file_splitter.NewRemoteFileSplitter(
		config.SpleeterWorkingDirPath,
		googleFileStore,
		localUsecase,
	))

	trackStore := trackstorage.NewDB(newDynamoDB(config.DynamoConfig))

	songSplitUsecase := splitter.NewTrackSplitter(
		remoteUsecase,
		trackStore,
		pathGenerator,
	)

	return split.NewJobHandler(songSplitUsecase)
}

func newSaveToDBJobHandler(trackStore trackentity.Store) save_stems_to_db.JobHandler {
	return save_stems_to_db.NewJobHandler(trackStore)
}
