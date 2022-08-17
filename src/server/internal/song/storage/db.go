package songstorage

import (
	"context"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/errors/markers"
	"github.com/guregu/dynamo"
	"github.com/veedubyou/chord-paper-be/src/server/internal/song/entity"
	"github.com/veedubyou/chord-paper-be/src/shared/lib/dynamo"
	"github.com/veedubyou/chord-paper-be/src/shared/lib/errors/mark"
)

const (
	SongsTable            = "Songs"
	newSongCondition      = "attribute_not_exists(" + idKey + ")"
	existingSongCondition = "attribute_exists(" + idKey + ")"
	lastSavedAtField      = "lastSavedAt"
	metadataField         = "metadata"
	ownerIndex            = "owner-index"
)

type DB struct {
	dynamoDB dynamolib.DynamoDBWrapper
}

func NewDB(dynamoDB dynamolib.DynamoDBWrapper) DB {
	return DB{
		dynamoDB: dynamoDB,
	}
}

func (d DB) GetSong(ctx context.Context, songID string) (songentity.Song, error) {
	if songID == "" {
		err := errors.New("Song ID is empty")
		return songentity.Song{}, mark.Wrap(err, SongNotFoundMark, "No ID provided to fetch song")
	}

	value := dbSong{}
	err := d.dynamoDB.Table(SongsTable).
		Get(idKey, songID).
		OneWithContext(ctx, &value)

	if err != nil {
		switch {
		case markers.Is(err, SongUnmarshalMark):
			return songentity.Song{}, err
		case errors.Is(err, dynamo.ErrNotFound):
			return songentity.Song{}, mark.Wrap(err, SongNotFoundMark, "Song for this ID couldn't be found")
		default:
			return songentity.Song{}, mark.Wrap(err, DefaultErrorMark, "Failed to fetch song due to unknown data store error")
		}
	}

	song := songentity.Song{}
	err = song.FromMap(value)
	if err != nil {
		return songentity.Song{}, mark.Wrap(err, SongUnmarshalMark, "Failed to unmarshal song into its entity form")
	}

	return song, nil
}

func (d DB) GetSongSummariesForUser(ctx context.Context, ownerID string) ([]songentity.SongSummary, error) {
	values := []dbSong{}
	err := d.dynamoDB.Table(SongsTable).
		Get(ownerKey, ownerID).
		Index(ownerIndex).
		Project(idKey, ownerKey, lastSavedAtField, metadataField).
		AllWithContext(ctx, &values)

	if err != nil {
		return nil, mark.Wrap(err,
			DefaultErrorMark,
			"Failed to fetch all songs for owner ID")
	}

	summaries := []songentity.SongSummary{}
	for _, value := range values {
		summary := songentity.SongSummary{}
		err := summary.FromMap(value)
		if err != nil {
			return nil, mark.Wrap(err,
				SongUnmarshalMark,
				"Failed to unmarshal song into its entity form")
		}

		summaries = append(summaries, summary)
	}

	return summaries, nil
}

func (d DB) CreateSong(ctx context.Context, newSong songentity.Song) error {
	if newSong.Defined.ID == "" {
		err := errors.New("Song ID is empty")
		return mark.Wrap(err, DefaultErrorMark, "No ID provided to create song")
	}

	err := d.putSong(ctx, newSong, false)
	if err != nil {
		if conditionalCheckFailed(err) {
			return mark.Wrap(err,
				SongAlreadyExistsMark,
				"Cannot create: A song of this ID already exists")

		}

		return errors.Wrap(err, "Failed to put song into DB")
	}

	return nil
}

func (d DB) UpdateSong(ctx context.Context, song songentity.Song) error {
	if song.Defined.ID == "" {
		err := errors.New("Song ID is empty")
		return mark.Wrap(err, SongNotFoundMark, "No ID provided to update song")
	}

	err := d.putSong(ctx, song, true)
	if err != nil {
		if conditionalCheckFailed(err) {
			return mark.Wrap(err,
				SongNotFoundMark,
				"Cannot update: Song of this ID cannot be found")

		}

		return errors.Wrap(err, "Failed to put song into DB")
	}

	return nil
}

func (d DB) putSong(ctx context.Context, song songentity.Song, expectSongExists bool) error {
	dbObject, err := song.ToMap()
	if err != nil {
		return mark.Wrap(err, SongUnmarshalMark, "Failed to convert song object to a map")
	}

	putExpr := d.dynamoDB.Table(SongsTable).Put(dbObject)

	if expectSongExists {
		putExpr = putExpr.If(existingSongCondition)
	} else {
		putExpr = putExpr.If(newSongCondition)
	}

	return putExpr.RunWithContext(ctx)
}

func (d DB) DeleteSong(ctx context.Context, songID string) error {
	if songID == "" {
		err := errors.New("Song ID is empty")
		return mark.Wrap(err, SongNotFoundMark, "No ID provided to delete song")
	}

	delExpr := d.dynamoDB.Table(SongsTable).Delete(idKey, songID)
	delExpr = delExpr.If(existingSongCondition)

	if err := delExpr.RunWithContext(ctx); err != nil {
		if conditionalCheckFailed(err) {
			return mark.Wrap(err, SongNotFoundMark, "Failed to find song to delete")
		}

		return mark.Wrap(err, DefaultErrorMark, "Failed to delete song")
	}

	return nil
}

func conditionalCheckFailed(err error) bool {
	_, ok := err.(*dynamodb.ConditionalCheckFailedException)
	return ok
}
