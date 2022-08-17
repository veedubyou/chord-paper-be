package track_split_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rabbitmq/amqp091-go"
	dynamolib "github.com/veedubyou/chord-paper-be/src/shared/lib/dynamo"
	. "github.com/veedubyou/chord-paper-be/src/shared/testing"
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
	db           dynamolib.DynamoDBWrapper
	rabbitMQConn *amqp091.Connection
)

var _ = BeforeSuite(func() {
	SetTestEnv()
	db = BeforeSuiteDB(region)
	rabbitMQConn = MakeRabbitMQConnection()
})

var _ = AfterSuite(func() {
	AfterSuiteDB(db)
	AfterSuiteRabbitMQ(rabbitMQConn)
})
