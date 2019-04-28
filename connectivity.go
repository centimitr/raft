package raft

import (
	"net"
	"net/http"
	"net/rpc"
)

type Connectivity struct {
	Peers       []*rpc.Client
	listener    net.Listener
	PeersUpdate chan []string
	OnError     OnError
}

func NewConnectivity() *Connectivity {
	return &Connectivity{
		PeersUpdate: make(chan []string),
	}
}

func (c *Connectivity) Port() int {
	if c.listener == nil {
		return 0
	}
	return c.listener.Addr().(*net.TCPAddr).Port
}

func (c *Connectivity) handleUpdates() {
	for addrs := range c.PeersUpdate {
		c.ConnectPeers(addrs)
	}
}

func (c *Connectivity) ListenAndServe(serviceName string, v interface{}, addr string) (err error) {
	if c.listener != nil {
		_ = c.listener.Close()
		c.listener = nil
	}
	s := rpc.NewServer()
	err = s.RegisterName(serviceName, v)
	if err != nil {
		return
	}
	c.listener, err = net.Listen("tcp", addr)
	if err != nil {
		return
	}
	go http.Serve(c.listener, s)
	go c.handleUpdates()
	return
}

func (c *Connectivity) ConnectPeer(addr string) (err error) {
	client, err := rpc.DialHTTP("tcp", addr)
	if err != nil {
		c.OnError.Check(err)
		return
	}
	c.Peers = append(c.Peers, client)
	return
}

func (c *Connectivity) ConnectPeers(addrs []string) {
	for _, addr := range addrs {
		err := c.ConnectPeer(addr)
		log("connect:", addr, err)
		c.OnError.Check(err)
	}
	return
}
