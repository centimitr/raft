package raft

import (
	"fmt"
	"net"
	"net/http"
	"net/rpc"
)

type PeerId = string

type Peer struct {
	Id   PeerId
	Addr string
	*rpc.Client
	Meta map[string]interface{}
}

func NewPeer(id string, addr string) *Peer {
	return &Peer{Id: id, Addr: addr, Meta: make(map[string]interface{})}
}

type Connectivity struct {
	listener    net.Listener
	PeersUpdate chan []*Peer
	Peers       map[PeerId]*Peer
	OnError     OnError
}

func NewConnectivity() *Connectivity {
	return &Connectivity{
		Peers:       make(map[PeerId]*Peer),
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
	_, ok := c.Peers[peer.Id]
	if ok {
		return true
	}
	return false
}

func (c *Connectivity) ConnectPeer(peer *Peer) (err error) {
	fmt.Println(peer.Id)
	if c.HasConnectedPeer(peer) {
		return
	}
	peer.Client, err = rpc.DialHTTP("tcp", peer.Addr)
	if err != nil {
		c.OnError.Check(err)
		return
	}
	c.Peers[peer.Id] = peer
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
