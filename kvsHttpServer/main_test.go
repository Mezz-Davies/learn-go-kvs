package kvsHttpServer

import (
	"encoding/json"
	"fmt"
	"gokvs/kvs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	uuid "github.com/google/uuid"
)

func TestRequests(t *testing.T) {
	testStore := make(kvs.KvsStoreType)
	testValKey := uuid.New()
	testVal := "test 1 val"
	testStore[testValKey] = testVal

	kvs.Start(testStore)
	defer kvs.Stop()
	t.Run("Fetch val that exists in store", func(t *testing.T) {
		request := newGetIdRequest(testValKey.String())
		response := httptest.NewRecorder()

		idResponseHandler(response, request)
		fmt.Println(response.Body.String())
		assertResponseBody(t, response.Body.String(), `"test 1 val"`)
	})

	t.Run("Fetch val that does not exist in store", func(t *testing.T) {
		test2ValKey := uuid.New()
		request := newGetIdRequest(test2ValKey.String())
		response := httptest.NewRecorder()

		idResponseHandler(response, request)
		fmt.Println(response.Body.String())
		assertResponseBody(t, response.Body.String(), "Requested resource does not exist.\n")
	})

	t.Run("Add value to store", func(t *testing.T) {
		jsonToSend := `{"value": "test 2 value"}`
		request := newPostRequest(jsonToSend)
		response := httptest.NewRecorder()

		responseHandler(response, request)
		var idReturned map[string]string
		json.Unmarshal(response.Body.Bytes(), &idReturned)
		fmt.Println("Got id:", idReturned["id"])
		if ok, err := kvs.IdIsValid(idReturned["id"]); !ok {
			t.Errorf("Id response returned err %v", err)
		}
		store := kvs.GetStoreCopy()
		uuidToCheck, _ := uuid.Parse(idReturned["id"])
		if _, ok := store[uuidToCheck]; !ok {
			t.Errorf("Id not found in store")
		}
	})

	t.Run("Update value in store", func(t *testing.T) {
		jsonToSend := `{"value": "Updated Test Value"}`
		request := newUpdateRequest(testValKey.String(), jsonToSend)
		response := httptest.NewRecorder()

		idResponseHandler(response, request)

		if response.Code != http.StatusAccepted {
			t.Errorf("Response code is not accepted. Returned code %v", response.Code)
		}
		store := kvs.GetStoreCopy()
		if val, ok := store[testValKey]; !ok || val != "Updated Test Value" {
			t.Errorf("Id in store incorrect value")
		}
	})

	t.Run("Delete value in store", func(t *testing.T) {
		request := newDeleteRequest(testValKey.String())
		response := httptest.NewRecorder()

		idResponseHandler(response, request)

		if response.Code != http.StatusAccepted {
			t.Errorf("Response code is not accepted. Returned code %v", response.Code)
		}
		store := kvs.GetStoreCopy()
		if _, ok := store[testValKey]; ok {
			t.Errorf("Id in store not deleted")
		}
	})
}

func newDeleteRequest(idToUpdate string) *http.Request {
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/kvs/%s", idToUpdate), nil)
	req.Header.Set("Content-Type", "application/json")
	return req
}
func newUpdateRequest(idToUpdate, jsonStringToSend string) *http.Request {
	reader := strings.NewReader(jsonStringToSend)
	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/kvs/%s", idToUpdate), reader)
	req.Header.Set("Content-Type", "application/json")
	return req
}
func newPostRequest(jsonStringToSend string) *http.Request {
	reader := strings.NewReader(jsonStringToSend)
	req, _ := http.NewRequest(http.MethodPost, "/kvs", reader)
	req.Header.Set("Content-Type", "application/json")
	return req
}

func newGetIdRequest(id string) *http.Request {
	requestUrl := fmt.Sprintf("/kvs/%s", id)
	req, _ := http.NewRequest(http.MethodGet, requestUrl, nil)
	return req
}

func assertResponseBody(t testing.TB, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("response body is wrong, got %q want %q", got, want)
	}
}
