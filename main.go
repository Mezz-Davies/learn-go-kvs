package main

import (
	"context"
	"fmt"
	"gokvs/kvs"
	"gokvs/kvsHttpServer"
	"gokvs/kvsLogger"
	"gokvs/kvsTcpServer"
	"os"
	"os/signal"
)

func main() {
	kvs.Start()
	defer kvs.Stop()

	kvsLogger.StartLogger()

	rootContext, cancel := context.WithCancel(context.Background())

	doneChannel := make(chan bool)

	go kvsHttpServer.StartHttpServer(rootContext, 8080, doneChannel)

	go kvsTcpServer.StartTcpServer(rootContext, 8081, doneChannel)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	defer signal.Stop(c)

	interrupt := <-c
	kvsLogger.Log(fmt.Sprintf("Signal '%v' received. Stopping child processes", interrupt))
	cancel()

	for i := 0; i < 2; i++ {
		<-doneChannel
	}
	kvsLogger.Log("Main: Exited")
}
