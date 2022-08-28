package start

import (
	"context"
	"encoding/json"
	trackentity "github.com/veedubyou/chord-paper-be/src/shared/track/entity"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/application/jobs/job_message"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/lib/cerr"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

const JobType string = "start_job"
const ErrorMessage string = "Failed to start processing audio splitting"

//counterfeiter:generate . StartJobHandler
type StartJobHandler interface {
	HandleStartJob(message []byte) (JobParams, error)
}

type JobParams struct {
	job_message.TrackIdentifier
}

func NewJobHandler(trackStore trackentity.Store) JobHandler {
	return JobHandler{
		trackStore: trackStore,
	}
}

type JobHandler struct {
	trackStore trackentity.Store
}

func (d JobHandler) HandleStartJob(message []byte) (JobParams, error) {
	params, err := unmarshalMessage(message)
	if err != nil {
		return JobParams{}, cerr.Wrap(err).Error("Failed to unmarshal message JSON")
	}

	errCtx := cerr.Field("tracklist_id", params.TrackListID).
		Field("track_id", params.TrackID)

	updater := func(track trackentity.Track) (trackentity.Track, error) {
		splitStemTrack, ok := track.(*trackentity.SplitRequestTrack)
		if !ok {
			return nil, errCtx.Error("Track from DB is not a split stem track")
		}

		if splitStemTrack.Status != trackentity.RequestedStatus {
			return nil, errCtx.Error("Track is not in requested status, abort processing to be safe")
		}

		splitStemTrack.Status = trackentity.ProcessingStatus

		return splitStemTrack, nil
	}

	err = d.trackStore.UpdateTrack(context.Background(), params.TrackListID, params.TrackID, updater)
	if err != nil {
		return JobParams{}, errCtx.Wrap(err).Error("Failed to set the track status")
	}

	return params, nil
}

func unmarshalMessage(message []byte) (JobParams, error) {
	params := JobParams{}
	err := json.Unmarshal(message, &params)
	if err != nil {
		return JobParams{}, cerr.Wrap(err).Error("Failed to unmarshal message JSON")
	}

	errctx := cerr.Field("job_params", params)

	if params.TrackListID == "" {
		return JobParams{}, errctx.Wrap(err).Error("Missing tracklist ID")
	}

	if params.TrackID == "" {
		return JobParams{}, errctx.Wrap(err).Error("Missing track ID")
	}

	return params, nil
}
