package swarmipc

import (
	"encoding/json"
	"log"
	"net"
	"os"
)

type IpcServer struct {
	myName    string
	server    *net.UDPConn
	Callbacks map[string]func([]byte)
}

func NewIpcServer(port int) (server *IpcServer, err error) {
	server = &IpcServer{
		Callbacks: make(map[string]func([]byte)),
	}
	server.server, err = net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.ParseIP("0.0.0.0"),
		Port: port,
	})
	if err != nil {
		return nil, err
	}
	server.myName, err = os.Hostname()
	go server.readLoop()
	return server, err
}

func (s *IpcServer) AddCallback(method string, callback func(message []byte)) {
	s.Callbacks[method] = callback
}

func (s *IpcServer) readLoop() {
	for {
		buffer := make([]byte, 1024)
		length, source, err := s.server.ReadFromUDP(buffer)
		if err != nil {
			log.Println(err)
			continue
		}
		message := buffer[:length]
		log.Printf("[%s] Received %s from %s", s.myName, string(message), source.String())
		var ipcCall ipcMessage
		err = json.Unmarshal(message, &ipcCall)
		if err != nil {
			log.Printf("error decoding call: %s", err.Error())
		}
		if callback, found := s.Callbacks[ipcCall.Method]; found {
			go callback(ipcCall.Message)
		}
	}
}
