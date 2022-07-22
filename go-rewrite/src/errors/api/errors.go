package api

import "github.com/cockroachdb/errors"

type ErrorCode string

var DefaultErrorCode = ErrorCode("unknown_error")

func WrapError(err *Error, msg string) *Error {
	return &Error{
		ErrorCode:     err.ErrorCode,
		UserMessage:   err.UserMessage,
		InternalError: errors.Wrap(err.InternalError, msg),
	}
}

func CommitError(err error, errorCode ErrorCode, userMessage string) *Error {
	return &Error{
		ErrorCode:     errorCode,
		UserMessage:   userMessage,
		InternalError: err,
	}
}

// supposedly, having a public error type in Go is an antipattern
// however, all usecase methods will return this expected structure
// and the gateways will all require this metadata to return a proper error
// so I believe it's worth trying a concrete error type
// instead of a bunch of package methods that guess at your error's insides
type Error struct {
	ErrorCode     ErrorCode
	UserMessage   string
	InternalError error
}

func (e Error) Cause() error {
	return e.InternalError
}

func (e Error) Error() string {
	return e.InternalError.Error()
}
