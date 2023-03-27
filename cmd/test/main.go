package main

import (
	"github.com/LeakIX/go-swarm-ipc"
	"log"
	"time"
)

func main() {
	server, err := swarmipc.NewIpcServer(3000)
	if err != nil {
		panic(err)
	}
	server.AddCallback("test", func(message []byte) {
		log.Println(string(message))
	})
	client := swarmipc.NewIpcClient("test_task", 3000)
	time.Sleep(5 * time.Second)
	for {
		client.CallBroadcast("test", []byte("test"))
		time.Sleep(5 * time.Second)
	}
}
