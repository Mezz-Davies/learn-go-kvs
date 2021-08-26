package kvsLogger

import (
	"fmt"
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
	for msg := range LogChannel {
		log.Println(msg)
	}
}

func Log(msg string) {
	LogChannel <- fmt.Sprintf("Log : %v", msg)
}

func Error(msg string) {
	LogChannel <- fmt.Sprintf("Error : %v", msg)
}

func Panic(msg string) {
	LogChannel <- fmt.Sprintf("Panic : %v", msg)
}

func Fatal(msg string) {
	LogChannel <- fmt.Sprintf("Fatal : %v", msg)
}
