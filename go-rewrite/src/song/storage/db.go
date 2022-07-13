package songstorage

import (
	"context"
	"github.com/google/uuid"
	"github.com/guregu/dynamo"
	"github.com/pkg/errors"
	z "github.com/veedubyou/chord-paper-be/go-rewrite/src/lib/errors/errlog"
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

	if z.Err(err) {
		return songentity.Song{}, errors.Wrap(err, "Couldn't fetch song")
	}

	return songentity.Song(value), nil
}
