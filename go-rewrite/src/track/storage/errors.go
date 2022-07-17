package trackstorage

import "github.com/cockroachdb/errors/domains"

var TrackListNotFound = domains.New("tracklist_not_found")
var TrackListUnmarshalMark = domains.New("tracklist_unmarshal_fail")
var DefaultErrorMark = domains.New("default_error")
