package main

import (
	"fmt"
	"github.com/devbycm/raft"
	"github.com/devbycm/ssdr"
	"time"
)

func register(svcAddr string, id string, peersUpdate chan<- []*raft.Peer) (err error) {
	r := ssdr.NewRegistryClient("ws://localhost:5000")
	err = r.QuickSubscribe("raft", id, svcAddr)
	if err != nil {
		return
	}
	go func() {
		for svl := range r.ServiceListUpdate {
			nodes := svl.Get("raft", id)
			peers := make([]*raft.Peer, len(nodes))
			for i, n := range nodes {
				peers[i] = raft.NewPeer(n.Id, n.Addr)
			}
			peersUpdate <- peers
		}
	}()
	return
}

type model struct {
	*raft.KV
}

func newModel() *model {
	return &model{
		KV: new(raft.KV),
	}
}

func app() (err error) {
	m := newModel()
	_ = m

	// raft: bind state machine
	r := raft.New(raft.Config{})
	r.BindStateMachine(m)

	err = r.Start()
	if check(err, "start") {
		return
	}

	// DNS: register addr proxy
	// todo: DNS
	raftAddr := PortString(r.Connectivity.Port())

	// service discovery
	err = register(raftAddr, r.Id.String(), r.Connectivity.PeersUpdate)
	if check(err, "service discovery") {
		return
	}

	a := false
	a = true
	a = false
	go func() {
		time.Sleep(1 * time.Second)
		for {
			if a {
				time.Sleep(time.Second)
				now := time.Now().String()
				fmt.Println("set:", now)
				_ = m.Set("time", now)
				fmt.Println("SET SUCCESSFULLY")
			} else {
				v, _ := m.Get("time")
				fmt.Println("get:", v)
			}
			time.Sleep(3 * time.Second)
		}
	}()

	<-r.Quit
	//
	//upgrader := websocket.Upgrader{}
	//upgrader.CheckOrigin = func(r *http.Request) bool {
	//	return true
	//}
	//
	//var conn *websocket.Conn
	//go func() {
	//	for {
	//		if conn != nil {
	//			_ = conn.WriteJSON(time.Now().UnixNano())
	//		}
	//		time.Sleep(time.Second)
	//	}
	//}()
	//
	//s := gin.New()
	//s.NoRoute(func(c *gin.Context) {
	//	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	//	fmt.Println(err)
	//})
	//_ = s.Run(":3000")
	//
	//check(err)
	return
}
