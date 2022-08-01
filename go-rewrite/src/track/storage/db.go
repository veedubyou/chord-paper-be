package trackstorage

import (
	"context"
	"github.com/cockroachdb/errors/markers"
	"github.com/guregu/dynamo"
	"github.com/pkg/errors"
	dynamolib "github.com/veedubyou/chord-paper-be/go-rewrite/src/lib/dynamo"
	"github.com/veedubyou/chord-paper-be/go-rewrite/src/lib/errors/handle"
	trackentity "github.com/veedubyou/chord-paper-be/go-rewrite/src/track/entity"
)

const (
	TracklistsTable = "TrackLists"
)

type DB struct {
	dynamoDB dynamolib.DynamoDBWrapper
}

func NewDB(dynamoDB dynamolib.DynamoDBWrapper) DB {
	return DB{
		dynamoDB: dynamoDB,
	}
}

func (d DB) GetTrackList(ctx context.Context, songID string) (trackentity.TrackList, error) {
	value := dbTrackList{}
	err := d.dynamoDB.Table(TracklistsTable).
		Get(idKey, songID).
		OneWithContext(ctx, &value)

	if err != nil {
		switch {
		case markers.Is(err, TrackListUnmarshalMark):
			return trackentity.TrackList{}, errors.Wrap(err, "Failed to fetch tracklist")
		case errors.Is(err, dynamo.ErrNotFound):
			return trackentity.TrackList{}, handle.Wrap(err, TrackListNotFound, "Tracklist is not found")
		default:
			return trackentity.TrackList{}, handle.Wrap(err, DefaultErrorMark, "Failed to fetch tracklist")
		}
	}

	return trackentity.TrackList(value), nil
}
