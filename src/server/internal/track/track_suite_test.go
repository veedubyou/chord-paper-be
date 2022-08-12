package track_test

import (
	"github.com/rabbitmq/amqp091-go"
	"github.com/veedubyou/chord-paper-be/src/server/internal/lib/dynamo"
	"github.com/veedubyou/chord-paper-be/src/server/internal/lib/testing"
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
	testlib.SetTestEnv()
	db = testlib.BeforeSuiteDB("track_integration_test")
	publisherConn = testlib.MakeRabbitMQConnection()
	consumerConn = testlib.MakeRabbitMQConnection()
})

var _ = AfterSuite(func() {
	testlib.AfterSuiteDB(db)
	testlib.AfterSuiteRabbitMQ(publisherConn)
})
