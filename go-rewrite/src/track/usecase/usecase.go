package trackusecase

import (
	"context"
	"github.com/google/uuid"
	"github.com/pkg/errors"
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
		return trackentity.TrackList{}, errors.Wrap(err, "Failed to GetTrackList")
	}

	return tracklist, nil
}
