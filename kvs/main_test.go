package kvs

import (
	"fmt"
	"testing"
	"time"

	uuid "github.com/google/uuid"
)

func TestGet(t *testing.T) {
	uuidToGet := uuid.New()
	expectedVal := 3.14
	initialTestState := make(map[uuid.UUID]interface{})
	initialTestState[uuidToGet] = expectedVal
	Start(initialTestState)
	defer Stop()

	v, err := Get(uuidToGet.String())
	if err != nil {
		t.Errorf("Get returned err %v\n", err)
	} else if v != expectedVal {
		t.Errorf("Expected %v got %v", expectedVal, v)
	}
}

func TestSet(t *testing.T) {
	Start()
	defer Stop()

	testValue := "This is a test"
	id, err := Set(testValue)
	if err != nil {
		t.Errorf("Set returned err %v\n", err)
	}
	idAsUUID, parseError := uuid.Parse(id)
	if parseError != nil {
		t.Errorf("Unparsable id returned, returns parse error %v", parseError)
	}
	if kvs[idAsUUID] != testValue {
		t.Errorf("Expected %v in kvs, got %v", testValue, kvs[idAsUUID])
	}
}

func TestUpdate(t *testing.T) {
	uuidToUpdate := uuid.New()
	expectedVal := 5
	initialTestState := make(map[uuid.UUID]interface{})
	initialTestState[uuidToUpdate] = 3.14
	Start(initialTestState)
	defer Stop()

	err := Update(uuidToUpdate.String(), expectedVal)
	if err != nil {
		t.Errorf("Update returned err %v\n", err)
	}
	if kvs[uuidToUpdate] != expectedVal {
		t.Errorf("Expected %v in kvs, got %v", expectedVal, kvs[uuidToUpdate])
	}
}

func TestDelete(t *testing.T) {
	uuidToDelete := uuid.New()
	initialVal := 5
	initialTestState := make(map[uuid.UUID]interface{})
	initialTestState[uuidToDelete] = initialVal
	Start(initialTestState)
	defer Stop()

	err := Delete(uuidToDelete.String())
	if err != nil {
		t.Errorf("Delete returned err %v\n", err)
	}
	if v, ok := kvs[uuidToDelete]; ok {
		t.Errorf("Expected value to be deleted, got %v", v)
	}
}
func BenchmarkKvs(b *testing.B) {
	Start()
	defer Stop()
	var id string

	var passCount = 0
	var testCount = 0
	resultsChan := make(chan bool)
	go func() {
		for result := range resultsChan {
			if result {
				passCount++
			}
			testCount++
		}
	}()
	for i := 0; i < b.N; i++ {
		val := fmt.Sprintf("%d test value", i)
		id, _ = Set(val)
		storedVal, err := Get(id)
		if err != nil {
			resultsChan <- false
		} else {
			resultsChan <- (val == storedVal)
		}
	}

	fmt.Printf("%d tests of %d passed\n", passCount, testCount)
}

func BenchmarkKvsParallel(b *testing.B) {
	Start()
	defer Stop()

	var passCount = 0
	var testCount = 0
	resultsChan := make(chan bool)
	go func() {
		for result := range resultsChan {
			if result {
				passCount++
			}
			testCount++
		}
	}()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			valToStore := time.Now().UnixNano()
			id, err := Set(valToStore)
			if err != nil {
				resultsChan <- false
			}

			valFromStore, err := Get(id)
			if err != nil {
				resultsChan <- false
			} else {
				resultsChan <- (valToStore == valFromStore)
			}
		}
	})

	fmt.Printf("%d tests of %d passed\n", passCount, testCount)
}
