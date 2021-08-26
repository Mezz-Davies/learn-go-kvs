package kvsTcpServer

import (
	"context"
	"encoding/json"
	"fmt"
	"gokvs/kvs"
	"gokvs/kvsLogger"
	"io"
	"log"
	"net"
	"strings"
	"sync"
)

type Operation struct {
	Operation string      `json:"op"`
	Value     interface{} `json:"val"`
	Id        string      `json:"id"`
	RequestId string      `json:"reqId"`
}

type Response struct {
	RequestId string      `json:"reqId"`
	Response  interface{} `json:"res"`
	Success   bool        `json:"success"`
}

var shuttingDown bool

func processOperation(op Operation) (interface{}, error) {
	switch op.Operation {
	case "STORE":
		return kvs.Set(op.Value)
	case "FETCH":
		return kvs.Get(op.Id)
	case "UPDATE":
		return nil, kvs.Update(op.Id, op.Value)
	case "DELETE":
		return nil, kvs.Delete(op.Id)
	default:
		return nil, fmt.Errorf("Invalid operation")
	}
}

func filterEmptyStrings(s []string) []string {
	n := 0
	for _, val := range s {
		trimmedVal := strings.TrimSpace(val)
		if trimmedVal != "" {
			s[n] = trimmedVal
			n++
		}
	}
	return s[:n]
}

func separateOperations(buf []byte) []Operation {
	opStrings := strings.Split(string(buf), "\n")
	opStrings = filterEmptyStrings(opStrings)

	opSlice := []Operation{}
	for _, opString := range opStrings {
		var operation Operation
		err := json.Unmarshal([]byte(opString), &operation)
		if err != nil {
			kvsLogger.Error(fmt.Sprintf("Operation decoding error : %v", err.Error()))
		} else {
			opSlice = append(opSlice, operation)
		}
	}
	return opSlice
}

/*
 *	Messages expected to be JSON objects with the following fields;
 *		reqId 	- for the client to be able to link requests and response
 *		op		- One of the following strings: "STORE", "GET", "UPDATE", "DELETE", "STOP"
 *		val		- Value to be stored (if relevant)
 *		id		- Id to be operated on (if relevant)
 *	Input must be delimited by a newline char ('\n')
 *	Responses will be delimieted by newline char ('\n)
 */
func handleConnection(wg *sync.WaitGroup, conn net.Conn) {
	defer conn.Close()
	buffer := make([]byte, 1024)
	fmt.Println("New connection from : ", conn.LocalAddr())
	wg.Add(1)
	for {
		n, err := conn.Read(buffer)
		if err != nil && err != io.EOF {
			log.Println("Read err", err)
		}
		if n == 0 {
			return
		}

		receivedOperations := separateOperations(buffer[:n])
		for _, operation := range receivedOperations {

			if operation.Operation == "STOP" {
				return
			}
			responseObject := Response{
				RequestId: operation.RequestId,
				Response:  nil,
				Success:   false,
			}
			valToReturn, err := processOperation(operation)
			if err != nil {
				responseObject.Response = err.Error()
			} else {
				responseObject.Response = valToReturn
				responseObject.Success = true
			}
			jsonResponse, err := json.Marshal(responseObject)

			if err != nil {
				kvsLogger.Error(fmt.Sprintf("TCP encoding error %v", err.Error()))
				_, writeError := conn.Write([]byte("Error encoding response.\n"))
				if writeError != nil {
					kvsLogger.Error(fmt.Sprintf("Write error %v", err.Error()))
				}
			} else {
				response := append(jsonResponse, []byte("\n")...)
				_, writeError := conn.Write(response)
				if writeError != nil {
					kvsLogger.Error(fmt.Sprintf("Write error %v", err.Error()))
				}
			}
		}

		if shuttingDown {
			wg.Done()
			return
		}
	}
}

func StartTcpServer(root context.Context, portNumber int, doneChannel chan<- bool) {
	var wg sync.WaitGroup
	PORT := fmt.Sprintf(":%d", portNumber)
	listener, err := net.Listen("tcp4", PORT)
	if err != nil {
		log.Panic(err)
		return
	}
	shuttingDown = false

	kvsLogger.Log(fmt.Sprintf("TCP listening on port %s", PORT))
	go func() {
		for {
			connection, err := listener.Accept()
			if err != nil {
				log.Panic(err)
				return
			}
			go handleConnection(&wg, connection)
		}
	}()

	<-root.Done()
	kvsLogger.Log("Closing TCP connection")
	shuttingDown = true

	wg.Wait()

	doneChannel <- true
}
