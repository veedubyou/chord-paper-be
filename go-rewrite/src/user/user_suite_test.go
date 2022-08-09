package user_test

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

func TestUser(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "User Suite")
}

var _ = BeforeSuite(func() {
	testlib.SetTestEnv()
	db = testlib.BeforeSuiteDB("user_integration_test")
})

var _ = AfterSuite(func() {
	testlib.AfterSuiteDB(db)
})
