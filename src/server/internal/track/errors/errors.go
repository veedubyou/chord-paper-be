package trackerrors

import (
	"github.com/veedubyou/chord-paper-be/src/server/internal/errors/api"
)

const (
	TrackListSizeExceeded = api.ErrorCode("track_list_size_exceeded")
	BadTracklistDataCode  = api.ErrorCode("bad_tracklist_data")
)
