package userusecase

import "github.com/cockroachdb/errors/domains"

var (
	NotAuthorizedMark          = domains.New("not_authorized")
	NoAccountMark              = domains.New("no_account")
	BadAuthorizationHeaderMark = domains.New("bad_header")
	DefaultErrorMark           = domains.New("default_error")
)
