package kvsHttpServer

import (
	"context"
	"encoding/json"
	"fmt"
	"gokvs/kvs"
	"gokvs/kvsLogger"
	"log"
	"net/http"
	"strings"
	"time"
)

type ParsedBody struct {
	Value interface{} `json:"value"`
}

func getAndValidateIdInput(req *http.Request) (string, error) {
	id := strings.TrimPrefix(req.URL.Path, "/kvs/")
	if len(id) == 0 {
		return "", fmt.Errorf("No id provided")
	}
	if isValid, validationError := kvs.IdIsValid(id); !isValid {
		return "", fmt.Errorf("ID format error: %s", validationError.Error())
	}
	return id, nil
}

func handleIdGet(w http.ResponseWriter, id string) {
	val, err := kvs.Get(id)
	clientErrorMessage := fmt.Sprintf("Could not GET on id %v", id)
	if err != nil {
		kvsLogger.Log(fmt.Sprintf("GET Error %v\n", err))
		http.Error(w, clientErrorMessage, http.StatusBadRequest)
		return
	}
	if val == nil {
		kvsLogger.Log("GET 404: Resource not found.")
		http.Error(w, "Requested resource does not exist.", http.StatusNotFound)
		return
	}
	jsonResult, err := json.Marshal(val)
	if err != nil {
		kvsLogger.Log(fmt.Sprintf("GET JSON formatting Error %v", err))
		http.Error(w, clientErrorMessage, http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResult)
}

func handleIdPut(w http.ResponseWriter, req *http.Request, id string) {
	var v ParsedBody
	err := json.NewDecoder(req.Body).Decode(&v)
	clientErrorMessage := fmt.Sprintf("Could not PUT on id %v", id)
	if err != nil {
		kvsLogger.Log(fmt.Sprintf("PUT Body JSON decoding Error %v\n", err))
		http.Error(w, clientErrorMessage, http.StatusBadRequest)
		return
	}
	err = kvs.Update(id, v.Value)
	if err != nil {
		kvsLogger.Log(fmt.Sprintf("PUT Error %v\n", err))
		http.Error(w, clientErrorMessage, http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func handleIdDelete(w http.ResponseWriter, id string) {
	err := kvs.Delete(id)
	clientErrorMessage := fmt.Sprintf("Could not DELETE on id %v", id)
	if err != nil {
		kvsLogger.Log(fmt.Sprintf("DELETE Error %v\n", err))
		http.Error(w, clientErrorMessage, http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func idResponseHandler(w http.ResponseWriter, req *http.Request) {
	id, err := getAndValidateIdInput(req)
	if err != nil {
		errMessage := fmt.Sprintf("Id validation error %v", err.Error())
		kvsLogger.Log(errMessage)
		http.Error(w, errMessage, http.StatusBadRequest)
		return
	}
	kvsLogger.Log(fmt.Sprintf("%v Request for id %v\n", req.Method, id))
	switch req.Method {
	case "GET":
		handleIdGet(w, id)
	case "PUT":
		handleIdPut(w, req, id)
	case "DELETE":
		handleIdDelete(w, id)
	default:
		http.Error(w, "Method not supported with /:id", http.StatusBadRequest)
		return
	}
}

func handlePost(w http.ResponseWriter, req *http.Request) {

}

func responseHandler(w http.ResponseWriter, req *http.Request) {
	kvsLogger.Log(fmt.Sprintf("%v Request\n", req.Method))
	switch req.Method {
	case "POST":
		var v ParsedBody
		err := json.NewDecoder(req.Body).Decode(&v)
		if err != nil {
			kvsLogger.Log(fmt.Sprintf("POST JSON decoding error %v", err))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		id, err := kvs.Set(v.Value)
		if err != nil {
			kvsLogger.Log(fmt.Sprintf("POST Error %v", err))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		rMap := make(map[string]interface{})
		rMap["id"] = id
		jsonResult, err := json.Marshal(rMap)
		if err != nil {
			kvsLogger.Log(fmt.Sprintf("POST JSON encoding error %v", err))
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

func StartHttpServer(root context.Context, portNumber int, doneChannel chan<- bool) {
	kvsLogger.Log("HTTP Method!")

	mux := http.NewServeMux()
	mux.Handle("/kvs", http.HandlerFunc(responseHandler))
	mux.Handle("/kvs/", http.HandlerFunc(idResponseHandler))

	srv := &http.Server{
		Addr:    ":" + fmt.Sprintf("%d", portNumber),
		Handler: mux,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen:%+s\n", err)
		}
	}()

	kvsLogger.Log(fmt.Sprintf("HTTP Server started on port %d", portNumber))

	select {
	case <-root.Done():
		kvsLogger.Log(fmt.Sprintf("HTTP Server stopping..."))
		ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer func() {
			cancel()
		}()

		if err := srv.Shutdown(ctxShutDown); err != nil {
			log.Fatalf("HTTP Server Shutdown failed: %+s", err)
		}

		kvsLogger.Log(fmt.Sprintf("HTTP Server exited properly"))
	}

	doneChannel <- true
}
