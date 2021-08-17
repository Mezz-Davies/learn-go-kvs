package kvs

import (
	"log"

	uuid "github.com/google/uuid"
)

type actionType int

const (
	getActionType = iota
	setActionType
	deleteActionType
)

type Action struct {
	actionType   actionType
	id           string
	val          interface{}
	replyChannel chan interface{}
}

var actionChannel chan Action
var kvs map[uuid.UUID]interface{}

func Start() chan<- Action {
	kvs = make(map[uuid.UUID]interface{})
	actionChannel = make(chan Action)

	go func() {
		for action := range actionChannel {
			switch action.actionType {
			case getActionType:
				if val, err := fetchFromKvs(action.id); err != nil {
					action.replyChannel <- val
				} else {
					action.replyChannel <- err
				}
			case setActionType:
				idToReturn := addToKvs(action.val)
				action.replyChannel <- idToReturn
			case deleteActionType:
				deleteResult := deleteFromKvs(action.id)
				action.replyChannel <- deleteResult
			default:
				log.Fatal("Unknown action type", action.actionType)
			}

		}
	}()

	return actionChannel
}
func addToKvs(value interface{}) string {
	key := uuid.New()

	kvs[key] = value

	return key.String()
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

func fetchFromKvs(ketToFetch string) (interface{}, error) {
	uuidToFetch, parseError := uuid.Parse(ketToFetch)
	if parseError != nil {
		return nil, parseError
	}
	if v, ok := kvs[uuidToFetch]; ok {
		return v, nil
	}
	return nil, nil
}

func Get(c chan<- Action, id string) interface{} {
	reply := make(chan interface{})

	action := Action{
		actionType:   getActionType,
		id:           id,
		val:          nil,
		replyChannel: reply,
	}

	c <- action
	val := <-reply
	return val
}

func Set(c chan<- Action, value interface{}) string {
	reply := make(chan interface{})

	action := Action{
		actionType:   setActionType,
		id:           "",
		val:          value,
		replyChannel: reply,
	}

	c <- action
	id := <-reply
	return id.(string)
}

func Delete(c chan<- Action, id string) error {
	reply := make(chan interface{})

	action := Action{
		actionType:   deleteActionType,
		id:           id,
		val:          nil,
		replyChannel: reply,
	}

	c <- action
	if err := <-reply; err != nil {
		return err.(error)
	}
	return nil
}
