package testing

import (
	"context"
	"fmt"
	. "github.com/onsi/gomega"
	google_id "github.com/veedubyou/chord-paper-be/src/server/google_id"
	dynamolib "github.com/veedubyou/chord-paper-be/src/shared/lib/dynamo"
	"github.com/veedubyou/chord-paper-be/src/shared/lib/errors/mark"
)

var (
	// in the system, is Google validated, and owner of songs and tracklists
	PrimaryUser = User{
		ID:       "primary-user-id",
		Name:     "Primary User Name",
		Email:    "primary@chordpaper.com",
		Verified: true,
	}

	// in the system, is Google validated, but not owner of songs and tracklists
	OtherUser = User{
		ID:       "other-user-id",
		Name:     "Other User Name",
		Email:    "other@chordpaper.com",
		Verified: true,
	}

	// is Google validated, but not verified and should not have access
	// also not saved to the DB
	UnverifiedUserNotInDB = User{
		ID:       "unverified-user-id",
		Name:     "Unverified User Name",
		Email:    "unverified@chordpaper.com",
		Verified: false,
	}

	// is Google validated, but not verified and should not have access
	// but saved to the DB
	UnverifiedUserInDB = User{
		ID:       "unverified-user-id-in-db",
		Name:     "Unverified User Name in DB",
		Email:    "unverified-in-db@chordpaper.com",
		Verified: false,
	}

	// is Google validated but not in the system
	NoAccountUser = User{
		ID:       "not-in-db-id",
		Name:     "Not In DB User",
		Email:    "adude@someoneelse.com",
		Verified: false,
	}

	// not Google validated, also not in the system
	GoogleUnauthorizedUser = User{
		ID:       "google-unauthorized-user-id",
		Name:     "Google Unauthorized User",
		Email:    "rando@notpaper.com",
		Verified: false,
	}
)

func TokenForUserID(userID string) string {
	return fmt.Sprintf("%s-token", userID)
}

var _ google_id.Validator = Validator{}

type Validator struct{}

func (t Validator) ValidateToken(ctx context.Context, requestToken string) (google_id.User, error) {
	validatedUsers := []User{PrimaryUser, OtherUser, UnverifiedUserNotInDB, NoAccountUser}

	for _, validatedUser := range validatedUsers {
		if requestToken == TokenForUserID(validatedUser.ID) {
			return google_id.User{
				GoogleID: validatedUser.ID,
				Name:     validatedUser.Name,
				Email:    validatedUser.Email,
			}, nil
		}
	}

	return google_id.User{}, mark.Message(google_id.NotValidatedMark, "User is not validated")
}

func EnsureUsers(db dynamolib.DynamoDBWrapper) {
	EnsureUser(db, PrimaryUser)
	EnsureUser(db, OtherUser)
	EnsureUser(db, UnverifiedUserInDB)
}

func EnsureUser(db dynamolib.DynamoDBWrapper, u User) {
	err := db.Table(UsersTable).Table.Put(u).Run()
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
}
