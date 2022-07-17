package songstorage

import "github.com/cockroachdb/errors/domains"

var SongUnmarshalMark = domains.New("song_unmarshal_fail")
var SongNotFoundMark = domains.New("song_not_found")
var DefaultErrorMark = domains.New("default_error")
