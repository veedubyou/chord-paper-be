package song_test

import (
	dynamolib "github.com/veedubyou/chord-paper-be/go-rewrite/src/lib/dynamo"
	testlib "github.com/veedubyou/chord-paper-be/go-rewrite/src/lib/testing"
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
	db = testlib.BeforeSuiteTestDB("song_integration_test")
})

var _ = AfterSuite(func() {
	testlib.AfterSuiteTestDB(db)
})
