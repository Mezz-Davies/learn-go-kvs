package kvs

import (
	"log"

	uuid "github.com/google/uuid"
)

type actionType int

const (
	getActionType = iota
	setActionType
	updateActionType
	deleteActionType
)

type Action struct {
	actionType   actionType
	id           string
	val          interface{}
	replyChannel chan interface{}
}

type Server struct {
	actionChannel chan Action
}

var actionChannel chan Action
var kvs map[uuid.UUID]interface{}

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

func Start() Server {
	kvs = make(map[uuid.UUID]interface{})
	actionChannel = make(chan Action)

	go func() {
		for action := range actionChannel {
			switch action.actionType {
			case getActionType:
				if val, err := getFromKvs(action.id); err == nil {
					action.replyChannel <- val
				} else {
					action.replyChannel <- err
				}
			case setActionType:
				idToReturn := setToKvs(action.val)
				action.replyChannel <- idToReturn
			case updateActionType:
				updateResult := updateKvs(action.id, action.val)
				action.replyChannel <- updateResult
			case deleteActionType:
				deleteResult := deleteFromKvs(action.id)
				action.replyChannel <- deleteResult
			default:
				log.Fatal("Unknown action type", action.actionType)
			}
		}
	}()

	server := Server{
		actionChannel: actionChannel,
	}
	return server
}

func (s *Server) Get(id string) interface{} {
	reply := make(chan interface{})

	action := Action{
		actionType:   getActionType,
		id:           id,
		val:          nil,
		replyChannel: reply,
	}

	s.actionChannel <- action
	val := <-reply
	return val
}

func (s *Server) Set(value interface{}) string {
	reply := make(chan interface{})

	action := Action{
		actionType:   setActionType,
		id:           "",
		val:          value,
		replyChannel: reply,
	}

	s.actionChannel <- action
	id := <-reply
	return id.(string)
}

func (s *Server) Update(id string, value interface{}) error {
	reply := make(chan interface{})

	action := Action{
		actionType:   updateActionType,
		id:           id,
		val:          value,
		replyChannel: reply,
	}

	s.actionChannel <- action
	if err := <-reply; err != nil {
		return err.(error)
	}
	return nil
}

func (s *Server) Delete(id string) error {
	reply := make(chan interface{})

	action := Action{
		actionType:   deleteActionType,
		id:           id,
		val:          nil,
		replyChannel: reply,
	}

	s.actionChannel <- action
	if err := <-reply; err != nil {
		return err.(error)
	}
	return nil
}
