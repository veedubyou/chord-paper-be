package userstorage

import (
	"context"
	"github.com/cockroachdb/errors/markers"
	"github.com/guregu/dynamo"
	"github.com/veedubyou/chord-paper-be/src/server/internal/user/entity"
	"github.com/veedubyou/chord-paper-be/src/shared/lib/dynamo"
	"github.com/veedubyou/chord-paper-be/src/shared/lib/errors/mark"
)

const (
	UsersTable = "Users"
)

type DB struct {
	dynamoDB dynamolib.DynamoDBWrapper
}

func NewDB(dynamoDB dynamolib.DynamoDBWrapper) DB {
	return DB{
		dynamoDB: dynamoDB,
	}
}

func (d DB) GetUser(ctx context.Context, userID string) (userentity.User, error) {
	value := dbUser{}
	err := d.dynamoDB.Table(UsersTable).
		Get(idKey, userID).
		Consistent(true).
		OneWithContext(ctx, &value)

	if err != nil {
		switch {
		case markers.Is(err, dynamo.ErrNotFound):
			return userentity.User{}, mark.Wrap(err, UserNotFoundMark, "User is not found")
		default:
			return userentity.User{}, mark.Wrap(err, DefaultErrorMark, "Failed to fetch user")
		}
	}

	return userentity.User{
		ID:       value.ID,
		Name:     value.Name,
		Email:    value.Email,
		Verified: value.Verified,
	}, nil
}

func (d DB) SetUser(ctx context.Context, user userentity.User) error {
	value := dbUser{
		ID:       user.ID,
		Name:     user.Name,
		Email:    user.Email,
		Verified: user.Verified,
	}

	err := d.dynamoDB.Table(UsersTable).Table.Put(value).RunWithContext(ctx)

	if err != nil {
		return mark.Wrap(err, DefaultErrorMark, "Failed to add user")
	}

	return nil
}
