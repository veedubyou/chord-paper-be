package track_test

import (
	"github.com/rabbitmq/amqp091-go"
	dynamolib "github.com/veedubyou/chord-paper-be/src/shared/lib/dynamo"
	. "github.com/veedubyou/chord-paper-be/src/shared/testing"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	db            dynamolib.DynamoDBWrapper
	publisherConn *amqp091.Connection
	consumerConn  *amqp091.Connection
)

func TestTrack(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Track Suite")
}

var _ = BeforeSuite(func() {
	SetTestEnv()
	db = BeforeSuiteDB("track_integration_test")
	publisherConn = MakeRabbitMQConnection()
	consumerConn = MakeRabbitMQConnection()
})

var _ = AfterSuite(func() {
	AfterSuiteDB(db)
	AfterSuiteRabbitMQ(publisherConn)
})
