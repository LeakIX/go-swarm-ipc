# Golang Swarm IPC

## Features

- Auto-discover services based on name
- Callback register and dispatch
- Non-blocking UDP

## Limitations

- UDP, expect loss
- Max 1024 bytes ( or your MTU )

## Example

```golang
package main

import (
	"github.com/LeakIX/go-swarm-ipc"
	"log"
	"time"
)

func main() {
	// start IPC server on UDP port 3000
	server, err := swarmipc.NewIpcServer(3000)
	if err != nil {
		panic(err)
	}
	// Add callback for method test
	server.AddCallback("test", func(message []byte) {
		log.Println(string(message))
	})
	// Create a client for task "test_task" ( your service name )
	client := swarmipc.NewIpcClient("test_task", 3000)
	time.Sleep(5 * time.Second)
	// Broadcast to all "test_task" services
	for {
		client.CallBroadcast("test", []byte("test"))
		time.Sleep(5 * time.Second)
	}
}

```
