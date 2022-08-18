package trackstorage_test

import (
	dynamolib "github.com/veedubyou/chord-paper-be/src/shared/lib/dynamo"
	testing2 "github.com/veedubyou/chord-paper-be/src/shared/testing"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	db dynamolib.DynamoDBWrapper
)

func TestStorage(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Storage Suite")
}

var _ = BeforeSuite(func() {
	testing2.SetTestEnv()
	db = testing2.BeforeSuiteDB("track_db_test")
})

var _ = AfterSuite(func() {
	testing2.AfterSuiteDB(db)
})
