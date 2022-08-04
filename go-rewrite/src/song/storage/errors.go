package songstorage

import "github.com/cockroachdb/errors/domains"

var (
	SongUnmarshalMark     = domains.New("song_unmarshal_fail")
	SongNotFoundMark      = domains.New("song_not_found")
	SongAlreadyExistsMark = domains.New("song_already_exists")
	DefaultErrorMark      = domains.New("default_error")
)
