package save_stems_to_db

import (
	"context"
	"encoding/json"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/application/jobs/job_message"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/application/tracks/entity"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/lib/cerr"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

var postSplitTrackType = map[entity.TrackType]entity.TrackType{
	entity.SplitTwoStemsType:  entity.TwoStemsType,
	entity.SplitFourStemsType: entity.FourStemsType,
	entity.SplitFiveStemsType: entity.FiveStemsType,
}

const JobType string = "save_stems_to_db"
const ErrorMessage string = "Failed to save stem URLs to database"

type JobParams struct {
	job_message.TrackIdentifier
	StemURLS map[string]string `json:"stem_urls"`
}

//counterfeiter:generate . SaveStemsJobHandler
type SaveStemsJobHandler interface {
	HandleSaveStemsToDBJob(message []byte) error
}

func NewJobHandler(trackStore entity.TrackStore) JobHandler {
	return JobHandler{
		trackStore: trackStore,
	}
}

type JobHandler struct {
	trackStore entity.TrackStore
}

func (s JobHandler) HandleSaveStemsToDBJob(message []byte) error {
	params, err := unmarshalMessage(message)
	if err != nil {
		return cerr.Wrap(err).Error("Failed to unmarshal message JSON")
	}

	errctx := cerr.Field("job_params", params)

	updater := func(track entity.Track) (entity.Track, error) {
		splitStemTrack, ok := track.(entity.SplitStemTrack)
		if !ok {
			return entity.BaseTrack{}, errctx.Error("Unexpected - track is not a split request")
		}

		newTrackType, ok := postSplitTrackType[splitStemTrack.TrackType]
		if !ok {
			return entity.BaseTrack{}, errctx.Field("track", splitStemTrack).
				Error("No matching entry for setting the new track type")
		}

		newTrack := entity.StemTrack{
			BaseTrack: entity.BaseTrack{
				TrackType: newTrackType,
			},
			StemURLs: params.StemURLS,
		}

		return newTrack, nil
	}

	err = s.trackStore.UpdateTrack(context.Background(), params.TrackListID, params.TrackID, updater)
	if err != nil {
		return errctx.Wrap(err).Error("Failed to update track")
	}

	return nil
}

func unmarshalMessage(message []byte) (JobParams, error) {
	params := JobParams{}
	err := json.Unmarshal(message, &params)
	if err != nil {
		return JobParams{}, cerr.Wrap(err).Error("Failed to unmarshal message JSON")
	}

	errctx := cerr.Field("job_params", params)

	if params.TrackListID == "" {
		return JobParams{}, errctx.Error("Missing tracklist ID")
	}

	if params.TrackID == "" {
		return JobParams{}, errctx.Error("Missing track ID")
	}

	if len(params.StemURLS) == 0 {
		return JobParams{}, errctx.Error("Missing stem URLS")
	}

	return params, nil
}
