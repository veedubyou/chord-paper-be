package trackusecase

import (
	"context"
	"encoding/json"
	"github.com/apex/log"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/errors/markers"
	"github.com/rabbitmq/amqp091-go"
	"github.com/veedubyou/chord-paper-be/src/server/internal/errors/api"
	"github.com/veedubyou/chord-paper-be/src/server/internal/song/usecase"
	"github.com/veedubyou/chord-paper-be/src/server/internal/track/errors"
	"github.com/veedubyou/chord-paper-be/src/shared/lib/rabbitmq"
	"github.com/veedubyou/chord-paper-be/src/shared/track/entity"
	"github.com/veedubyou/chord-paper-be/src/shared/track/storage"
)

type Usecase struct {
	db          trackentity.Store
	songUsecase songusecase.Usecase
	publisher   rabbitmq.QueuePublisher
}

func NewUsecase(db trackentity.Store, songUsecase songusecase.Usecase, publisher rabbitmq.QueuePublisher) Usecase {
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

	newSplitRequests := initializeNewSplitRequests(tracklist)

	tracklist.EnsureTrackIDs()

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
	go u.publishAllSplitRequests(tracklist.Defined.SongID, newSplitRequests)

	return tracklist, nil
}

type failedSplitJob struct {
	track *trackentity.SplitRequestTrack
	err   error
}

func initializeNewSplitRequests(tracklist trackentity.TrackList) []*trackentity.SplitRequestTrack {
	newSplitRequests := []*trackentity.SplitRequestTrack{}
	for _, track := range tracklist.Defined.Tracks {
		if track.IsNew() {
			if splitRequest, ok := track.(*trackentity.SplitRequestTrack); ok {
				splitRequest.InitializeRequest()
				newSplitRequests = append(newSplitRequests, splitRequest)
			}
		}
	}

	return newSplitRequests
}

func (u Usecase) publishAllSplitRequests(tracklistID string, newSplitRequests []*trackentity.SplitRequestTrack) {
	failedSplitJobs := []failedSplitJob{}
	for _, track := range newSplitRequests {
		err := u.publishSplitJob(tracklistID, track.ID)
		if err != nil {
			err = errors.Wrap(err, "Failed to publish split job for track")
			failedSplitJobs = append(failedSplitJobs, failedSplitJob{
				track: track,
				err:   err,
			})
		}
	}

	for _, failedJob := range failedSplitJobs {
		u.markSplitJobFailed(tracklistID, failedJob)
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
	updater := func(track trackentity.Track) (trackentity.Track, error) {
		failedTrack := failedSplitJob.track
		failedTrack.Status = "error"
		failedTrack.StatusMessage = ""
		failedTrack.StatusDebugLog = failedSplitJob.err.Error()
		failedTrack.Progress = 10
		return failedTrack, nil
	}

	trackID := failedSplitJob.track.ID

	err := u.db.UpdateTrack(context.Background(), tracklistID, trackID, updater)
	if err != nil {
		log.WithField("track", failedSplitJob.track).
			Error("Failed to set track in DB")
		return
	}
}
