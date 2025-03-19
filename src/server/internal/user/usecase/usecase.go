package userusecase

import (
	"context"
	"fmt"
	"github.com/apex/log"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/errors/markers"
	google_id2 "github.com/veedubyou/chord-paper-be/src/server/google_id"
	"github.com/veedubyou/chord-paper-be/src/server/internal/errors/api"
	"github.com/veedubyou/chord-paper-be/src/server/internal/errors/auth"
	"github.com/veedubyou/chord-paper-be/src/server/internal/user/entity"
	"github.com/veedubyou/chord-paper-be/src/server/internal/user/storage"
	"strings"
	"sync"
)

const (
	bearerPrefix = "Bearer "
)

type Usecase struct {
	db              userstorage.DB
	googleValidator google_id2.Validator
}

func NewUsecase(db userstorage.DB, googleValidator google_id2.Validator) Usecase {
	return Usecase{
		db:              db,
		googleValidator: googleValidator,
	}
}

func (u Usecase) VerifyOwner(ctx context.Context, authHeader string, ownerID string) *api.Error {
	var userFromGoogle google_id2.User
	var validateHeaderErr *api.Error
	var getOwnerErr *api.Error

	waitgroup := sync.WaitGroup{}
	waitgroup.Add(2)

	validateHeader := func() {
		defer waitgroup.Done()
		userFromGoogle, validateHeaderErr = u.validateHeader(ctx, authHeader)
	}

	// check for the owner's ID optimistically
	// and then match it to the auth header's
	// if it's wrong, we'll need to do more auth checks to see which error to return
	// but if it's right then we'll save some time
	getOwner := func() {
		defer waitgroup.Done()
		_, getOwnerErr = u.getUser(ctx, ownerID)
	}

	go validateHeader()
	go getOwner()

	waitgroup.Wait()

	if validateHeaderErr != nil {
		return api.WrapError(validateHeaderErr, "Failed to validate auth header")
	}

	if userFromGoogle.GoogleID != ownerID {
		if _, apiErr := u.getUser(ctx, userFromGoogle.GoogleID); apiErr != nil {
			return api.WrapError(apiErr, "Failed to find user account")
		}

		return api.CommitError(
			errors.New("Owner ID and user Google ID don't match"),
			auth.WrongOwnerCode,
			"The user requesting access doesn't match the owner")
	}

	if getOwnerErr != nil {
		return api.WrapError(getOwnerErr, "Failed to find user account")
	}

	return nil
}

func (u Usecase) Login(ctx context.Context, authHeader string) (userentity.User, *api.Error) {
	userFromGoogle, apiErr := u.validateHeader(ctx, authHeader)
	if apiErr != nil {
		return userentity.User{}, api.WrapError(apiErr, "Failed to validate auth header")
	}

	userFromDB, apiErr := u.getUser(ctx, userFromGoogle.GoogleID)
	if apiErr != nil {
		if apiErr.ErrorCode == auth.NoAccountCode {
			go func() {
				err := u.addUnverifiedUser(context.Background(), userFromGoogle)
				if err != nil {
					log.Error(fmt.Sprintf("Failed to add %s, %s, %s", userFromGoogle.GoogleID, userFromGoogle.Name, userFromGoogle.Email))
					log.Error(err.Error())
				}
			}()
		}

		return userentity.User{}, api.WrapError(apiErr, "Failed to fetch user")
	}

	if !userFromDB.Verified {
		err := errors.New("User not verified")
		return userentity.User{}, api.CommitError(err, auth.UnvalidatedAccountCode, "User not verified during login")
	}

	return userFromDB, nil
}

func (u Usecase) addUnverifiedUser(ctx context.Context, googleUser google_id2.User) error {
	newUser := userentity.User{
		ID:       googleUser.GoogleID,
		Name:     googleUser.Name,
		Email:    googleUser.Email,
		Verified: false,
	}

	return u.db.SetUser(ctx, newUser)
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

func (u Usecase) validateHeader(ctx context.Context, header string) (google_id2.User, *api.Error) {
	if !strings.HasPrefix(header, bearerPrefix) {
		return google_id2.User{}, api.CommitError(
			errors.New("Auth header doesn't have the bearer prefix"),
			auth.BadAuthorizationHeaderCode,
			"Authorization header has unexpected shape")
	}

	token := strings.TrimPrefix(header, bearerPrefix)
	userFromGoogle, err := u.googleValidator.ValidateToken(ctx, token)
	if err != nil {
		err = errors.Wrap(err, "Failed to validate Google ID token")
		switch {
		case markers.Is(err, google_id2.NotValidatedMark):
			return google_id2.User{}, api.CommitError(err,
				auth.NotGoogleAuthorizedCode,
				"Your Google login doesn't seem to be valid. Please try again")

		case markers.Is(err, google_id2.MalformedClaimsMark):
			fallthrough
		default:
			return google_id2.User{}, api.CommitError(err,
				api.DefaultErrorCode,
				"Unknown error: Couldn't verify your Google login status")
		}
	}
	return userFromGoogle, nil
}
