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

type putMap map[string]interface{}

func (p putMap) MarshalDynamo() (*dynamodb.AttributeValue, error) {
	var fields map[string]interface{} = p
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

func (d DynamoDBWrapper) Table(tableName string) DynamoTableWrapper {
	return DynamoTableWrapper{
		Table: d.DB.Table(tableName),
	}
}

func (d DynamoTableWrapper) Put(input interface{}) *dynamo.Put {
	if m, ok := input.(map[string]interface{}); ok {
		return d.Table.Put(putMap(m))
	}

	return d.Table.Put(input)
}
