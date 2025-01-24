package swarmipc

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sort"
	"strconv"
	"sync"
	"time"
)

// IpcClient is a Docker Swarm oriented IPC manager
type IpcClient struct {
	taskName   string
	localIps   []net.IP
	servers    []net.IP
	port       string
	serverLock sync.RWMutex
}

// NewIpcClient Creates a new IPC client on the given UDP port
// taskName is the service name in Docker Swarm
func NewIpcClient(taskName string, port int) *IpcClient {
	c := &IpcClient{
		taskName: taskName,
		port:     strconv.Itoa(port),
	}
	c.updateLocalIp()
	go c.updateServerList()
	return c
}

// updateLocalIp Gets a local list of IPs in the IPC manager
func (c *IpcClient) updateLocalIp() {
	// Iterate network interfaces and get non loop-back/multicast IPs
	ifaces, err := net.Interfaces()
	if err != nil {
		return
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			default:
				continue
			}
			if ip.IsLoopback() || ip.IsMulticast() {
				continue
			}
			c.localIps = append(c.localIps, ip)
		}
	}
}

// updateServerList requests the list of replicas for the configured Swarm service name
// It does that by requesting A records from tasks.<service_name>
func (c *IpcClient) updateServerList() {
	for {
		addrs, err := net.LookupIP(fmt.Sprintf("tasks.%s", c.taskName))
		if err != nil {
			time.Sleep(10 * time.Second)
			continue
		}
		c.serverLock.Lock()
		c.servers = addrs
		c.serverLock.Unlock()
		time.Sleep(10 * time.Second)
	}

}

// AmIMaster sorts all servers IPs and checks if we are the first one
// Can be used so only 1 replica evaluates to true
func (c *IpcClient) AmIMaster() bool {
	if len(c.localIps) < 0 {
		return false
	}
	if len(c.servers) < 1 {
		return false
	}
	c.serverLock.RLock()
	localServerList := make([]net.IP, len(c.servers))
	copy(localServerList, c.servers)
	c.serverLock.RUnlock()
	var ipList []string
	for _, ip := range localServerList {
		ipList = append(ipList, ip.String())
	}
	sort.Sort(sort.StringSlice(ipList))
	for _, ip := range c.localIps {
		if ip.String() == ipList[0] {
			return true
		}
	}
	return false
}

// CallBroadcast calls a method on all replicas
func (c *IpcClient) CallBroadcast(method string, message []byte) error {
	var lastError error
	c.serverLock.RLock()
	localServerList := make([]net.IP, len(c.servers))
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

// Call calls a method on a specific replica
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

// ipcMessage Message structure for IPC communication
type ipcMessage struct {
	Method  string `json:"m"`
	Message []byte `json:"d"`
}
