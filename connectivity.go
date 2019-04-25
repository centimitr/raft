package raft

import (
	"net"
	"net/http"
	"net/rpc"
)

type Connectivity struct {
	Peers    []*rpc.Client
	listener net.Listener
}

func NewConnectivity() *Connectivity {
	return new(Connectivity)
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

func (c *Connectivity) ConnectPeer(addr string) (err error) {
	client, err := rpc.DialHTTP("tcp", addr)
	if err != nil {
		return
	}
	c.Peers = append(c.Peers, client)
	return
}

func (c *Connectivity) ConnectPeers(addrs []string, onError OnError) {
	for _, addr := range addrs {
		err := c.ConnectPeer(addr)
		onError.Check(err)
	}
	return
}

//func (c *Connectivity) ForEachPeer(fn func(client *rpc.Client) error, onError OnError) {
//	for _, p := range c.Peers {
//		err := fn(p)
//		onError.Check(err)
//	}
//}
