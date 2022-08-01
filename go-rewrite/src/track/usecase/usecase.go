package trackusecase

import (
	"context"
	"github.com/cockroachdb/errors/markers"
	"github.com/pkg/errors"
	"github.com/veedubyou/chord-paper-be/go-rewrite/src/errors/api"
	trackentity "github.com/veedubyou/chord-paper-be/go-rewrite/src/track/entity"
	trackstorage "github.com/veedubyou/chord-paper-be/go-rewrite/src/track/storage"
)

type Usecase struct {
	db trackstorage.DB
}

func NewUsecase(db trackstorage.DB) Usecase {
	return Usecase{
		db: db,
	}
}

func (u Usecase) GetTrackList(ctx context.Context, songID string) (trackentity.TrackList, *api.Error) {
	tracklist, err := u.db.GetTrackList(ctx, songID)
	if err != nil {
		err = errors.Wrap(err, "Failed to get tracklist from DB")
		switch {
		case markers.Is(err, trackstorage.TrackListNotFound):
			// presume the model where all songs have a tracklist
			// just whether they've been filled in or not
			return trackentity.NewTrackList(songID), nil

		case markers.Is(err, trackstorage.TrackListUnmarshalMark):
			fallthrough
		case markers.Is(err, trackstorage.DefaultErrorMark):
			fallthrough
		default:
			return trackentity.TrackList{}, api.CommitError(err,
				api.DefaultErrorCode,
				"Unknown Error: Failed to fetch Track List")
		}
	}

	return tracklist, nil
}
