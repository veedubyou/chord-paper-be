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
	"sync"
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

func (u Usecase) GetSongSummariesForUser(ctx context.Context, authHeader string, ownerID string) ([]songentity.SongSummary, *api.Error) {
	waitgroup := sync.WaitGroup{}

	waitgroup.Add(2)

	var verifyOwnerErr *api.Error
	verifyOwner := func() {
		defer waitgroup.Done()
		verifyOwnerErr = u.userUsecase.VerifyOwner(ctx, authHeader, ownerID)
	}

	var summaries []songentity.SongSummary
	var getSummariesErr error
	getSummaries := func() {
		defer waitgroup.Done()
		summaries, getSummariesErr = u.db.GetSongSummariesForUser(ctx, ownerID)
	}

	go verifyOwner()
	go getSummaries()

	waitgroup.Wait()

	if verifyOwnerErr != nil {
		return nil, api.WrapError(verifyOwnerErr, "Could not verify owner ID")
	}

	if getSummariesErr != nil {
		getSummariesErr = errors.Wrap(getSummariesErr, "Could not fetch song summaries")
		switch {
		case markers.Is(getSummariesErr, songstorage.SongUnmarshalMark):
			fallthrough
		case markers.Is(getSummariesErr, songstorage.DefaultErrorMark):
			fallthrough
		default:
			return nil, api.CommitError(getSummariesErr,
				api.DefaultErrorCode,
				"Ran into issues fetching your songs. Please contact the developer")
		}
	}

	return summaries, nil
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

func (u Usecase) DeleteSong(ctx context.Context, authHeader string, songID uuid.UUID) *api.Error {
	if apiErr := u.verifySongOwner(ctx, authHeader, songID); apiErr != nil {
		return apiErr
	}

	err := u.db.DeleteSong(ctx, songID)
	if err != nil {
		err = errors.Wrap(err, "Failed to delete song")
		switch {
		case markers.Is(err, songstorage.SongNotFoundMark):
			return api.CommitError(err,
				songerrors.SongNotFoundCode,
				"The song requested to be deleted couldn't be found")

		case markers.Is(err, songstorage.DefaultErrorMark):
			fallthrough
		default:
			return api.CommitError(err,
				api.DefaultErrorCode,
				"Unknown error: Failed to delete song")
		}
	}

	return nil
}

func (u Usecase) verifySongOwner(ctx context.Context, authHeader string, songID uuid.UUID) *api.Error {
	song, apiErr := u.GetSong(ctx, songID)
	if apiErr != nil {
		return api.WrapError(apiErr, "Failed to fetch song")
	}

	apiErr = u.userUsecase.VerifyOwner(ctx, authHeader, song.Defined.Owner)
	if apiErr != nil {
		return api.WrapError(apiErr, "Failed to verify song owner")
	}

	return nil
}
