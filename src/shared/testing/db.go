package testing

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
	. "github.com/onsi/gomega"
	"github.com/veedubyou/chord-paper-be/src/shared/lib/dynamo"
)

const (
	SongsTable      = "Songs"
	UsersTable      = "Users"
	TrackListsTable = "TrackLists"
)

type song struct {
	ID    string `dynamo:"id,hash"`
	Owner string `dynamo:"owner" index:"owner-index,hash"`
}

type tracklist struct {
	SongID string `dynamo:"song_id,hash"`
}

type User struct {
	ID       string `dynamo:"id,hash"`
	Name     string `dynamo:"username"`
	Email    string `dynamo:"email"`
	Verified bool   `dynamo:"verified"`
}

func MakeTestDB(testRegion string) dynamolib.DynamoDBWrapper {
	dbSession := session.Must(session.NewSession())

	config := aws.NewConfig().
		WithCredentials(credentials.NewStaticCredentials(DynamoAccessKeyID, DynamoSecretAccessKey, "")).
		WithEndpoint(DynamoDBHost).
		WithRegion(testRegion)

	db := dynamo.New(dbSession, config)
	return dynamolib.NewDynamoDBWrapper(db)
}

func ResetDB(db dynamolib.DynamoDBWrapper) {
	DeleteAllTables(db)
	CreateAllTables(db)
	EnsureUsers(db)
}

func BeforeSuiteDB(testRegion string) dynamolib.DynamoDBWrapper {
	db := MakeTestDB(testRegion)
	DeleteAllTables(db)
	return db
}

func AfterSuiteDB(db dynamolib.DynamoDBWrapper) {
	DeleteAllTables(db)
}

func CreateAllTables(db dynamolib.DynamoDBWrapper) {
	err := db.CreateTable(SongsTable, song{}).Run()
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	err = db.CreateTable(UsersTable, User{}).Run()
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	err = db.CreateTable(TrackListsTable, tracklist{}).Run()
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
}

func DeleteAllTables(db dynamolib.DynamoDBWrapper) {
	tableResults := db.ListTables()
	tableNames := ExpectSuccess(tableResults.All())

	for _, tableName := range tableNames {
		err := db.Table(tableName).DeleteTable().Run()
		ExpectWithOffset(1, err).NotTo(HaveOccurred())
	}
}
