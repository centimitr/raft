package raft

import (
	"net"
	"net/http"
	"net/rpc"
)

type Connectivity struct {
	listener net.Listener
}

func (c *Connectivity) Port() int {
	if c.listener == nil {
		return 0
	}
	return c.listener.Addr().(*net.TCPAddr).Port
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
	return
}

type ConnectInfo struct {
	Addrs []string
	Index int
}
