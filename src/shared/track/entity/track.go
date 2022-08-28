package trackentity

import (
	"encoding/json"
	"github.com/cockroachdb/errors"
	"github.com/google/uuid"
	"github.com/veedubyou/chord-paper-be/src/shared/lib/jsonlib"
)

const (
	InitialProgressPercentage = 5
)

type StemTrackType string

const (
	TwoStemsType  StemTrackType = "2stems"
	FourStemsType StemTrackType = "4stems"
	FiveStemsType StemTrackType = "5stems"
)

var stemTrackTypes = map[string]bool{
	string(TwoStemsType):  true,
	string(FourStemsType): true,
	string(FiveStemsType): true,
}

type SplitRequestType string

const (
	SplitTwoStemsType  SplitRequestType = "split_2stems"
	SplitFourStemsType SplitRequestType = "split_4stems"
	SplitFiveStemsType SplitRequestType = "split_5stems"
)

var splitTrackTypes = map[string]bool{
	string(SplitTwoStemsType):  true,
	string(SplitFourStemsType): true,
	string(SplitFiveStemsType): true,
}

type Tracks []Track

func (t *Tracks) UnmarshalJSON(b []byte) error {
	*t = nil

	objSlice := []map[string]any{}
	err := json.Unmarshal(b, &objSlice)
	if err != nil {
		return errors.Wrap(err, "Could not unmarshal json data into defined fields")
	}

	for _, obj := range objSlice {
		track, err := UnmarshalTrack(obj)
		if err != nil {
			return errors.Wrap(err, "Failed to unmarshal element track")
		}

		*t = append(*t, track)
	}

	return nil
}

type TrackFields struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

type GenericTrackFields struct {
	TrackFields
	TrackType string `json:"track_type"`
}

type GenericTrack struct {
	jsonlib.Flatten[GenericTrackFields]
}

type StemTrack struct {
	TrackFields
	TrackType StemTrackType     `json:"track_type"`
	StemURLs  map[string]string `json:"stem_urls"`
}

type SplitRequestStatus string

const (
	RequestedStatus  SplitRequestStatus = "requested"
	ProcessingStatus SplitRequestStatus = "processing"
	ErrorStatus      SplitRequestStatus = "error"
)

type SplitRequestTrack struct {
	TrackFields
	TrackType      SplitRequestType   `json:"track_type"`
	OriginalURL    string             `json:"original_url"`
	Status         SplitRequestStatus `json:"job_status"`
	StatusMessage  string             `json:"job_status_message"`
	StatusDebugLog string             `json:"job_status_debug_log"`
	Progress       int                `json:"job_progress"`
}

func (g GenericTrack) GetID() string {
	return g.Defined.ID
}

func (g GenericTrack) IsNew() bool {
	return g.Defined.ID == ""
}

func (g *GenericTrack) CreateID() {
	if !g.IsNew() {
		panic("Cannot assign an ID to a track that already has one")
	}

	g.Defined.ID = uuid.New().String()
}

func (g GenericTrack) ToMap() (map[string]any, error) {
	return jsonlib.StructToMap(g)
}

func (s StemTrack) GetID() string {
	return s.ID
}

func (s StemTrack) IsNew() bool {
	return s.ID == ""
}

func (s *StemTrack) CreateID() {
	if !s.IsNew() {
		panic("Cannot assign an ID to a track that already has one")
	}

	s.ID = uuid.New().String()
}

func (s StemTrack) ToMap() (map[string]any, error) {
	return jsonlib.StructToMap(s)
}

func (s SplitRequestTrack) GetID() string {
	return s.ID
}

func (s SplitRequestTrack) IsNew() bool {
	return s.ID == ""
}

func (s *SplitRequestTrack) CreateID() {
	if !s.IsNew() {
		panic("Cannot assign an ID to a track that already has one")
	}

	s.ID = uuid.New().String()
}

func (s SplitRequestTrack) ToMap() (map[string]any, error) {
	return jsonlib.StructToMap(s)
}

func (s *SplitRequestTrack) InitializeRequest() {
	s.Status = "requested"
	s.StatusMessage = "The splitting job for the audio has been requested"
	s.StatusDebugLog = ""
	s.Progress = InitialProgressPercentage
}

type Track interface {
	GetID() string
	IsNew() bool
	CreateID()
	ToMap() (map[string]any, error)
}

func UnmarshalTrack(contents map[string]any) (Track, error) {
	trackTypeIface, ok := contents["track_type"]
	if !ok {
		return nil, errors.New("no track_type field on this track")
	}

	trackType, ok := trackTypeIface.(string)
	if !ok {
		return nil, errors.New("track_type field is not stored as a string")
	}

	switch {
	case splitTrackTypes[trackType]:
		splitTrack, err := jsonlib.MapToStruct[SplitRequestTrack](contents)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to create split track from map")
		}

		return &splitTrack, nil

	case stemTrackTypes[trackType]:
		stemTrack, err := jsonlib.MapToStruct[StemTrack](contents)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to create stem track from map")
		}

		return &stemTrack, nil
	default:
		genericTrack := GenericTrack{}
		err := genericTrack.FromMap(contents)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to create generic track from map")
		}

		return &genericTrack, nil
	}
}
