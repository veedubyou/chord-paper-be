package trackusecase

import (
	"context"
	"encoding/json"
	"github.com/apex/log"
	"github.com/cockroachdb/errors/markers"
	"github.com/pkg/errors"
	"github.com/rabbitmq/amqp091-go"
	"github.com/veedubyou/chord-paper-be/server/src/internal/errors/api"
	"github.com/veedubyou/chord-paper-be/server/src/internal/lib/rabbitmq"
	songusecase "github.com/veedubyou/chord-paper-be/server/src/internal/song/usecase"
	trackentity "github.com/veedubyou/chord-paper-be/server/src/internal/track/entity"
	trackerrors "github.com/veedubyou/chord-paper-be/server/src/internal/track/errors"
	trackstorage "github.com/veedubyou/chord-paper-be/server/src/internal/track/storage"
)

type Usecase struct {
	db          trackstorage.DB
	songUsecase songusecase.Usecase
	publisher   rabbitmq.Publisher
}

func NewUsecase(db trackstorage.DB, songUsecase songusecase.Usecase, publisher rabbitmq.Publisher) Usecase {
	return Usecase{
		db:          db,
		songUsecase: songUsecase,
		publisher:   publisher,
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

		case markers.Is(err, trackstorage.UnmarshalMark):
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

func (u Usecase) SetTrackList(ctx context.Context, authHeader string, songID string, tracklist trackentity.TrackList) (trackentity.TrackList, *api.Error) {
	if apiErr := u.songUsecase.VerifySongOwnerBySongID(ctx, authHeader, songID); apiErr != nil {
		return trackentity.TrackList{},
			api.WrapError(apiErr, "Cannot verify the song owner for the tracklist")
	}

	// just overwrite the song ID in case there's any discrepancies
	tracklist.Defined.SongID = songID

	newTrackIDs := tracklist.EnsureTrackIDs()

	for i := range tracklist.Defined.Tracks {
		track := &tracklist.Defined.Tracks[i]
		isNewTrack := newTrackIDs[track.Defined.ID]
		if isNewTrack && track.IsSplitRequest() {
			track.InitializeSplitJob()
		}
	}

	err := u.db.SetTrackList(ctx, tracklist)
	if err != nil {
		err = errors.Wrap(err, "Failed to set tracklist")
		switch {
		case markers.Is(err, trackstorage.TrackSizeExceeded):
			return trackentity.TrackList{}, api.CommitError(err,
				trackerrors.TrackListSizeExceeded,
				"The amount of tracks inside the tracklist has exceeded the maximum: 10")
		case markers.Is(err, trackstorage.IDEmptyMark):
			// this should have been handled in the ID assignment above
			fallthrough
		case markers.Is(err, trackstorage.DefaultErrorMark):
			fallthrough
		default:
			return trackentity.TrackList{}, api.CommitError(err,
				api.DefaultErrorCode,
				"Unknown error: Failed to save the track list. Please contact the developer")
		}
	}

	// do this as non-blocking as it's a long term async work
	go u.publishAllSplitRequests(tracklist, newTrackIDs)

	return tracklist, nil
}

type failedSplitJob struct {
	track trackentity.Track
	err   error
}

func (u Usecase) publishAllSplitRequests(tracklist trackentity.TrackList, newTrackIDs map[string]bool) {
	failedSplitJobs := []failedSplitJob{}
	for _, track := range tracklist.Defined.Tracks {
		isNewTrack := newTrackIDs[track.Defined.ID]
		if isNewTrack && track.IsSplitRequest() {
			err := u.publishSplitJob(tracklist.Defined.SongID, track.Defined.ID)
			if err != nil {
				err = errors.Wrap(err, "Failed to publish split job for track")
				failedSplitJobs = append(failedSplitJobs, failedSplitJob{
					track: track,
					err:   err,
				})
			}
		}
	}

	for _, failedJob := range failedSplitJobs {
		u.markSplitJobFailed(tracklist.Defined.SongID, failedJob)
	}
}

type TrackIdentifier struct {
	TrackListID string `json:"tracklist_id"`
	TrackID     string `json:"track_id"`
}

func (u Usecase) publishSplitJob(tracklistID string, trackID string) error {
	jsonBytes, err := json.Marshal(TrackIdentifier{
		TrackListID: tracklistID,
		TrackID:     trackID,
	})

	if err != nil {
		return errors.Wrap(err, "Failed to marshal tracklist and track IDs for queue msg")
	}

	publishMsg := amqp091.Publishing{
		Type: "start_job",
		Body: jsonBytes,
	}

	err = u.publisher.Publish(publishMsg)
	if err != nil {
		return errors.Wrap(err, "Failed to publish message to rabbitmq")
	}

	return nil
}

func (u Usecase) markSplitJobFailed(tracklistID string, failedSplitJob failedSplitJob) {
	failedJobFields := trackentity.SplitJobFields{
		Status:         "error",
		StatusMessage:  "",
		StatusDebugLog: failedSplitJob.err.Error(),
		Progress:       10,
	}

	failedTrack := failedSplitJob.track
	err := failedTrack.SetSplitJobFields(failedJobFields)
	if err != nil {
		log.WithField("failedJobFields", failedJobFields).
			Error("Failed to set the fields into the current track")
		return
	}

	err = u.db.UpdateTrack(context.Background(), tracklistID, failedTrack)
	if err != nil {
		log.WithField("track", failedTrack).
			Error("Failed to set track in DB")
		return
	}
}
