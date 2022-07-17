package songgateway

import "github.com/veedubyou/chord-paper-be/go-rewrite/src/gateway_errors"

var _ = []gateway_errors.GatewayError{
	SongNotFoundError{},
	InvalidIDError{},
}

func NewSongNotFoundError(err error) SongNotFoundError {
	return SongNotFoundError{
		ErrorMsger: gateway_errors.NewErrorMsger("The specified song couldn't be found", err),
	}
}

type SongNotFoundError struct {
	gateway_errors.NotFoundStatus
	gateway_errors.ErrorMsger
}

func (SongNotFoundError) Code() string { return "song_not_found" }

func NewInvalidIDError(err error) InvalidIDError {
	return InvalidIDError{
		ErrorMsger: gateway_errors.NewErrorMsger("The specified ID is malformed", err),
	}
}

type InvalidIDError struct {
	gateway_errors.BadRequestStatus
	gateway_errors.ErrorMsger
}

func (InvalidIDError) Code() string { return "invalid_id" }
