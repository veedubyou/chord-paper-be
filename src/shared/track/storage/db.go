package trackstorage

import (
	"context"
	"fmt"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/errors/markers"
	"github.com/guregu/dynamo"
	"github.com/veedubyou/chord-paper-be/src/shared/lib/dynamo"
	"github.com/veedubyou/chord-paper-be/src/shared/lib/errors/mark"
	"github.com/veedubyou/chord-paper-be/src/shared/track/entity"
)

const (
	TracklistsTable = "TrackLists"
	maxTrackSize    = 10
)

var _ trackentity.Store = DB{}

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
		case markers.Is(err, UnmarshalMark):
			return trackentity.TrackList{}, errors.Wrap(err, "Failed to fetch tracklist")
		case errors.Is(err, dynamo.ErrNotFound):
			return trackentity.TrackList{}, mark.Wrap(err, TrackListNotFound, "Tracklist is not found")
		default:
			return trackentity.TrackList{}, mark.Wrap(err, DefaultErrorMark, "Failed to fetch tracklist")
		}
	}

	tracklist := trackentity.TrackList{}
	err = tracklist.FromMap(value)
	if err != nil {
		return trackentity.TrackList{},
			mark.Wrap(err, UnmarshalMark, "Failed to transform DB map back to entity tracklist")
	}

	return tracklist, nil
}

func (d DB) SetTrackList(ctx context.Context, tracklist trackentity.TrackList) error {
	if tracklist.Defined.SongID == "" {
		return mark.Message(IDEmptyMark, "Song ID is not defined on tracklist")
	}

	if len(tracklist.Defined.Tracks) > maxTrackSize {
		return mark.Message(TrackSizeExceeded, "The tracklist has more tracks than allowed")
	}

	for _, track := range tracklist.Defined.Tracks {
		if track.GetID() == "" {
			return mark.Message(IDEmptyMark, "A track in the tracklist has an empty ID")
		}
	}

	dbObject, err := tracklist.ToMap()
	if err != nil {
		return mark.Wrap(err,
			MarshalMark,
			"Failed to transform entity tracklist to a generic map object")
	}

	err = d.dynamoDB.Table(TracklistsTable).Put(dbObject).RunWithContext(ctx)
	if err != nil {
		return mark.Wrap(err,
			DefaultErrorMark,
			"Failed to put the tracklist in the DB")
	}

	return nil
}

func (d DB) UpdateTrack(ctx context.Context, tracklistID string, trackID string, updater trackentity.TrackUpdater) error {
	if trackID == "" {
		return mark.Message(TrackNotFound, "No track ID was provided")
	}

	tracklist, err := d.GetTrackList(ctx, tracklistID)
	if err != nil {
		return mark.Wrap(err, TrackListNotFound, "Can't find the tracklist")
	}

	track, err := tracklist.GetTrack(trackID)
	if err != nil {
		return mark.Wrap(err, TrackNotFound, "Can't find the track in this tracklist")
	}

	updatedTrack, err := updater(track)
	if err != nil {
		return mark.Wrap(err, DefaultErrorMark, "The updater failed to make changes to the track")
	}

	trackAsMap, err := updatedTrack.ToMap()
	if err != nil {
		return mark.Wrap(err, MarshalMark, "Failed to marshal track entity to map")
	}

	for i := 0; i < maxTrackSize; i++ {
		err = d.setTrackDBAtIndex(ctx, tracklistID, track.GetID(), trackAsMap, i)
		if err == nil {
			break
		}
	}

	if err != nil {
		return mark.Wrap(err, TrackNotFound, "Unable to set the track")
	}

	return nil
}

func (d DB) setTrackDBAtIndex(ctx context.Context, tracklistID string, trackID string, trackAsMap map[string]any, trackIndex int) error {
	err := d.dynamoDB.Table(TracklistsTable).
		Update(idKey, tracklistID).
		Set(fmt.Sprintf("tracks[%d]", trackIndex), trackAsMap).
		If(fmt.Sprintf("tracks[%d].id = ?", trackIndex), trackID).
		RunWithContext(ctx)

	if err != nil {
		return errors.Wrap(err, "Failed to update track at index")
	}

	return nil
}
