package main

import (
	"gokvs/kvs"
	"gokvs/kvsHttpServer"
)

func main() {
	kvs.Start()
	defer kvs.Stop()

	kvsHttpServer.StartHttpServer()
}
