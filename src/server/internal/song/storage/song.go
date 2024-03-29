package songstorage

import (
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/guregu/dynamo"
	"github.com/veedubyou/chord-paper-be/src/shared/lib/dynamo"
	"github.com/veedubyou/chord-paper-be/src/shared/lib/errors/mark"
)

const (
	idKey    = "id"
	ownerKey = "owner"
)

var _ dynamo.ItemUnmarshaler = &dbSong{}

type dbSong map[string]any

func (d *dbSong) UnmarshalDynamoItem(dynamoItem map[string]*dynamodb.AttributeValue) error {
	if err := dynamolib.ValidateStringField(dynamoItem, idKey); err != nil {
		return mark.Wrap(err, SongUnmarshalMark, "Failed to validate ID field")
	}

	if err := dynamolib.ValidateStringField(dynamoItem, ownerKey); err != nil {
		return mark.Wrap(err, SongUnmarshalMark, "Failed to validate owner field")
	}

	plainMap := map[string]any{}
	err := dynamo.UnmarshalItem(dynamoItem, &plainMap)
	if err != nil {
		return mark.Wrap(err, SongUnmarshalMark, "Failed to unmarshal dynamo item")
	}

	*d = plainMap

	return nil
}
