package auth

import "github.com/veedubyou/chord-paper-be/server/src/errors/api"

const (
	NotGoogleAuthorizedCode    = api.ErrorCode("failed_google_verification")
	NoAccountCode              = api.ErrorCode("no_account")
	WrongOwnerCode             = api.ErrorCode("wrong_owner")
	BadAuthorizationHeaderCode = api.ErrorCode("bad_header")
)
