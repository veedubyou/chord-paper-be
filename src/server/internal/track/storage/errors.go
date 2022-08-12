package trackstorage

import "github.com/cockroachdb/errors/domains"

var (
	TrackListNotFound = domains.New("tracklist_not_found")
	TrackSizeExceeded = domains.New("track_size_exceeded")
	MarshalMark       = domains.New("marshal_fail")
	UnmarshalMark     = domains.New("unmarshal_fail")
	TrackNotFound     = domains.New("track_not_found")
	IDEmptyMark       = domains.New("id_empty")
	DefaultErrorMark  = domains.New("default_error")
)
