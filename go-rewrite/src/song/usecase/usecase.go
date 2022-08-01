package songusecase

import (
	"context"
	"github.com/cockroachdb/errors/markers"
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

func (u Usecase) GetSong(ctx context.Context, songID string) (songentity.Song, *api.Error) {
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
	song.SetSavedAtToNow()
	createdSong, err := u.db.CreateSong(ctx, song)
	if err != nil {
		return songentity.Song{}, api.CommitError(
			errors.Wrap(err, "Failed to create the song in the DB"),
			api.DefaultErrorCode,
			"Unknown error: Failed to create the song")
	}

	return createdSong, nil
}

func (u Usecase) UpdateSong(ctx context.Context, authHeader string, songID string, song songentity.Song) (songentity.Song, *api.Error) {
	dbSong, apiErr := u.GetSong(ctx, songID)
	if apiErr != nil {
		return songentity.Song{}, api.WrapError(apiErr, "Failed to fetch song")
	}

	freshlyFetchedSong := FreshlyFetchedSong(dbSong)

	apiErr = u.verifySongOwnerBySong(ctx, authHeader, freshlyFetchedSong)
	if apiErr != nil {
		return songentity.Song{}, api.WrapError(apiErr, "Cannot verify that this user owns this song")
	}

	// set these security fields in case someone wants to pull a fast one
	song.Defined.ID = freshlyFetchedSong.Defined.ID
	song.Defined.Owner = freshlyFetchedSong.Defined.Owner

	apiErr = protectSongFromOverwriting(song, freshlyFetchedSong)
	if apiErr != nil {
		return songentity.Song{}, api.WrapError(apiErr, "Song protected from overwriting")
	}

	song.SetSavedAtToNow()

	updatedSong, err := u.db.UpdateSong(ctx, song)
	if err != nil {
		err = errors.Wrap(err, "Failed to update song in DB")
		switch {
		case markers.Is(err, songstorage.SongNotFoundMark):
			return songentity.Song{}, api.CommitError(err,
				songerrors.SongNotFoundCode,
				"The song to be saved doesn't currently exist in the database")
		case markers.Is(err, songstorage.DefaultErrorMark):
			fallthrough
		default:
			return songentity.Song{}, api.CommitError(err,
				api.DefaultErrorCode,
				"Unknown error: Failed to save updated song")
		}
	}

	return updatedSong, nil
}

func protectSongFromOverwriting(songToUpdate songentity.Song, songFromDB FreshlyFetchedSong) *api.Error {
	if songToUpdate.Defined.LastSavedAt == nil {
		err := errors.New("Song payload doesn't have a last saved at field set")
		return api.CommitError(err,
			songerrors.SongOverwriteCode,
			"This song doesn't have a last saved at time, it may not have been created yet. Please upload the initial copy first")
	}

	if songFromDB.Defined.LastSavedAt == nil {
		// unexpected, any song in the DB should have a last saved at
		// but don't block on this
		return nil
	}

	// prevent overwriting a more recent save if the last saved at timestamp is greater than the current one
	// example:
	// A --> B
	//   \----->C
	//
	// Suppose I open the song at time A on two computers, make edits on both somewhat absentmindedly
	// I first save at time B - the payload for the song for the last saved at would be A
	// presume this succeeds, the server copy is now from time B and its last saved at time is also B
	//
	// Now I save another copy at time C, but the predecessor of my copy at time C was from A
	// If I just go with last write wins, then all changes from the copy at time B would be overwritten
	// So by comparing the timestamp at the save at C (A vs B), the save will fail to protect overwriting data.
	//
	// The user can then refetch the copy at time B and copy over their changes from time C (manual merge)
	// and form a copy that will be accepted by the server:
	//
	// A --> B----->B+C
	//   \----->C---/
	if songToUpdate.Defined.LastSavedAt.Before(*songFromDB.Defined.LastSavedAt) {
		err := errors.New("New song has an earlier last saved at than the current last saved time")
		return api.CommitError(err,
			songerrors.SongOverwriteCode,
			"Unable to save - there's been a more recent copy of this song saved and saving it will clobber it. Please try reloading this song in a new tab and copy your work over")
	}

	return nil
}

func (u Usecase) DeleteSong(ctx context.Context, authHeader string, songID string) *api.Error {
	if apiErr := u.verifySongOwnerBySongID(ctx, authHeader, songID); apiErr != nil {
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

func (u Usecase) verifySongOwnerBySongID(ctx context.Context, authHeader string, songID string) *api.Error {
	song, apiErr := u.GetSong(ctx, songID)
	if apiErr != nil {
		return api.WrapError(apiErr, "Failed to fetch song")
	}

	return u.verifySongOwnerBySong(ctx, authHeader, FreshlyFetchedSong(song))
}

// this type alias is to make it very explicit that the input is for a song that was just fetched
// from the DB, not a song that was provided by the API caller. since they have the same type
// it can be easy to just thread it through. the caller must acknowledge this difference through an explicit cast
type FreshlyFetchedSong songentity.Song

func (u Usecase) verifySongOwnerBySong(ctx context.Context, authHeader string, song FreshlyFetchedSong) *api.Error {
	apiErr := u.userUsecase.VerifyOwner(ctx, authHeader, song.Defined.Owner)
	if apiErr != nil {
		return api.WrapError(apiErr, "Failed to verify song owner")
	}

	return nil
}
