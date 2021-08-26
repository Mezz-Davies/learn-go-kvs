package kvsLogger

import (
	"fmt"
	"log"
	"os"
	"sync"
)

type LogChannelType chan string

var LogChannel LogChannelType
var customLog *log.Logger
var wg sync.WaitGroup

func StartLogger(rootWg *sync.WaitGroup) chan string {
	customLog = log.New(os.Stdout, "", 0)
	LogChannel = make(LogChannelType)
	go processLogChannelEntries()
	return LogChannel
}

func WaitForLoggerToComplete() {
	wg.Wait()
}

func processLogChannelEntries() {
	defer close(LogChannel)
	for msg := range LogChannel {
		log.Println(msg)
		wg.Done()
	}
}

func Log(msg string) {
	wg.Add(1)
	LogChannel <- fmt.Sprintf("Log : %v", msg)
}

func Error(msg string) {
	wg.Add(1)
	LogChannel <- fmt.Sprintf("Error : %v", msg)
}

func Panic(msg string) {
	wg.Add(1)
	LogChannel <- fmt.Sprintf("Panic : %v", msg)
}

func Fatal(msg string) {
	wg.Add(1)
	LogChannel <- fmt.Sprintf("Fatal : %v", msg)
}
