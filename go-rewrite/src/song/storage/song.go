package songstorage

import (
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/guregu/dynamo"
	"github.com/pkg/errors"
	dynamolib "github.com/veedubyou/chord-paper-be/go-rewrite/src/lib/dynamo"
)

const (
	idKey    = "id"
	ownerKey = "owner"
)

var _ dynamo.ItemUnmarshaler = &dbSong{}

type dbSong map[string]interface{}

func (d *dbSong) UnmarshalDynamoItem(dynamoItem map[string]*dynamodb.AttributeValue) error {
	if err := dynamolib.ValidateStringField(dynamoItem, idKey); err != nil {
		return errors.Wrap(err, "Failed to validate id field")
	}

	if err := dynamolib.ValidateStringField(dynamoItem, ownerKey); err != nil {
		return errors.Wrap(err, "Failed to validate owner field")
	}

	plainMap := map[string]interface{}{}
	err := dynamo.UnmarshalItem(dynamoItem, &plainMap)
	if err != nil {
		return errors.Wrap(err, "Failed to unmarshal dynamo item")
	}

	*d = plainMap

	return nil
}
