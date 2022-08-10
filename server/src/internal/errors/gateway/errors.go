package gateway

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/veedubyou/chord-paper-be/server/src/errors/api"
	"github.com/veedubyou/chord-paper-be/server/src/errors/auth"
	songerrors "github.com/veedubyou/chord-paper-be/server/src/song/errors"
	trackerrors "github.com/veedubyou/chord-paper-be/server/src/track/errors"
	"net/http"
)

var httpStatusCodeMap = map[api.ErrorCode]int{
	api.DefaultErrorCode:              http.StatusInternalServerError,
	auth.NotGoogleAuthorizedCode:      http.StatusUnauthorized,
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

type JSONAPIError struct {
	Code         string `json:"code"`
	Msg          string `json:"msg"`
	ErrorDetails string `json:"error_details"`
}

func ErrorResponse(c echo.Context, err *api.Error) error {
	statusCode, ok := httpStatusCodeMap[err.ErrorCode]
	if !ok {
		msg := fmt.Sprintf("Error code %s has no HTTP status code mapping", err.ErrorCode)
		panic(msg)
	}

	return c.JSON(statusCode, JSONAPIError{
		Code:         string(err.ErrorCode),
		Msg:          err.UserMessage,
		ErrorDetails: err.Error(),
	})
}
