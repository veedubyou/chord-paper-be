package trackstorage

import (
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/guregu/dynamo"
	"github.com/pkg/errors"
	dynamolib "github.com/veedubyou/chord-paper-be/go-rewrite/src/lib/dynamo"
)

const (
	idKey = "song_id"
)

var _ dynamo.ItemUnmarshaler = &dbTrackList{}

type dbTrackList map[string]interface{}

func (d *dbTrackList) UnmarshalDynamoItem(dynamoItem map[string]*dynamodb.AttributeValue) error {
	if err := dynamolib.ValidateStringField(dynamoItem, idKey); err != nil {
		return errors.Wrap(err, "Failed to validate id field")
	}

	plainMap := map[string]interface{}{}
	err := dynamo.UnmarshalItem(dynamoItem, &plainMap)
	if err != nil {
		return errors.Wrap(err, "Failed to unmarshal dynamo item")
	}

	*d = plainMap

	return nil
}
