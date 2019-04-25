package main

import (
	"fmt"
	"github.com/devbycm/raft"
	"github.com/devbycm/ssdr"
	"log"
	"net/http"
)

//func onPeersUpdate(services ssdr.ServiceListValue) {
//	log.Println(services)
//}

func onPeerConnectError(err error) {
	log.Println(err)
}

func peersUpdate(addrs *[]string) (r *ssdr.RegistryClient, err error) {
	r = ssdr.NewRegistryClient("ws://localhost:5000")
	err = r.QuickSubscribe("raft")
	if err != nil {
		return
	}
	var svl ssdr.ServiceListValue
	svl = <-r.ServiceListInitial
	*addrs = svl.Get("raft")
	go func() {
		for svl = range r.ServiceListUpdate {
			*addrs = svl.Get("raft")
		}
	}()
	return
}

func main() {
	var peerAddrs []string
	_, err := peersUpdate(&peerAddrs)
	if check(err, "service discovery") {
		return
	}
	fmt.Println(peerAddrs)

	r := raft.New(raft.Config{})
	err = r.SetupConnectivity(peerAddrs, onPeerConnectError)
	if check(err, "connect peers") {
		return
	}

	sm := new(StateMachine)
	r.BindStateMachine(sm)
	go r.Run()

	err = http.ListenAndServe(":0", nil)
	check(err)
}
