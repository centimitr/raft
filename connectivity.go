package raft

import (
	"net"
	"net/http"
	"net/rpc"
)

type Peer struct {
	Id   string
	Addr string
	*rpc.Client
}

func NewPeer(id string, addr string) *Peer {
	return &Peer{Id: id, Addr: addr}
}

type Connectivity struct {
	Peers       []*Peer
	listener    net.Listener
	PeersUpdate chan []*Peer
	OnError     OnError
}

func NewConnectivity() *Connectivity {
	return &Connectivity{
		PeersUpdate: make(chan []*Peer),
	}
}

func (c *Connectivity) Port() int {
	if c.listener == nil {
		return 0
	}
	return c.listener.Addr().(*net.TCPAddr).Port
}

func (c *Connectivity) handleUpdates() {
	for peers := range c.PeersUpdate {
		c.ConnectPeers(peers)
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

func (c *Connectivity) HasConnectedPeer(peer *Peer) bool {
	for _, p := range c.Peers {
		if p.Id == peer.Id {
			return true
		}
	}
	return false
}

func (c *Connectivity) ConnectPeer(peer *Peer) (err error) {
	if c.HasConnectedPeer(peer) {
		return
	}
	peer.Client, err = rpc.DialHTTP("tcp", peer.Addr)
	if err != nil {
		c.OnError.Check(err)
		return
	}
	c.Peers = append(c.Peers, peer)
	return
}

func (c *Connectivity) ConnectPeers(peers []*Peer) {
	for _, peer := range peers {
		err := c.ConnectPeer(peer)
		// todo: remove debug
		log("connect:", peer.Addr, err)
		c.OnError.Check(err)
	}
	return
}
