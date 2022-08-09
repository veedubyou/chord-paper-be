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
	db               dynamolib.DynamoDBWrapper
	rabbitConnection *amqp091.Connection
)

func TestTrack(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Track Suite")
}

var _ = BeforeSuite(func() {
	testlib.SetTestEnv()
	db = testlib.BeforeSuiteDB("track_integration_test")

	rabbitConnection = testlib.MakeRabbitMQConnection()
})

var _ = AfterSuite(func() {
	testlib.AfterSuiteDB(db)
})
