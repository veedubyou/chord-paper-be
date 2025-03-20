package gateway

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/veedubyou/chord-paper-be/src/server/api_error"
	"github.com/veedubyou/chord-paper-be/src/server/internal/errors/api"
	"github.com/veedubyou/chord-paper-be/src/server/internal/errors/auth"
	"github.com/veedubyou/chord-paper-be/src/server/internal/song/errors"
	"github.com/veedubyou/chord-paper-be/src/server/internal/track/errors"
	"net/http"
)

var httpStatusCodeMap = map[api.ErrorCode]int{
	api.DefaultErrorCode:              http.StatusInternalServerError,
	auth.NotGoogleAuthorizedCode:      http.StatusUnauthorized,
	auth.UnvalidatedAccountCode:       http.StatusUnauthorized,
	auth.NoAccountCode:                http.StatusUnauthorized,
	auth.BadAuthorizationHeaderCode:   http.StatusBadRequest,
	auth.WrongOwnerCode:               http.StatusForbidden,
	songerrors.SongNotFoundCode:       http.StatusNotFound,
	songerrors.ExistingSongCode:       http.StatusBadRequest,
	songerrors.BadSongDataCode:        http.StatusBadRequest,
	songerrors.SongOverwriteCode:      http.StatusBadRequest,
	trackerrors.TrackListSizeExceeded: http.StatusBadRequest,
	trackerrors.BadTracklistDataCode:  http.StatusBadRequest,
}

func ErrorResponse(c echo.Context, err *api.Error) error {
	statusCode, ok := httpStatusCodeMap[err.ErrorCode]
	if !ok {
		msg := fmt.Sprintf("Error code %s has no HTTP status code mapping", err.ErrorCode)
		panic(msg)
	}

	return c.JSON(statusCode, api_error.JSONAPIError{
		Code:         string(err.ErrorCode),
		Msg:          err.UserMessage,
		ErrorDetails: err.Error(),
	})
}
