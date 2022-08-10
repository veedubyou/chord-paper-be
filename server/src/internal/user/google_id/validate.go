package google_id

import (
	"context"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/errors/domains"
	"github.com/cockroachdb/errors/markers"
	"github.com/veedubyou/chord-paper-be/server/src/internal/lib/errors/mark"
	userentity "github.com/veedubyou/chord-paper-be/server/src/internal/user/entity"
	"google.golang.org/api/idtoken"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Validator
type Validator interface {
	ValidateToken(ctx context.Context, requestToken string) (userentity.User, error)
}

type GoogleValidator struct {
	ClientID string
}

func (g GoogleValidator) ValidateToken(ctx context.Context, requestToken string) (userentity.User, error) {
	validationResult, err := idtoken.Validate(ctx, requestToken, g.ClientID)
	if err != nil {
		return userentity.User{}, mark.Wrap(err, NotValidatedMark, "Token could not be validated")
	}

	sub, err := getStringField(validationResult.Claims, "sub")
	if err != nil {
		return userentity.User{}, mark.Wrap(err, MalformedClaimsMark, "sub field on claims is malformed")
	}

	name, err := getStringField(validationResult.Claims, "name")
	if err != nil && !markers.Is(err, keyNotFound) {
		return userentity.User{}, mark.Wrap(err, MalformedClaimsMark, "name field on claims is malformed")
	}

	email, err := getStringField(validationResult.Claims, "email")
	if err != nil && !markers.Is(err, keyNotFound) {
		return userentity.User{}, mark.Wrap(err, MalformedClaimsMark, "email field on claims is malformed")
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
