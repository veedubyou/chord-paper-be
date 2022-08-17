package song_test

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

func TestSong(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Song Suite")
}

var _ = BeforeSuite(func() {
	testing2.SetTestEnv()
	db = testing2.BeforeSuiteDB("song_integration_test")
})

var _ = AfterSuite(func() {
	testing2.AfterSuiteDB(db)
})
