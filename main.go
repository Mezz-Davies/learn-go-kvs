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
	fmt.Println("ID Request received!")
	var v ParsedBody
	switch req.Method {
	case "GET":
		id := strings.TrimPrefix(req.URL.Path, "/kvs/")
		val := server.Get(id)
		jsonResult, err := json.Marshal(val)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Write(jsonResult)
	case "PUT":
		id := strings.TrimPrefix(req.URL.Path, "/kvs/")
		err := json.NewDecoder(req.Body).Decode(&v)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err = server.Update(id, v)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

	case "DELETE":
		id := strings.TrimPrefix(req.URL.Path, "/kvs/")
		err := server.Update(id, v)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	default:
		http.Error(w, "Method not supported with /:id", http.StatusBadRequest)
		return
	}
}

func responseHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Println("Request received!")
	var v ParsedBody
	rMap := make(map[string]interface{})
	switch req.Method {
	case "POST":
		err := json.NewDecoder(req.Body).Decode(&v)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		id := server.Set(v)
		rMap["id"] = id
		jsonResult, err := json.Marshal(rMap)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Write(jsonResult)
	default:
		http.Error(w, "Method not supported without /:id", http.StatusBadRequest)
		return
	}
}
func main() {
	server = kvs.Start()
	wg.Add(3)

	http.HandleFunc("/kvs", responseHandler)
	http.HandleFunc("/kvs/", idResponseHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
