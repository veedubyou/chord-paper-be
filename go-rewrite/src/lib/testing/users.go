package testlib

import (
	"context"
	"fmt"
	. "github.com/onsi/gomega"
	dynamolib "github.com/veedubyou/chord-paper-be/go-rewrite/src/lib/dynamo"
	"github.com/veedubyou/chord-paper-be/go-rewrite/src/lib/errors/handle"
	userentity "github.com/veedubyou/chord-paper-be/go-rewrite/src/user/entity"
	"github.com/veedubyou/chord-paper-be/go-rewrite/src/user/google_id"
	userstorage "github.com/veedubyou/chord-paper-be/go-rewrite/src/user/storage"
)

var (
	// in the system, is Google validated, and owner of songs and tracklists
	PrimaryUser = User{
		ID:    "primary-user-id",
		Name:  "Primary User Name",
		Email: "primary@chordpaper.com",
	}

	// in the system, is Google validated, but not owner of songs and tracklists
	OtherUser = User{
		ID:    "other-user-id",
		Name:  "Other User Name",
		Email: "other@chordpaper.com",
	}

	// is Google validated but not in the system
	NoAccountUser = User{
		ID:    "not-in-db-id",
		Name:  "Not In DB User",
		Email: "adude@someoneelse.com",
	}

	// not Google validated, also not in the system
	GoogleUnauthorizedUser = User{
		ID:    "google-unauthorized-user-id",
		Name:  "Google Unauthorized User",
		Email: "rando@notpaper.com",
	}
)

func TokenForUserID(userID string) string {
	return fmt.Sprintf("%s-token", userID)
}

var _ google_id.Validator = TestingValidator{}

type TestingValidator struct{}

func (t TestingValidator) ValidateToken(ctx context.Context, requestToken string) (userentity.User, error) {
	validatedUsers := []User{PrimaryUser, OtherUser, NoAccountUser}

	for _, validatedUser := range validatedUsers {
		if requestToken == TokenForUserID(validatedUser.ID) {
			return userentity.User{
				ID:    validatedUser.ID,
				Name:  validatedUser.Name,
				Email: validatedUser.Email,
			}, nil
		}
	}

	return userentity.User{}, handle.Message(google_id.NotValidatedMark, "User is not validated")
}

func EnsureUsers(db dynamolib.DynamoDBWrapper) {
	EnsureUser(db, PrimaryUser)
	EnsureUser(db, OtherUser)
}

func EnsureUser(db dynamolib.DynamoDBWrapper, u User) {
	err := db.Table(userstorage.UsersTable).Put(u).Run()
	Expect(err).NotTo(HaveOccurred())
}
