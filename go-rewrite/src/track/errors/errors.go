package trackerrors

import (
	"github.com/veedubyou/chord-paper-be/go-rewrite/src/errors/api"
)

const (
	TrackListSizeExceeded = api.ErrorCode("track_list_size_exceeded")
	BadTracklistDataCode  = api.ErrorCode("bad_tracklist_data")
)
