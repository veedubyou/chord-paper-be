package google_id

import "github.com/cockroachdb/errors/domains"

var (
	MalformedClaimsMark        = domains.New("malformed_claims")
	BadAuthorizationHeaderMark = domains.New("bad_authorization_header")
	NotValidatedMark           = domains.New("not_validated")
)
