package gateway

import (
	"github.com/labstack/echo/v4"
	"github.com/veedubyou/chord-paper-be/go-rewrite/src/errors/api"
	"net/http"
)

//TODO
var httpStatusCodeMap = map[api.ErrorCode]int{}

type JSONAPIError struct {
	Code         string `json:"code"`
	Msg          string `json:"msg"`
	ErrorDetails string `json:"error_details"`
}

func ErrorResponse(c echo.Context, err *api.Error) error {
	statusCode, ok := httpStatusCodeMap[err.ErrorCode]
	if !ok {
		statusCode = http.StatusInternalServerError
	}

	return c.JSON(statusCode, JSONAPIError{
		Code:         string(err.ErrorCode),
		Msg:          err.UserMessage,
		ErrorDetails: err.Error(),
	})
}
