package main

import (
	"fmt"

	uuid "github.com/google/uuid"
)

var kvs map[uuid.UUID]interface{}

func init() {
	kvs = make(map[uuid.UUID]interface{})
}
func addToKvs(value interface{}) string {
	key := uuid.New()

	kvs[key] = value

	return key.String()
}

func deleteFromKvs(keyToDelete string) error {
	uuidToDelete, parseError := uuid.Parse(keyToDelete)
	if parseError != nil {
		return parseError
	}
	if _, ok := kvs[uuidToDelete]; ok {
		delete(kvs, uuidToDelete)
	}
	return nil
}

func main() {
	fmt.Println("Storing values in KVS")
	key1 := addToKvs("Value 1")
	key2 := addToKvs(2)
	key3 := addToKvs(0xAB)

	fmt.Println(key1, key2, key3)

	fmt.Println(kvs)

	fmt.Printf("Deleting %s\n", key1)

	deleteError := deleteFromKvs(key1)
	if deleteError != nil {
		fmt.Printf("Error while deleting: %s\n", deleteError)
	}

	fmt.Println(kvs)
}
