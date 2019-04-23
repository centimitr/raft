package main

import (
	"log"
	"net/http"
	"raft"
)

func onPeerConnectError(err error) {
	log.Println(err)
}

func main() {
	r := raft.New(raft.Config{})

	var peerAddrs []string
	err := r.SetupConnectivity(peerAddrs, onPeerConnectError)
	if check(err) {
		return
	}
	go r.Run()
	_ = http.ListenAndServe(":3000", nil)
}
