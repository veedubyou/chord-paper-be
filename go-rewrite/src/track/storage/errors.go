package trackstorage

import "github.com/cockroachdb/errors/domains"

var (
	TrackListNotFound      = domains.New("tracklist_not_found")
	TrackListUnmarshalMark = domains.New("tracklist_unmarshal_fail")
	DefaultErrorMark       = domains.New("default_error")
)
