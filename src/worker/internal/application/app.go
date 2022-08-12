package application

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/rabbitmq/amqp091-go"
	"github.com/veedubyou/chord-paper-be/src/shared/lib/env"
	"github.com/veedubyou/chord-paper-be/src/shared/values/envvar"
	"github.com/veedubyou/chord-paper-be/src/shared/values/local"
	"github.com/veedubyou/chord-paper-be/src/shared/values/region"
	filestore "github.com/veedubyou/chord-paper-be/src/worker/internal/application/cloud_storage/store"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/application/executor"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/application/jobs/job_router"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/application/jobs/save_stems_to_db"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/application/jobs/split"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/application/jobs/split/splitter"
	file_splitter2 "github.com/veedubyou/chord-paper-be/src/worker/internal/application/jobs/split/splitter/file_splitter"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/application/jobs/start"
	transfer2 "github.com/veedubyou/chord-paper-be/src/worker/internal/application/jobs/transfer"
	download2 "github.com/veedubyou/chord-paper-be/src/worker/internal/application/jobs/transfer/download"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/application/publish"
	trackstore "github.com/veedubyou/chord-paper-be/src/worker/internal/application/tracks/store"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/application/worker"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/lib/cerr"
	"os"
)

func ensureOk(err error) {
	if err != nil {
		panic(err)
	}
}

type App struct {
	worker worker.QueueWorker
}

func NewApp() App {
	rabbitMQURL := rabbitURL()
	consumerConn, err := amqp091.Dial(rabbitMQURL)
	ensureOk(err)
	producerConn, err := amqp091.Dial(rabbitMQURL)
	ensureOk(err)

	return App{
		worker: newWorker(consumerConn, producerConn),
	}
}

func (a *App) Start() error {
	err := a.worker.Start()
	if err != nil {
		return cerr.Wrap(err).Error("Failed to start worker")
	}

	return nil
}

func newWorker(consumerConn *amqp091.Connection, producerConn *amqp091.Connection) worker.QueueWorker {
	publisher := newPublisher(producerConn)

	trackStore := trackstore.NewDynamoDBTrackStore(newDynamoDB())
	queueWorker, err := worker.NewQueueWorkerFromConnection(
		consumerConn,
		queueName(),
		newJobRouter(trackStore, publisher))
	ensureOk(err)
	return queueWorker
}

func rabbitURL() string {
	switch env.Get() {
	case env.Production:
		return envvar.MustGet(envvar.RABBITMQ_URL)
	case env.Development:
		return local.RabbitMQHost

	default:
		panic("Unrecognized environment")
	}
}

func queueName() string {
	switch env.Get() {
	case env.Production:
		return envvar.MustGet(envvar.RABBITMQ_QUEUE_NAME)
	case env.Development:
		return local.RabbitMQQueueName

	default:
		panic("Unrecognized environment")
	}

}

func newPublisher(conn *amqp091.Connection) publish.RabbitMQPublisher {
	publisher, err := publish.NewRabbitMQPublisher(conn, queueName())
	ensureOk(err)
	return publisher
}

func newDynamoDB() *dynamodb.DynamoDB {
	dbSession := session.Must(session.NewSession())

	config := aws.NewConfig().WithCredentials(credentials.NewEnvCredentials())

	switch env.Get() {
	case env.Production:
		config = config.WithRegion(region.Prod)

	case env.Development:
		config = config.WithEndpoint(local.DynamoDBHost).
			WithRegion(local.DynamoDBRegion)

	default:
		panic("Unrecognized environment")
	}

	return dynamodb.New(dbSession, config)
}

func newGoogleFileStore() filestore.GoogleFileStore {
	jsonKey := envvar.MustGet(envvar.GOOGLE_CLOUD_KEY)

	fileStore, err := filestore.NewGoogleFileStore(jsonKey)
	ensureOk(err)
	return fileStore
}

func newJobRouter(trackStore trackstore.DynamoDBTrackStore, publisher publish.Publisher) job_router.JobRouter {
	return job_router.NewJobRouter(
		trackStore,
		publisher,
		newStartJobHandler(trackStore),
		newDownloadJobHandler(),
		newSplitJobHandler(),
		newSaveToDBJobHandler(trackStore))
}

func newStartJobHandler(trackStore trackstore.DynamoDBTrackStore) start.JobHandler {
	return start.NewJobHandler(trackStore)
}

func newDownloadJobHandler() transfer2.JobHandler {
	youtubeDLBinPath := envvar.MustGet("YOUTUBEDL_BIN_PATH")
	workingDir := envvar.MustGet("YOUTUBEDL_WORKING_DIR_PATH")
	err := os.MkdirAll(workingDir, os.ModePerm)
	ensureOk(err)

	youtubedler := download2.NewYoutubeDLer(youtubeDLBinPath, executor.BinaryFileExecutor{})
	genericdler := download2.NewGenericDLer()

	selectdler := download2.NewSelectDLer(youtubedler, genericdler)

	trackStore := trackstore.NewDynamoDBTrackStore(newDynamoDB())
	bucketName := envvar.MustGet(envvar.GOOGLE_CLOUD_STORAGE_BUCKET_NAME)
	trackDownloader, err := transfer2.NewTrackTransferrer(selectdler, trackStore, newGoogleFileStore(), bucketName, workingDir)
	ensureOk(err)

	return transfer2.NewJobHandler(trackDownloader)
}

func newSplitJobHandler() split.JobHandler {
	workingDir := envvar.MustGet("SPLEETER_WORKING_DIR_PATH")
	spleeterBinPath := envvar.MustGet("SPLEETER_BIN_PATH")
	err := os.MkdirAll(workingDir, os.ModePerm)
	ensureOk(err)

	localUsecase, err := file_splitter2.NewLocalFileSplitter(workingDir, spleeterBinPath, executor.BinaryFileExecutor{})
	ensureOk(err)

	googleFileStore := newGoogleFileStore()
	remoteUsecase, err := file_splitter2.NewRemoteFileSplitter(workingDir, googleFileStore, localUsecase)
	ensureOk(err)

	trackStore := trackstore.NewDynamoDBTrackStore(newDynamoDB())
	songSplitUsecase := splitter.NewTrackSplitter(remoteUsecase, trackStore, "chord-paper-tracks")

	return split.NewJobHandler(songSplitUsecase)
}

func newSaveToDBJobHandler(trackStore trackstore.DynamoDBTrackStore) save_stems_to_db.JobHandler {
	return save_stems_to_db.NewJobHandler(trackStore)
}
