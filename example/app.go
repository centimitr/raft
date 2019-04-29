package main

import (
	"github.com/devbycm/raft"
	"github.com/devbycm/ssdr"
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

func app(kv *raft.KV) (err error) {
	// raft: bind state machine
	r := raft.New(raft.Config{})
	r.BindStateMachine(kv)

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

	//<-r.Quit
	//r.OnRoleChange = func() {
	//	srp.NewServer("", s.Addr())
	//}

	//s := gin.New()
	//s.GET("/add", func(context *gin.Context) {
	//
	//})
	//s.NoRoute(func(c *gin.Context) {
	//	v, _ := kv.GetDefault("cnt", 0)
	//	c.String(http.StatusOK, strconv.Itoa(v.(int)))
	//})
	//
	// service: listen
	//err = http.ListenAndServe(":3000", s)
	//check(err)
	return
}
