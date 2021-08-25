package kvsLogger

import (
	"log"
	"os"
)

type LogChannelType chan string

var LogChannel LogChannelType
var customLog *log.Logger

func StartLogger() chan string {
	customLog = log.New(os.Stdout, "", 0)
	LogChannel = make(LogChannelType)
	go processLogChannelEntries()
	return LogChannel
}

func processLogChannelEntries() {
	select {
	case msg := <-LogChannel:
		customLog.Println(msg)
	}
}

func Log(msg string) {
	log.Printf("Log : %v\n", msg)
	//LogChannel <- msg
}

func Error(msg string) {
	log.Printf("Error : %v\n", msg)
	//LogChannel <- msg
}

func Panic(msg string) {
	log.Printf("Panic : %v\n", msg)
	//LogChannel <- msg
}

func Fatal(msg string) {
	log.Printf("Fatal : %v\n", msg)
	//LogChannel <- msg
}
