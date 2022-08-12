package userstorage

import (
	"context"
	"github.com/cockroachdb/errors/markers"
	"github.com/guregu/dynamo"
	"github.com/veedubyou/chord-paper-be/src/server/internal/lib/dynamo"
	"github.com/veedubyou/chord-paper-be/src/server/internal/lib/errors/mark"
	"github.com/veedubyou/chord-paper-be/src/server/internal/user/entity"
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
		ID:    value.ID,
		Name:  value.Name,
		Email: value.Email,
	}, nil
}