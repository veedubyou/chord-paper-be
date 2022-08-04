package trackentity

const (
	idKey    = "song_id"
	trackKey = "tracks"
)

type TrackList map[string]interface{}

func NewTrackList(songID string) TrackList {
	trackList := TrackList{}
	trackList[idKey] = songID
	trackList[trackKey] = []interface{}{}

	return trackList
}
