package track_split_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rabbitmq/amqp091-go"
	dynamolib "github.com/veedubyou/chord-paper-be/src/shared/lib/dynamo"
	. "github.com/veedubyou/chord-paper-be/src/shared/testing"
	"os"
	"testing"
)

const (
	region     = "track_split_integration_test"
	bucketName = "chord-paper-tracks-test"
)

func TestTrackSplit(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TrackSplit Suite")
}

var (
	db               dynamolib.DynamoDBWrapper
	rabbitMQConn     *amqp091.Connection
	githubRepository string
)

var _ = BeforeSuite(func() {
	SetTestEnv()
	db = BeforeSuiteDB(region)
	rabbitMQConn = MakeRabbitMQConnection()

	repo, isSet := os.LookupEnv("GITHUB_REPOSITORY")
	if isSet {
		githubRepository = repo
		os.Unsetenv("GITHUB_REPOSITORY")
	}
})

var _ = AfterSuite(func() {
	AfterSuiteDB(db)
	AfterSuiteRabbitMQ(rabbitMQConn)

	if githubRepository != "" {
		os.Setenv("GITHUB_REPOSITORY", githubRepository)
	}
})
