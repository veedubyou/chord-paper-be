package songusecase

import (
	"context"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	songentity "github.com/veedubyou/chord-paper-be/go-rewrite/src/song/entity"
	songstorage "github.com/veedubyou/chord-paper-be/go-rewrite/src/song/storage"
)

type Usecase struct {
	db songstorage.DB
}

func NewUsecase(db songstorage.DB) Usecase {
	return Usecase{
		db: db,
	}
}

func (u Usecase) GetSong(ctx context.Context, songID uuid.UUID) (songentity.Song, error) {
	song, err := u.db.GetSong(ctx, songID)
	if err != nil {
		return songentity.Song{}, errors.Wrap(err, "Failed to GetTrackList")
	}

	return song, nil
}