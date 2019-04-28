package main

import (
	"github.com/devbycm/raft"
	"github.com/devbycm/ssdr"
)

func register(svcAddr string, id string, peersUpdate chan<- []string) (err error) {
	r := ssdr.NewRegistryClient("ws://localhost:5000")
	err = r.QuickSubscribe("raft", id, svcAddr)
	if err != nil {
		return
	}
	go func() {
		for svl := range r.ServiceListUpdate {
			peersUpdate <- svl.GetAddrs("raft", id)
		}
	}()
	return
}

func app(kv raft.StateMachine) (err error) {
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

	<-r.Quit
	//r.OnRoleChange = func() {
	//	srp.NewServer("", s.Addr())
	//}

	// service: listen
	//err = http.ListenAndServe(":0", nil)
	//check(err)
	return
}
