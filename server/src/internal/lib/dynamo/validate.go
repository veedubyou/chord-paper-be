package dynamolib

import (
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/pkg/errors"
)

func ValidateStringField(dynamoItem map[string]*dynamodb.AttributeValue, key string) error {
	value, ok := dynamoItem[key]
	if !ok {
		return errors.Errorf("No %s key was found", key)
	}

	if value.S == nil {
		return errors.Errorf("%s key is not in expected string format", key)
	}

	return nil
}
