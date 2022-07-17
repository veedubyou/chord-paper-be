package songstorage

import (
	"context"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/errors/markers"
	"github.com/google/uuid"
	"github.com/guregu/dynamo"
	"github.com/veedubyou/chord-paper-be/go-rewrite/src/lib/errors/handle"
	songentity "github.com/veedubyou/chord-paper-be/go-rewrite/src/song/entity"
)

const (
	SongsTable = "Songs"
)

type DB struct {
	dynamoDB *dynamo.DB
}

func NewDB(dynamoDB *dynamo.DB) DB {
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

	return songentity.Song(value), nil
}
