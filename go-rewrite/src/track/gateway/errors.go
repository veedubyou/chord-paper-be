package trackgateway

import "github.com/veedubyou/chord-paper-be/go-rewrite/src/gateway_errors"

var _ = []gateway_errors.GatewayError{
	InvalidIDError{},
}

func NewInvalidIDError(err error) InvalidIDError {
	return InvalidIDError{
		ErrorMsger: gateway_errors.NewErrorMsger("The specified tracklist ID is malformed", err),
	}
}

type InvalidIDError struct {
	gateway_errors.BadRequestStatus
	gateway_errors.ErrorMsger
}

func (InvalidIDError) Code() string { return "invalid_id" }
