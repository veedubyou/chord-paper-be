package track_split_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rabbitmq/amqp091-go"
	dynamolib "github.com/veedubyou/chord-paper-be/src/shared/lib/dynamo"
	. "github.com/veedubyou/chord-paper-be/src/shared/testing"
	"os"
	"testing"
)

const (
	region            = "track_split_integration_test"
	bucketName        = "chord-paper-tracks-test"
	GITHUB_REPOSITORY = "GITHUB_REPOSITORY"
)

func TestTrackSplit(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TrackSplit Suite")
}

var (
	db               dynamolib.DynamoDBWrapper
	rabbitMQConn     *amqp091.Connection
	githubRepository string
)

// JFC this is some awful stuff
// https://github.com/deezer/spleeter/blob/0d64981fb8e46fdc05d1aca450a2a5c2499b116b/spleeter/model/provider/github.py#L89
// spleeter will look at the GITHUB_REPOSITORY env var as a pointer to where to download the model files
// however Github workflows will also set this same env var to the current repo (i.e. chord-paper-be)
// so then spleeter will try to download the models from my repo
// there's no way to toggle this in spleeter and Github workflows don't allow overriding the GITHUB_REPOSITORY env var
// so the only way is to do it in code
func clearGithubRepoEnvVar() {
	repo, isSet := os.LookupEnv(GITHUB_REPOSITORY)
	if isSet {
		githubRepository = repo
		os.Unsetenv(GITHUB_REPOSITORY)
	}
}

func restoreGithubRepoEnvVar() {
	if githubRepository != "" {
		os.Setenv(GITHUB_REPOSITORY, githubRepository)
	}
}

var _ = BeforeSuite(func() {
	SetTestEnv()
	db = BeforeSuiteDB(region)
	rabbitMQConn = MakeRabbitMQConnection()

	clearGithubRepoEnvVar()
})

var _ = AfterSuite(func() {
	AfterSuiteDB(db)
	AfterSuiteRabbitMQ(rabbitMQConn)

	restoreGithubRepoEnvVar()
})
