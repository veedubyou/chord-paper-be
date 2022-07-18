package usergateway

import "github.com/veedubyou/chord-paper-be/go-rewrite/src/gateway_errors"

var _ = []gateway_errors.GatewayError{
	BadAuthorizationHeaderError{},
	NotAuthorizedError{},
	NotAuthorizedError{},
}

func NewBadAuthorizationHeaderError(err error) BadAuthorizationHeaderError {
	return BadAuthorizationHeaderError{
		ErrorMsger: gateway_errors.NewErrorMsger("The authorization header is malformed", err),
	}
}

type BadAuthorizationHeaderError struct {
	gateway_errors.BadRequestStatus
	gateway_errors.ErrorMsger
}

func (BadAuthorizationHeaderError) Code() string { return "bad_authorization_header" }

func NewNoAccountError(err error) NoAccountError {
	return NoAccountError{
		ErrorMsger: gateway_errors.NewErrorMsger("An account for this user is not found", err),
	}
}

type NoAccountError struct {
	gateway_errors.UnauthorizedErrorStatus
	gateway_errors.ErrorMsger
}

func (NoAccountError) Code() string { return "no_account" }

func NewNotAuthorizedError(err error) NotAuthorizedError {
	return NotAuthorizedError{
		ErrorMsger: gateway_errors.NewErrorMsger("Failed to validate the Google ID provided", err),
	}
}

type NotAuthorizedError struct {
	gateway_errors.UnauthorizedErrorStatus
	gateway_errors.ErrorMsger
}

func (NotAuthorizedError) Code() string { return "failed_google_verification" }
