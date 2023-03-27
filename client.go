package swarmipc

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

type IpcClient struct {
	taskName   string
	myIp       net.IP
	servers    []net.IP
	port       string
	serverLock sync.RWMutex
}

func NewIpcClient(taskName string, port int) *IpcClient {
	myHostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	myAddr, _ := net.LookupIP(myHostname)

	c := &IpcClient{
		myIp:     myAddr[0],
		taskName: taskName,
		port:     strconv.Itoa(port),
	}
	go c.updateServerList()
	return c
}

func (c *IpcClient) updateServerList() {
	for {
		addrs, err := net.LookupIP(fmt.Sprintf("tasks.%s", c.taskName))
		if err != nil {
			log.Printf("no server found for task %s : %s", c.taskName, err)
		}
		c.serverLock.Lock()
		c.servers = addrs
		c.serverLock.Unlock()
		time.Sleep(10 * time.Second)
	}

}

func (c *IpcClient) CallBroadcast(method string, message []byte) error {
	var lastError error
	var localServerList []net.IP
	c.serverLock.RLock()
	copy(localServerList, c.servers)
	c.serverLock.RUnlock()
	for _, ip := range localServerList {
		err := c.Call(ip, method, message)
		if err != nil {
			lastError = err
			log.Println(err)
		}
	}
	return lastError
}

func (c *IpcClient) Call(ip net.IP, method string, message []byte) error {
	conn, err := net.Dial("udp", net.JoinHostPort(ip.String(), c.port))
	if err != nil {
		return err
	}
	if len(message) > 1024 {
		return errors.New("message too big")
	}
	ipc, err := json.Marshal(ipcMessage{
		Method:  method,
		Message: message,
	})
	if err != nil {
		return err
	}
	_, err = io.Copy(conn, bytes.NewReader(ipc))
	if err != nil {
		return err
	}
	return conn.Close()
}

type ipcMessage struct {
	Method  string `json:"m"`
	Message []byte `json:"d"`
}
