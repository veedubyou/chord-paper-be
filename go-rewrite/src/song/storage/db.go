package songstorage

import (
	"context"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/errors/markers"
	"github.com/google/uuid"
	"github.com/guregu/dynamo"
	dynamolib "github.com/veedubyou/chord-paper-be/go-rewrite/src/lib/dynamo"
	"github.com/veedubyou/chord-paper-be/go-rewrite/src/lib/errors/handle"
	songentity "github.com/veedubyou/chord-paper-be/go-rewrite/src/song/entity"
)

const (
	SongsTable            = "Songs"
	newSongCondition      = "attribute_not_exists(" + idKey + ")"
	existingSongCondition = "attribute_exists(" + idKey + ")"
)

type DB struct {
	dynamoDB dynamolib.DynamoDBWrapper
}

func NewDB(dynamoDB dynamolib.DynamoDBWrapper) DB {
	return DB{
		dynamoDB: dynamoDB,
	}
}

func (d DB) GetSong(ctx context.Context, songID uuid.UUID) (songentity.Song, error) {
	value := dbSong{}
	err := d.dynamoDB.Table(SongsTable).
		Get(idKey, songID.String()).
		OneWithContext(ctx, &value)

	if err != nil {
		switch {
		case markers.Is(err, SongUnmarshalMark):
			return songentity.Song{}, err
		case errors.Is(err, dynamo.ErrNotFound):
			return songentity.Song{}, handle.Wrap(err, SongNotFoundMark, "Song for this ID couldn't be found")
		default:
			return songentity.Song{}, handle.Wrap(err, DefaultErrorMark, "Failed to fetch song due to unknown data store error")
		}
	}

	song := songentity.Song{}
	err = song.FromMap(value)
	if err != nil {
		return songentity.Song{}, handle.Wrap(err, SongUnmarshalMark, "Failed to unmarshal song into its entity form")
	}

	return song, nil
}

func (d DB) CreateSong(ctx context.Context, newSong songentity.Song) (songentity.Song, error) {
	err := d.putSong(ctx, newSong, false)
	if err != nil {
		if conditionalCheckFailed(err) {
			return songentity.Song{}, handle.Wrap(err,
				SongAlreadyExistsMark,
				"Cannot create: A song of this ID already exists")

		}

		return songentity.Song{}, errors.Wrap(err, "Failed to put song into DB")
	}

	return newSong, nil
}

func (d DB) putSong(ctx context.Context, song songentity.Song, expectSongExists bool) error {
	dbObject, err := song.ToMap()
	if err != nil {
		return handle.Wrap(err, SongUnmarshalMark, "Failed to convert song object to a map")
	}

	putExpr := d.dynamoDB.Table(SongsTable).Put(dbObject)

	if expectSongExists {
		putExpr = putExpr.If(existingSongCondition)
	} else {
		putExpr = putExpr.If(newSongCondition)
	}

	return putExpr.RunWithContext(ctx)
}

func (d DB) DeleteSong(ctx context.Context, songID uuid.UUID) error {
	delExpr := d.dynamoDB.Table(SongsTable).Delete(idKey, songID.String())
	delExpr = delExpr.If(existingSongCondition)

	if err := delExpr.RunWithContext(ctx); err != nil {
		if conditionalCheckFailed(err) {
			return handle.Wrap(err, SongNotFoundMark, "Failed to find song to delete")
		}

		return handle.Wrap(err, DefaultErrorMark, "Failed to delete song")
	}

	return nil
}

func conditionalCheckFailed(err error) bool {
	_, ok := err.(*dynamodb.ConditionalCheckFailedException)
	return ok
}
