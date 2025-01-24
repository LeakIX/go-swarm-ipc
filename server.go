package swarmipc

import (
	"encoding/json"
	"net"
	"os"
)

// IpcServer is a listener for IPC calls in a Docker Swarm cluster
type IpcServer struct {
	myName    string
	server    *net.UDPConn
	Callbacks map[string]func([]byte)
}

// NewIpcServer creates and IPC server listening on the specified UDP port
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

// AddCallback adds a method to be called when an IPC is received
func (s *IpcServer) AddCallback(method string, callback func(message []byte)) {
	s.Callbacks[method] = callback
}

// readLoop reads messages on the UCP port and triggers callback accordingly
func (s *IpcServer) readLoop() {
	for {
		buffer := make([]byte, 1024)
		length, _, err := s.server.ReadFromUDP(buffer)
		if err != nil {
			continue
		}
		message := buffer[:length]
		var ipcCall ipcMessage
		err = json.Unmarshal(message, &ipcCall)
		if err != nil {
			continue
		}
		if callback, found := s.Callbacks[ipcCall.Method]; found {
			go callback(ipcCall.Message)
		}
	}
}
