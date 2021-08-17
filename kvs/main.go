package kvs

import (
	uuid "github.com/google/uuid"
)

var kvs map[uuid.UUID]interface{}

func init() {
	kvs = make(map[uuid.UUID]interface{})
}
func Add(value interface{}) string {
	key := uuid.New()

	kvs[key] = value

	return key.String()
}

func Delete(keyToDelete string) error {
	uuidToDelete, parseError := uuid.Parse(keyToDelete)
	if parseError != nil {
		return parseError
	}
	if _, ok := kvs[uuidToDelete]; ok {
		delete(kvs, uuidToDelete)
	}
	return nil
}

func Fetch(ketToFetch string) (interface{}, error) {
	uuidToFetch, parseError := uuid.Parse(ketToFetch)
	if parseError != nil {
		return nil, parseError
	}
	if v, ok := kvs[uuidToFetch]; ok {
		return v, nil
	}
	return nil, nil
}
