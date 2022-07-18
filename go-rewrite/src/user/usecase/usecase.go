package userusecase

import (
	"context"
	"github.com/cockroachdb/errors/markers"
	"github.com/veedubyou/chord-paper-be/go-rewrite/src/lib/errors/handle"
	userentity "github.com/veedubyou/chord-paper-be/go-rewrite/src/user/entity"
	"github.com/veedubyou/chord-paper-be/go-rewrite/src/user/google_id"
	"github.com/veedubyou/chord-paper-be/go-rewrite/src/user/userstorage"
)

type Usecase struct {
	db userstorage.DB
}

func NewUsecase(db userstorage.DB) Usecase {
	return Usecase{
		db: db,
	}
}

func (u Usecase) Login(ctx context.Context, authHeader string) (userentity.User, error) {
	userFromGoogle, err := google_id.ValidateHeader(ctx, authHeader)
	if err != nil {
		switch {
		case markers.Is(err, google_id.NotValidatedMark):
			return userentity.User{}, handle.Wrap(err, NotAuthorizedMark, "Couldn't validate authorization header")
		case markers.Is(err, google_id.BadAuthorizationHeaderMark):
			return userentity.User{}, handle.Wrap(err, BadAuthorizationHeaderMark, "Malformed authorization header")
		case markers.Is(err, google_id.MalformedClaimsMark):
		default:
			return userentity.User{}, handle.Wrap(err, DefaultErrorMark, "Failed to validate Google ID token")
		}
	}

	userFromDB, err := u.db.GetUser(ctx, userFromGoogle.ID)
	if err != nil {
		switch {
		case markers.Is(err, userstorage.UserNotFoundMark):
			return userentity.User{}, handle.Wrap(err, NoAccountMark, "Account could not be found for this user")

		case markers.Is(err, userstorage.DefaultErrorMark):
		default:
			return userentity.User{}, handle.Wrap(err, DefaultErrorMark, "User information could not be retrieved")
		}
	}

	return userFromDB, nil
}
