package trackstorage_test

import (
	dynamolib "github.com/veedubyou/chord-paper-be/server/src/lib/dynamo"
	testlib "github.com/veedubyou/chord-paper-be/server/src/lib/testing"
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
	testlib.SetTestEnv()
	db = testlib.BeforeSuiteDB("track_db_test")
})

var _ = AfterSuite(func() {
	testlib.AfterSuiteDB(db)
})
