package trackstorage

import (
	"context"
	"github.com/google/uuid"
	"github.com/guregu/dynamo"
	"github.com/pkg/errors"
	trackentity "github.com/veedubyou/chord-paper-be/go-rewrite/src/track/entity"
)

const (
	TracklistsTable = "TrackLists"
)

type DB struct {
	dynamoDB *dynamo.DB
}

func NewDB(dynamoDB *dynamo.DB) DB {
	return DB{
		dynamoDB: dynamoDB,
	}
}

func (d DB) GetTrackList(ctx context.Context, songID uuid.UUID) (trackentity.TrackList, error) {
	value := dbTrackList{}
	err := d.dynamoDB.Table(TracklistsTable).
		Get(idKey, songID.String()).
		OneWithContext(ctx, &value)

	if err != nil {
		if errors.Is(err, dynamo.ErrNotFound) {
			return trackentity.NewTrackList(songID.String()), nil
		}

		return trackentity.TrackList{}, errors.Wrap(err, "Couldn't fetch tracklist")
	}

	return trackentity.TrackList(value), nil
}
