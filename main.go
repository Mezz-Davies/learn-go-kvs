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
		log.Println("No id provided.")
		http.Error(w, fmt.Sprintf("Please provide an ID"), http.StatusBadRequest)
		return
	}
	if isValid, validationError := kvs.IdIsValid(id); !isValid {
		log.Println("Invalid id provided.")
		http.Error(w, fmt.Sprintf("ID format error: %s", validationError.Error()), http.StatusBadRequest)
		return
	}
	log.Printf("%v Request for id %v\n", req.Method, id)
	switch req.Method {
	case "GET":
		val, err := kvs.Get(id)
		if err != nil {
			log.Printf("GET Error %v\n", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if val == nil {
			log.Printf("GET 404: Resource not found.\n")
			http.Error(w, "Requested resource does not exist.", http.StatusNotFound)
			return
		}
		jsonResult, err := json.Marshal(val)
		if err != nil {
			log.Printf("GET JSON formatting Error %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResult)
	case "PUT":
		var v ParsedBody
		err := json.NewDecoder(req.Body).Decode(&v)
		if err != nil {
			log.Printf("PUT JSON decoding Error %v\n", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err = kvs.Update(id, v)
		if err != nil {
			log.Printf("PUT Error %v\n", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	case "DELETE":
		err := kvs.Delete(id)
		if err != nil {
			log.Printf("DELETE Error %v\n", err)
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
	log.Printf("%v Request\n", req.Method)
	switch req.Method {
	case "POST":
		err := json.NewDecoder(req.Body).Decode(&v)
		if err != nil {
			log.Printf("POST JSON decoding error %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		id, err := kvs.Set(v)
		if err != nil {
			log.Printf("POST Error %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		rMap["id"] = id
		jsonResult, err := json.Marshal(rMap)
		if err != nil {
			log.Printf("POST JSON encoding error %v", err)
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
