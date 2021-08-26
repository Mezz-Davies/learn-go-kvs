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
	"sync"
)

func main() {
	var rootWg sync.WaitGroup

	kvs.Start()
	defer kvs.Stop()
	kvsLogger.StartLogger(&rootWg)
	defer kvsLogger.WaitForLoggerToComplete()

	rootContext, cancel := context.WithCancel(context.Background())

	go kvsHttpServer.StartHttpServer(rootContext, &rootWg, 8080)

	go kvsTcpServer.StartTcpServer(rootContext, &rootWg, 8081)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	defer signal.Stop(c)

	interrupt := <-c
	kvsLogger.Log(fmt.Sprintf("Signal '%v' received. Stopping child processes", interrupt))
	cancel()

	rootWg.Wait()

	kvsLogger.Log("Main: Exited")
}
