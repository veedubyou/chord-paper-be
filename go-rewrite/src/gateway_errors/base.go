package gateway_errors

import (
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"net/http"
)

func ErrorResponse(c echo.Context, s GatewayError) error {
	return c.JSON(s.StatusCode(), JSONGatewayError{
		Code: s.Code(),
		Msg:  s.Msg(),
	})
}

type JSONGatewayError struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
}

type GatewayError interface {
	StatusCode() int
	Code() string
	Msg() string
}

func NewErrorMsger(message string, err error) ErrorMsger {
	return ErrorMsger{
		Message: message,
		Err:     err,
	}
}

type ErrorMsger struct {
	Message string
	Err     error
}

func (m ErrorMsger) Msg() string {
	err := errors.Wrap(m.Err, m.Message)
	return err.Error()
}

func NewInternalError(err error) InternalError {
	return InternalError{
		ErrorMsger: NewErrorMsger("Something unexpected happened. It's not your fault, it's ours. But it might not be fixed until we find it.", err),
	}
}

type InternalError struct {
	InternalErrorStatus
	ErrorMsger
}

func (InternalError) Code() string { return "internal_error" }

type NotFoundStatus struct{}

func (NotFoundStatus) StatusCode() int { return http.StatusNotFound }

type BadRequestStatus struct{}

func (BadRequestStatus) StatusCode() int { return http.StatusBadRequest }

type InternalErrorStatus struct{}

func (InternalErrorStatus) StatusCode() int { return http.StatusInternalServerError }

type UnauthorizedErrorStatus struct{}

func (UnauthorizedErrorStatus) StatusCode() int { return http.StatusUnauthorized }
