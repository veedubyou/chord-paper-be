package song_test

import (
	dynamolib "github.com/veedubyou/chord-paper-be/server/src/internal/lib/dynamo"
	testlib "github.com/veedubyou/chord-paper-be/server/src/internal/lib/testing"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	db dynamolib.DynamoDBWrapper
)

func TestSong(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Song Suite")
}

var _ = BeforeSuite(func() {
	testlib.SetTestEnv()
	db = testlib.BeforeSuiteDB("song_integration_test")
})

var _ = AfterSuite(func() {
	testlib.AfterSuiteDB(db)
})
