package trackentity

import (
	"github.com/cockroachdb/errors"
	"github.com/veedubyou/chord-paper-be/src/shared/lib/jsonlib"
)

type TrackList struct {
	jsonlib.Flatten[TrackListFields]
}

type TrackListFields struct {
	SongID string `json:"song_id"`
	Tracks Tracks `json:"tracks"`
}

func NewTrackList(songID string) TrackList {
	trackList := TrackList{}
	trackList.Defined.SongID = songID
	trackList.Defined.Tracks = []Track{}

	return trackList
}

func (t *TrackList) EnsureTrackIDs() {
	for _, track := range t.Defined.Tracks {
		if track.IsNew() {
			track.CreateID()
		}
	}
}

func (t TrackList) GetTrack(trackID string) (Track, error) {
	for _, track := range t.Defined.Tracks {
		if track.GetID() == trackID {
			return track, nil
		}
	}

	return nil, errors.New("Failed to find track of the specified ID in tracklist")
}
