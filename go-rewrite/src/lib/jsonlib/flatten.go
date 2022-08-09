package jsonlib

import (
	"encoding/json"
	"github.com/cockroachdb/errors"
)

// T should be a struct or map[string]
type Flatten[T any] struct {
	Defined T
	Extra   map[string]interface{}
}

func (f Flatten[T]) MarshalJSON() ([]byte, error) {
	outputMap := map[string]interface{}{}

	for k, v := range f.Extra {
		outputMap[k] = v
	}

	definedFieldsMap, err := StructToMap(f.Defined)
	if err != nil {
		return nil, errors.Wrap(err, "Could not convert defined fields into a map")
	}

	for k, v := range definedFieldsMap {
		outputMap[k] = v
	}

	return json.Marshal(outputMap)
}

func (f *Flatten[T]) UnmarshalJSON(b []byte) error {
	definedFieldsObj := *new(T)
	err := json.Unmarshal(b, &definedFieldsObj)
	if err != nil {
		return errors.Wrap(err, "Could not unmarshal json data into defined fields")
	}

	definedFieldsMap, err := StructToMap(definedFieldsObj)
	if err != nil {
		return errors.Wrap(err, "Could not convert defined fields to a map")
	}

	objectMap := map[string]interface{}{}
	err = json.Unmarshal(b, &objectMap)
	if err != nil {
		return errors.Wrap(err, "Could not unmarshal json data into a map")
	}

	extras := map[string]interface{}{}
	for k, v := range objectMap {
		if _, ok := definedFieldsMap[k]; !ok {
			extras[k] = v
		}
	}

	*f = Flatten[T]{
		Defined: definedFieldsObj,
		Extra:   extras,
	}

	return nil
}

func (f Flatten[T]) ToMap() (map[string]interface{}, error) {
	return StructToMap(f)
}

func (f *Flatten[T]) FromMap(m map[string]interface{}) error {
	jsonBytes, err := json.Marshal(m)
	if err != nil {
		return errors.Wrap(err, "Could not marshal map")
	}

	newObj := Flatten[T]{}
	err = json.Unmarshal(jsonBytes, &newObj)
	if err != nil {
		return errors.Wrap(err, "Could not unmarshal json map to object")
	}

	*f = newObj
	return nil
}

func StructToMap(s interface{}) (map[string]interface{}, error) {
	jsonBytes, err := json.Marshal(s)
	if err != nil {
		return nil, errors.Wrap(err, "Could not marshal struct")
	}

	fieldsMap := map[string]interface{}{}
	err = json.Unmarshal(jsonBytes, &fieldsMap)
	if err != nil {
		return nil, errors.Wrap(err, "Could not unmarshal struct into a map")
	}

	return fieldsMap, nil
}
