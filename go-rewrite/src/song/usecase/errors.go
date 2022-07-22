package songusecase

import (
	"github.com/veedubyou/chord-paper-be/go-rewrite/src/errors/api"
)

var (
	SongNotFoundCode = api.ErrorCode("song_not_found")
	ExistingSongCode = api.ErrorCode("create_song_exists")
)
