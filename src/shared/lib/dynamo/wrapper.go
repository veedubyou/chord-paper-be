package dynamolib

import "github.com/guregu/dynamo"

import (
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

var encoder = dynamodbattribute.NewEncoder(func(e *dynamodbattribute.Encoder) {
	e.MarshalOptions.EnableEmptyCollections = true
	e.NullEmptyString = false
	e.NullEmptyByteSlice = false
})

type putMap map[string]any

func (p putMap) MarshalDynamo() (*dynamodb.AttributeValue, error) {
	var fields map[string]any = p
	return encoder.Encode(fields)
}

func NewDynamoDBWrapper(db *dynamo.DB) DynamoDBWrapper {
	return DynamoDBWrapper{DB: db}
}

type DynamoDBWrapper struct {
	*dynamo.DB
}

type DynamoTableWrapper struct {
	dynamo.Table
}

type DynamoUpdateWrapper struct {
	*dynamo.Update
}

func (d DynamoDBWrapper) Table(tableName string) DynamoTableWrapper {
	return DynamoTableWrapper{
		Table: d.DB.Table(tableName),
	}
}

func (d DynamoTableWrapper) Put(input map[string]any) *dynamo.Put {
	return d.Table.Put(putMap(input))
}

func (d DynamoTableWrapper) Update(hashKey string, value any) DynamoUpdateWrapper {
	return DynamoUpdateWrapper{
		Update: d.Table.Update(hashKey, value),
	}
}

func (d DynamoUpdateWrapper) Set(path string, value map[string]any) DynamoUpdateWrapper {
	return DynamoUpdateWrapper{
		Update: d.Update.Set(path, putMap(value)),
	}
}
