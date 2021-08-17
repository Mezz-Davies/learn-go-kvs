package main

import (
	"fmt"

	"gokvs/kvs"
)

func main() {
	fmt.Println("Storing values in KVS")
	key1 := kvs.Add("Value 1")
	key2 := kvs.Add(2)
	key3 := kvs.Add(0xAB)

	fmt.Println(key1, key2, key3)

	fmt.Printf("Deleting %s\n", key1)

	deleteError := kvs.Delete(key1)
	if deleteError != nil {
		fmt.Printf("Error while deleting: %s\n", deleteError)
	}

	val, err := kvs.Fetch(key2)
	if err != nil {
		fmt.Println("Fetch Error", err)
	}
	fmt.Println(val)
}
