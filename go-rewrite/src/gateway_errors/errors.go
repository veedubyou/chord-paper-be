package gateway_errors

// interface check
var _ = []GatewayError{
	SongNotFoundError{},
	InvalidIDError{},
	InternalError{},
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

func NewSongNotFoundError(err error) SongNotFoundError {
	return SongNotFoundError{
		ErrorMsger: NewErrorMsger("The specified song couldn't be found", err),
	}
}

type SongNotFoundError struct {
	NotFoundStatus
	ErrorMsger
}

func (SongNotFoundError) Code() string { return "song_not_found" }

func NewInvalidIDError(err error) InvalidIDError {
	return InvalidIDError{
		ErrorMsger: NewErrorMsger("The specified ID is malformed", err),
	}
}

type InvalidIDError struct {
	BadRequestStatus
	ErrorMsger
}

func (InvalidIDError) Code() string { return "invalid_id" }
