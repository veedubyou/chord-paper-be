package songstorage

import (
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/guregu/dynamo"
	"github.com/pkg/errors"
	z "github.com/veedubyou/chord-paper-be/go-rewrite/src/lib/errors/errlog"
)

const (
	idKey    = "id"
	ownerKey = "owner"
)

var _ dynamo.ItemUnmarshaler = &dbSong{}

type dbSong map[string]interface{}

func (d *dbSong) UnmarshalDynamoItem(dynamoItem map[string]*dynamodb.AttributeValue) error {
	if err := validateStringField(dynamoItem, idKey); z.Err(err) {
		return errors.Wrap(err, "Failed to validate id field")
	}

	if err := validateStringField(dynamoItem, ownerKey); z.Err(err) {
		return errors.Wrap(err, "Failed to validate owner field")
	}

	plainMap := map[string]interface{}{}
	err := dynamo.UnmarshalItem(dynamoItem, &plainMap)
	if z.Err(err) {
		return errors.Wrap(err, "Failed to unmarshal dynamo item")
	}

	*d = plainMap

	return nil
}

func validateStringField(dynamoItem map[string]*dynamodb.AttributeValue, key string) error {
	value, ok := dynamoItem[key]
	if !ok {
		return errors.Errorf("No %s key was found", key)
	}

	if value.S == nil {
		return errors.Errorf("%s key is not in expected string format", key)
	}

	return nil
}
