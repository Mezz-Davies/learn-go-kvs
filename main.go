package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"gokvs/kvs"
)

var wg sync.WaitGroup
var server kvs.Server

type ParsedBody struct {
	Value interface{} `json:"value"`
}

func idResponseHandler(w http.ResponseWriter, req *http.Request) {
	id := strings.TrimPrefix(req.URL.Path, "/kvs/")
	if len(id) == 0 {
		http.Error(w, fmt.Sprintf("Please provide an ID"), http.StatusBadRequest)
		return
	}
	if isValid, validationError := kvs.IdIsValid(id); !isValid {
		http.Error(w, fmt.Sprintf("ID format error: %s", validationError.Error()), http.StatusBadRequest)
		return
	}
	var v ParsedBody
	switch req.Method {
	case "GET":
		val, err := kvs.Get(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		fmt.Println(val, "returned from kvs")
		if val == nil {
			http.Error(w, "Requested resource does not exist.", http.StatusBadRequest)
			return
		}
		jsonResult, err := json.Marshal(val)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResult)
	case "PUT":
		err := json.NewDecoder(req.Body).Decode(&v)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err = kvs.Update(id, v)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	case "DELETE":
		err := kvs.Delete(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	default:
		http.Error(w, "Method not supported with /:id", http.StatusBadRequest)
		return
	}
}

func responseHandler(w http.ResponseWriter, req *http.Request) {
	var v ParsedBody
	rMap := make(map[string]interface{})
	switch req.Method {
	case "POST":
		err := json.NewDecoder(req.Body).Decode(&v)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		id, err := kvs.Set(v)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		rMap["id"] = id
		jsonResult, err := json.Marshal(rMap)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResult)
	default:
		http.Error(w, "Method not supported without /:id", http.StatusBadRequest)
		return
	}
}
func main() {
	kvs.Start()
	defer kvs.Stop()

	// Manage calls to /kvs
	http.HandleFunc("/kvs", responseHandler)

	// Manage calls to /kvs/:id
	http.HandleFunc("/kvs/", idResponseHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
