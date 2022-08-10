package songerrors

import (
	"github.com/veedubyou/chord-paper-be/server/src/errors/api"
)

const (
	SongNotFoundCode  = api.ErrorCode("song_not_found")
	ExistingSongCode  = api.ErrorCode("create_song_exists")
	BadSongDataCode   = api.ErrorCode("bad_song_data")
	SongOverwriteCode = api.ErrorCode("update_song_overwrite")
)
