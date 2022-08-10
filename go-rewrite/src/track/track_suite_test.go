package track_test

import (
	"github.com/rabbitmq/amqp091-go"
	dynamolib "github.com/veedubyou/chord-paper-be/go-rewrite/src/lib/dynamo"
	testlib "github.com/veedubyou/chord-paper-be/go-rewrite/src/lib/testing"
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
