package main

import (
	"encoding/json"
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

func responseHandler(w http.ResponseWriter, req *http.Request) {
	var v ParsedBody
	rMap := make(map[string]interface{})
	switch req.Method {
	case "GET":
		id := strings.TrimPrefix(req.URL.Path, "/")
		val := server.Get(id)
		rMap["value"] = val
		jsonResult, err := json.Marshal(rMap)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Write(jsonResult)
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
	case "PUT":
		id := strings.TrimPrefix(req.URL.Path, "/")
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
		id := strings.TrimPrefix(req.URL.Path, "/")
		err := server.Update(id, v)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

}
func main() {
	server = kvs.Start()
	wg.Add(3)

	http.HandleFunc("/", responseHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
