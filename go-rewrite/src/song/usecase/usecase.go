package songusecase

import (
	"context"
	"github.com/cockroachdb/errors/markers"
	"github.com/google/uuid"
	"github.com/veedubyou/chord-paper-be/go-rewrite/src/lib/errors/handle"
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
		switch {
		case markers.Is(err, songstorage.SongNotFoundMark):
			return songentity.Song{}, handle.Wrap(err, SongNotFoundMark, "Song can't be found")

		case markers.Is(err, songstorage.SongUnmarshalMark):
		case markers.Is(err, songstorage.DefaultErrorMark):
		default:
			return songentity.Song{}, handle.Wrap(err, DefaultErrorMark, "Failed to get song")
		}
	}

	return song, nil
}
