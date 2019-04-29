package main

import (
	"fmt"
	"github.com/devbycm/raft"
	"github.com/devbycm/ssdr"
	"log"
	"net"
	"net/http"
	"strings"
)

func check(err error, vs ...string) bool {
	if err != nil {
		if len(vs) > 0 {
			log.Println(strings.Join(vs, ": ")+":", err)
		} else {
			log.Println(err)
		}
		return true
	} else {
		return false
	}
}

func PortString(port int) string {
	return fmt.Sprintf(":%d", port)
}

type DefaultPortServer struct {
	ln net.Listener
}

func (s *DefaultPortServer) Port() int {
	if s.ln == nil {
		return 0
	}
	return s.ln.Addr().(*net.TCPAddr).Port
}

func (s *DefaultPortServer) Addr() string {
	return PortString(s.Port())
}

func (s *DefaultPortServer) Listen() (err error) {
	s.ln, err = net.Listen("tcp", ":0")
	return
}

func (s *DefaultPortServer) Serve(handler http.Handler) error {
	return http.Serve(s.ln, handler)
}

func ServiceToPeers(nodes []*ssdr.ServiceNode) []*raft.Peer {
	peers := make([]*raft.Peer, len(nodes))
	for i, n := range nodes {
		peers[i] = raft.NewPeer(n.Id, n.Addr)
	}
	return peers
}
