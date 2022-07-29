package gateway

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/veedubyou/chord-paper-be/go-rewrite/src/errors/api"
	"github.com/veedubyou/chord-paper-be/go-rewrite/src/errors/auth"
	songerrors "github.com/veedubyou/chord-paper-be/go-rewrite/src/song/errors"
	"net/http"
)

var httpStatusCodeMap = map[api.ErrorCode]int{
	api.DefaultErrorCode:            http.StatusInternalServerError,
	auth.NotGoogleAuthorizedCode:    http.StatusUnauthorized,
	auth.NoAccountCode:              http.StatusUnauthorized,
	auth.BadAuthorizationHeaderCode: http.StatusBadRequest,
	auth.WrongOwnerCode:             http.StatusForbidden,
	songerrors.SongNotFoundCode:     http.StatusNotFound,
	songerrors.ExistingSongCode:     http.StatusUnprocessableEntity,
	songerrors.BadSongDataCode:      http.StatusBadRequest,
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
