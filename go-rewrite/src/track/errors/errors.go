package trackerrors

import (
	"github.com/veedubyou/chord-paper-be/go-rewrite/src/errors/api"
)

const (
	TrackListSizeExceeded = api.ErrorCode("track_list_size_exceeded")
	ExistingSongCode      = api.ErrorCode("create_song_exists")
	NoTracklistIDCode     = api.ErrorCode("no_tracklist_id")
	NoTrackIDCode         = api.ErrorCode("no_track_id")
	SongOverwriteCode     = api.ErrorCode("update_song_overwrite")
	BadTracklistDataCode  = api.ErrorCode("bad_tracklist_data")
)
