package auth

import "github.com/veedubyou/chord-paper-be/go-rewrite/src/errors/api"

var (
	NotGoogleAuthorizedCode    = api.ErrorCode("failed_google_verification")
	NoAccountCode              = api.ErrorCode("no_account")
	WrongOwnerCode             = api.ErrorCode("wrong_owner")
	BadAuthorizationHeaderCode = api.ErrorCode("bad_header")
)
