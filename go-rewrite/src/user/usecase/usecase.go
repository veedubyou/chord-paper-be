package userusecase

import (
	"context"
	"github.com/cockroachdb/errors/markers"
	"github.com/pkg/errors"
	"github.com/veedubyou/chord-paper-be/go-rewrite/src/errors/api"
	"github.com/veedubyou/chord-paper-be/go-rewrite/src/errors/auth"
	userentity "github.com/veedubyou/chord-paper-be/go-rewrite/src/user/entity"
	"github.com/veedubyou/chord-paper-be/go-rewrite/src/user/google_id"
	"github.com/veedubyou/chord-paper-be/go-rewrite/src/user/storage"
	"golang.org/x/sync/errgroup"
	"strings"
)

const (
	bearerPrefix = "Bearer "
)

type Usecase struct {
	db             userstorage.DB
	googleClientID string
}

func NewUsecase(db userstorage.DB, googleClientID string) Usecase {
	return Usecase{
		db:             db,
		googleClientID: googleClientID,
	}
}

func (u Usecase) VerifyOwner(ctx context.Context, authHeader string, ownerID string) *api.Error {
	group, ctx := errgroup.WithContext(ctx)

	validateHeader := func() error {
		userFromGoogle, apiErr := u.validateHeader(ctx, authHeader)
		if apiErr != nil {
			return api.WrapError(apiErr, "Failed to validate auth header")
		}

		if userFromGoogle.ID != ownerID {
			return api.CommitError(errors.New("Owner ID and user Google ID don't match"),
				auth.WrongOwnerCode,
				"The user requesting access doesn't match the owner")
		}

		return nil
	}

	getUser := func() error {
		_, apiErr := u.getUser(ctx, ownerID)
		if apiErr != nil {
			return api.WrapError(apiErr, "Failed to get user")
		}

		return nil
	}

	group.Go(validateHeader)
	group.Go(getUser)

	if err := group.Wait(); err != nil {
		apiErr, ok := err.(*api.Error)
		if !ok {
			panic("Verify owner: Error is expected to be API error type!")
		}

		return api.WrapError(apiErr, "Could not verify valid user")
	}

	return nil
}

func (u Usecase) Login(ctx context.Context, authHeader string) (userentity.User, *api.Error) {
	userFromGoogle, apiErr := u.validateHeader(ctx, authHeader)
	if apiErr != nil {
		return userentity.User{}, api.WrapError(apiErr, "Failed to validate auth header")
	}

	userFromDB, apiErr := u.getUser(ctx, userFromGoogle.ID)
	if apiErr != nil {
		return userentity.User{}, api.WrapError(apiErr, "Failed to fetch user")
	}

	return userFromDB, nil
}

func (u Usecase) getUser(ctx context.Context, userID string) (userentity.User, *api.Error) {
	userFromDB, err := u.db.GetUser(ctx, userID)
	if err != nil {
		switch {
		case markers.Is(err, userstorage.UserNotFoundMark):
			return userentity.User{}, api.CommitError(err,
				auth.NoAccountCode,
				"A Chord Paper account could not be found for this user")

		case markers.Is(err, userstorage.DefaultErrorMark):
			fallthrough
		default:
			return userentity.User{}, api.CommitError(err,
				api.DefaultErrorCode,
				"User information could not be retrieved")
		}
	}

	return userFromDB, nil
}

func (u Usecase) validateHeader(ctx context.Context, header string) (userentity.User, *api.Error) {
	if !strings.HasPrefix(header, bearerPrefix) {
		return userentity.User{}, api.CommitError(
			errors.New("Auth header doesn't have the bearer prefix"),
			auth.BadAuthorizationHeaderCode,
			"Authorization header has unexpected shape")
	}

	token := strings.TrimPrefix(header, bearerPrefix)
	userFromGoogle, err := google_id.ValidateToken(ctx, u.googleClientID, token)
	if err != nil {
		err = errors.Wrap(err, "Failed to validate Google ID token")
		switch {
		case markers.Is(err, google_id.NotValidatedMark):
			return userentity.User{}, api.CommitError(err,
				auth.NotGoogleAuthorizedCode,
				"Your Google login doesn't seem to be valid. Please try again")

		case markers.Is(err, google_id.MalformedClaimsMark):
			fallthrough
		default:
			return userentity.User{}, api.CommitError(err,
				api.DefaultErrorCode,
				"Unknown error: Couldn't verify your Google login status")
		}
	}
	return userFromGoogle, nil
}
