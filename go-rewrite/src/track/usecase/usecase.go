package trackusecase

import (
	"context"
	"github.com/cockroachdb/errors/markers"
	"github.com/google/uuid"
	"github.com/veedubyou/chord-paper-be/go-rewrite/src/lib/errors/handle"
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

func (u Usecase) GetTrackList(ctx context.Context, songID uuid.UUID) (trackentity.TrackList, error) {
	tracklist, err := u.db.GetTrackList(ctx, songID)
	if err != nil {
		switch {
		case markers.Is(err, trackstorage.TrackListNotFound):
			// presume the model where all songs have a tracklist
			// just whether they've been filled in or not
			return trackentity.NewTrackList(songID.String()), nil

		case markers.Is(err, trackstorage.TrackListUnmarshalMark):
		case markers.Is(err, trackstorage.DefaultErrorMark):
		default:
			return trackentity.TrackList{}, handle.Wrap(err, DefaultErrorMark, "Failed to GetTrackList")
		}
	}

	return tracklist, nil
}
