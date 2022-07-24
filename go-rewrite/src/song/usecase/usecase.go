package songusecase

import (
	"context"
	"github.com/cockroachdb/errors/markers"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/veedubyou/chord-paper-be/go-rewrite/src/errors/api"
	"github.com/veedubyou/chord-paper-be/go-rewrite/src/errors/auth"
	songentity "github.com/veedubyou/chord-paper-be/go-rewrite/src/song/entity"
	songerrors "github.com/veedubyou/chord-paper-be/go-rewrite/src/song/errors"
	songstorage "github.com/veedubyou/chord-paper-be/go-rewrite/src/song/storage"
	userusecase "github.com/veedubyou/chord-paper-be/go-rewrite/src/user/usecase"
)

type Usecase struct {
	db          songstorage.DB
	userUsecase userusecase.Usecase
}

func NewUsecase(db songstorage.DB, userUsecase userusecase.Usecase) Usecase {
	return Usecase{
		db:          db,
		userUsecase: userUsecase,
	}
}

func (u Usecase) GetSong(ctx context.Context, songID uuid.UUID) (songentity.Song, *api.Error) {
	song, err := u.db.GetSong(ctx, songID)
	if err != nil {
		err = errors.Wrap(err, "Failed to get the song from DB")

		switch {
		case markers.Is(err, songstorage.SongNotFoundMark):
			return songentity.Song{}, api.CommitError(err,
				songerrors.SongNotFoundCode,
				"The song can't be found")

		case markers.Is(err, songstorage.SongUnmarshalMark):
			fallthrough
		case markers.Is(err, songstorage.DefaultErrorMark):
			fallthrough
		default:
			return songentity.Song{}, api.CommitError(err,
				api.DefaultErrorCode,
				"Unknown error: Failed to fetch the song")
		}
	}

	return song, nil
}

func (u Usecase) CreateSong(ctx context.Context, authHeader string, song songentity.Song) (songentity.Song, *api.Error) {
	if !song.IsNew() {
		return songentity.Song{}, api.CommitError(
			errors.New("Song has an assigned ID already"),
			songerrors.ExistingSongCode,
			"This song is not in an unsaved state and can't be created")
	}

	if song.Defined.Owner == "" {
		return songentity.Song{}, api.CommitError(
			errors.New("Owner field is empty"),
			auth.WrongOwnerCode,
			"No owner is defined on the song. Check that your login status is valid")
	}

	apiErr := u.userUsecase.VerifyOwner(ctx, authHeader, song.Defined.Owner)
	if apiErr != nil {
		return songentity.Song{}, api.WrapError(apiErr, "Failed to verify song owner")
	}

	song.CreateID()
	song.SetSavedTime()
	createdSong, err := u.db.CreateSong(ctx, song)
	if err != nil {
		return songentity.Song{}, api.CommitError(
			errors.Wrap(err, "Failed to create the song in the DB"),
			api.DefaultErrorCode,
			"Unknown error: Failed to create the song")
	}

	return createdSong, nil
}
