package google_id

import "github.com/cockroachdb/errors/domains"

var (
	MalformedClaimsMark = domains.New("malformed_claims")
	NotValidatedMark    = domains.New("not_validated")
)
