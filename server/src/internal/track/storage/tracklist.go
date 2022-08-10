package trackstorage

import (
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/guregu/dynamo"
	dynamolib "github.com/veedubyou/chord-paper-be/server/src/lib/dynamo"
	"github.com/veedubyou/chord-paper-be/server/src/lib/errors/mark"
)

const (
	idKey = "song_id"
)

var _ dynamo.ItemUnmarshaler = &dbTrackList{}

type dbTrackList map[string]interface{}

func (d *dbTrackList) UnmarshalDynamoItem(dynamoItem map[string]*dynamodb.AttributeValue) error {
	if err := dynamolib.ValidateStringField(dynamoItem, idKey); err != nil {
		return mark.Wrap(err, UnmarshalMark, "Failed to validate id field")
	}

	plainMap := map[string]interface{}{}
	err := dynamo.UnmarshalItem(dynamoItem, &plainMap)
	if err != nil {
		return mark.Wrap(err, UnmarshalMark, "Failed to unmarshal dynamo item")
	}

	*d = plainMap

	return nil
}
