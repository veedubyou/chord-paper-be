package songgateway

import (
	"github.com/veedubyou/chord-paper-be/go-rewrite/src/errors/api"
)

var (
	BadSongDataCode = api.ErrorCode("bad_song_data")
	InvalidIDCode   = api.ErrorCode("invalid_id")
)
