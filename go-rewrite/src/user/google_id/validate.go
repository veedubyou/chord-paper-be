package google_id

import (
	"context"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/errors/domains"
	"github.com/cockroachdb/errors/markers"
	"github.com/veedubyou/chord-paper-be/go-rewrite/src/lib/errors/handle"
	userentity "github.com/veedubyou/chord-paper-be/go-rewrite/src/user/entity"
	"google.golang.org/api/idtoken"
	"strings"
)

const (
	BEARER_PREFIX = "Bearer "
	// should get injected as an env var, but YAGNI for now as it's not a secret
	// and there's no other case to reflect this need
	clientID = "650853277550-ta69qbfcvdl6tb5ogtnh2d07ae9rcdlf.apps.googleusercontent.com"
)

func ValidateHeader(ctx context.Context, header string) (userentity.User, error) {
	if !strings.HasPrefix(header, BEARER_PREFIX) {
		return userentity.User{}, handle.Message(BadAuthorizationHeaderMark, "Authorization header has unexpected shape")
	}

	token := strings.TrimPrefix(header, BEARER_PREFIX)
	return ValidateToken(ctx, token)
}

func ValidateToken(ctx context.Context, requestToken string) (userentity.User, error) {
	validationResult, err := idtoken.Validate(ctx, requestToken, clientID)
	if err != nil {
		return userentity.User{}, handle.Wrap(err, NotValidatedMark, "Token could not be validated")
	}

	sub, err := getStringField(validationResult.Claims, "sub")
	if err != nil {
		return userentity.User{}, handle.Wrap(err, MalformedClaimsMark, "sub field on claims is malformed")
	}

	name, err := getStringField(validationResult.Claims, "name")
	if err != nil && !markers.Is(err, keyNotFound) {
		return userentity.User{}, handle.Wrap(err, MalformedClaimsMark, "name field on claims is malformed")
	}

	email, err := getStringField(validationResult.Claims, "email")
	if err != nil && !markers.Is(err, keyNotFound) {
		return userentity.User{}, handle.Wrap(err, MalformedClaimsMark, "email field on claims is malformed")
	}

	return userentity.User{
		ID:    sub,
		Name:  name,
		Email: email,
	}, nil
}

var (
	keyNotFound    = domains.New("The specified key couldn't be found in the claims")
	valueNotString = domains.New("Unexpected: the retrieved value is not string type")
)

func getStringField(claims map[string]interface{}, key string) (string, error) {
	value, ok := claims[key]
	if !ok {
		return "", errors.Wrap(keyNotFound, "The key "+key+" couldn't be found")
	}

	valueStr, ok := value.(string)
	if !ok {
		return "", errors.Wrap(valueNotString, "The key "+key+" has a non string value")
	}

	return valueStr, nil
}
