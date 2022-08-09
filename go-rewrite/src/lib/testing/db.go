package testlib

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
	. "github.com/onsi/gomega"
	dynamolib "github.com/veedubyou/chord-paper-be/go-rewrite/src/lib/dynamo"
	songstorage "github.com/veedubyou/chord-paper-be/go-rewrite/src/song/storage"
	trackstorage "github.com/veedubyou/chord-paper-be/go-rewrite/src/track/storage"
	userstorage "github.com/veedubyou/chord-paper-be/go-rewrite/src/user/storage"
)

type song struct {
	ID    string `dynamo:"id,hash"`
	Owner string `dynamo:"owner" index:"owner-index,hash"`
}

type tracklist struct {
	SongID string `dynamo:"song_id,hash"`
}

type User struct {
	ID    string `dynamo:"id,hash"`
	Name  string `dynamo:"username"`
	Email string `dynamo:"email"`
}

func MakeTestDB(testRegion string) dynamolib.DynamoDBWrapper {
	dbSession := session.Must(session.NewSession())

	config := aws.NewConfig().
		WithCredentials(credentials.NewStaticCredentials("local", "local", "")).
		WithEndpoint("http://localhost:8000").
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
	err := db.CreateTable(songstorage.SongsTable, song{}).Run()
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	err = db.CreateTable(userstorage.UsersTable, User{}).Run()
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	err = db.CreateTable(trackstorage.TracklistsTable, tracklist{}).Run()
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
