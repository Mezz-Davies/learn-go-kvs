package kvs

import (
	"errors"
	"expvar"
	"log"

	uuid "github.com/google/uuid"
)

const (
	getActionType = iota
	setActionType
	updateActionType
	deleteActionType
)

type KvsStoreType map[uuid.UUID]interface{}

var kvs KvsStoreType

type actionType int

type Action struct {
	actionType actionType
	id         string
	val        interface{}
}

var actionChannel chan Action
var replyChannel chan interface{}
var resultsChannel chan kvsResult

type KvsMetricsStruct struct {
	Size                 int
	Operations           int
	SuccessfulOperations int
}

var kvsSize int
var kvsOps int
var kvsSuccessfulOps int

type kvsResult struct {
	actionType actionType
	success    bool
}

/*
 * Synchronous KVS Access methods
 */
func getFromKvs(ketToFetch string) (interface{}, error) {
	uuidToFetch, parseError := uuid.Parse(ketToFetch)
	if parseError != nil {
		return nil, parseError
	}
	if v, ok := kvs[uuidToFetch]; ok {
		return v, nil
	}
	return nil, nil
}
func setToKvs(value interface{}) string {
	key := uuid.New()

	kvs[key] = value

	return key.String()
}

func updateKvs(keyToUpdate string, value interface{}) error {
	uuidToUpdate, parseError := uuid.Parse(keyToUpdate)
	if parseError != nil {
		return parseError
	}
	kvs[uuidToUpdate] = value
	return nil
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

/*
 *	Channel monitor functions, to be run in own GoRoutines
 */
func monitorStoreOperations(storeActionChannel <-chan Action) {
	for action := range storeActionChannel {
		switch action.actionType {
		case getActionType:
			if val, err := getFromKvs(action.id); err == nil {
				replyChannel <- val
			} else {
				replyChannel <- err
			}
		case setActionType:
			idToReturn := setToKvs(action.val)
			replyChannel <- idToReturn
		case updateActionType:
			updateResult := updateKvs(action.id, action.val)
			replyChannel <- updateResult
		case deleteActionType:
			deleteResult := deleteFromKvs(action.id)
			replyChannel <- deleteResult
		default:
			log.Fatal("Unknown action type", action.actionType)
		}
	}
}

func monitorResultsChannel(resultsChannel <-chan kvsResult) {
	for result := range resultsChannel {
		kvsOps += 1
		if result.success == true {
			kvsSuccessfulOps += 1

			switch {
			case result.actionType == setActionType:
				kvsSize += 1
			case result.actionType == deleteActionType:
				kvsSize -= 1
			}
		}
	}
}

func registerResult(actionType actionType, success bool) {
	resultToRegister := kvsResult{
		actionType: actionType,
		success:    success,
	}
	resultsChannel <- resultToRegister
}

// Function to describe exported metrics.
func KvsMetrics() interface{} {
	return KvsMetricsStruct{
		Size:                 kvsSize,
		Operations:           kvsOps,
		SuccessfulOperations: kvsSuccessfulOps,
	}
}

/*
 *	Initialises kvs. Should only be called during main thread startup.
 *	Kvs is then ready to be used concurrently by calling Accessor methods below.
 */
func Start(initState ...KvsStoreType) {
	kvs = make(KvsStoreType)
	actionChannel = make(chan Action)
	replyChannel = make(chan interface{})
	resultsChannel = make(chan kvsResult)

	// Set initial state of store
	for _, state := range initState {
		for k, v := range state {
			kvs[k] = v
			kvsSize++
		}
	}

	// ExpVars
	expvar.Publish("Kvs Metrics", expvar.Func(KvsMetrics))

	// init channel monitoring
	go monitorStoreOperations(actionChannel)
	go monitorResultsChannel(resultsChannel)
}

func Stop() {
	close(actionChannel)
	close(replyChannel)
	close(resultsChannel)
}

func IdIsValid(id string) (bool, error) {
	_, parseError := uuid.Parse(id)
	if parseError != nil {
		return false, parseError
	}
	return true, nil
}

func Get(id string) (interface{}, error) {
	if id == "" {
		registerResult(getActionType, false)
		return "", errors.New("No id provided.")
	}
	action := Action{
		actionType: getActionType,
		id:         id,
		val:        nil,
	}
	actionChannel <- action
	val := <-replyChannel
	registerResult(getActionType, true)
	return val, nil
}

func Set(value interface{}) (string, error) {
	if value == nil {
		registerResult(setActionType, false)
		return "", errors.New("Nil value given. Value will not be stored.")
	}
	action := Action{
		actionType: setActionType,
		id:         "",
		val:        value,
	}
	actionChannel <- action
	id := <-replyChannel
	registerResult(setActionType, true)
	return id.(string), nil
}

func Update(id string, value interface{}) error {
	if id == "" {
		registerResult(updateActionType, false)
		return errors.New("No id provided.")
	}
	if value == nil {
		return errors.New("Nil value given. Value will not be stored.")
	}
	action := Action{
		actionType: updateActionType,
		id:         id,
		val:        value,
	}
	actionChannel <- action
	if err := <-replyChannel; err != nil {
		return err.(error)
	}
	registerResult(updateActionType, true)
	return nil
}

func Delete(id string) error {
	if id == "" {
		registerResult(deleteActionType, false)
		return errors.New("No id provided.")
	}
	action := Action{
		actionType: deleteActionType,
		id:         id,
		val:        nil,
	}

	actionChannel <- action
	if err := <-replyChannel; err != nil {
		return err.(error)
	}
	registerResult(deleteActionType, true)
	return nil
}

/*
 *	Exports a copy of the current KVS. This is used for testing
 */
func GetStoreCopy() KvsStoreType {
	return kvs
}
