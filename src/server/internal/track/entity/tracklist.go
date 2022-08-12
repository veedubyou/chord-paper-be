package trackentity

import (
	"github.com/cockroachdb/errors"
	"github.com/google/uuid"
	"github.com/veedubyou/chord-paper-be/src/server/internal/lib/jsonlib"
)

const (
	InitialProgressPercentage = 5
)

var splitTrackTypes = map[string]bool{
	"split_2stems": true,
	"split_4stems": true,
	"split_5stems": true,
}

type TrackList struct {
	jsonlib.Flatten[TrackListFields]
}

type Track struct {
	jsonlib.Flatten[TrackFields]
}

type TrackFields struct {
	ID        string `json:"id"`
	TrackType string `json:"track_type"`
}

type TrackListFields struct {
	SongID string  `json:"song_id"`
	Tracks []Track `json:"tracks"`
}

type SplitJobFields struct {
	Status         string `json:"job_status"`
	StatusMessage  string `json:"job_status_message"`
	StatusDebugLog string `json:"job_status_debug_log"`
	Progress       int    `json:"job_progress"`
}

func NewTrackList(songID string) TrackList {
	trackList := TrackList{}
	trackList.Defined.SongID = songID
	trackList.Defined.Tracks = []Track{}

	return trackList
}

func (t *TrackList) EnsureTrackIDs() map[string]bool {
	newTrackIDs := map[string]bool{}
	for i := range t.Defined.Tracks {
		track := &t.Defined.Tracks[i]
		if track.IsNew() {
			track.CreateID()
			newTrackIDs[track.Defined.ID] = true
		}
	}

	return newTrackIDs
}

func (t Track) IsNew() bool {
	return t.Defined.ID == ""
}

func (t *Track) CreateID() {
	if !t.IsNew() {
		panic("Cannot assign an ID to a track that already has one")
	}

	t.Defined.ID = uuid.New().String()
}

func (t Track) IsSplitRequest() bool {
	return splitTrackTypes[t.Defined.TrackType]
}

func (t *Track) SetSplitJobFields(fields SplitJobFields) error {
	jobFieldsMap, err := jsonlib.StructToMap(fields)
	if err != nil {
		return errors.Wrap(err, "Failed to convert job fields to a map")
	}

	if t.Extra == nil {
		t.Extra = map[string]interface{}{}
	}

	for k, v := range jobFieldsMap {
		t.Extra[k] = v
	}

	return nil
}

func (t *Track) InitializeSplitJob() {
	if !t.IsSplitRequest() {
		panic(errors.New("InitializeSplitJob called on a non split request"))
	}

	jobFields := SplitJobFields{
		Status:         "requested",
		StatusMessage:  "The splitting job for the audio has been requested",
		StatusDebugLog: "",
		Progress:       InitialProgressPercentage,
	}

	err := t.SetSplitJobFields(jobFields)

	if err != nil {
		// this failure is really unexpected - the fields are static so it should either
		// always fail or never fail. opting to not return the error here because it's expected
		// to never fail
		panic(err)
	}
}
