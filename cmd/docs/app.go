package main

import (
	"encoding/json"
	"fmt"
	"github.com/devbycm/raft"
	"github.com/devbycm/scli"
	"github.com/google/uuid"
	"io/ioutil"
	"log"
	"net/http"
	"net/rpc"
	"net/url"
	"time"
)

func main() {
	new(scli.App).
		Action(":registry :service", start).
		Run()
}

func start(c *scli.Context) {
	var err error

	// random ID for publishing raft service
	id := uuid.New().String()
	//log.Println("id:", id)

	// Connectivity wraps a RPC listener
	// RPC delegate is published to RPC registry
	connectivity := new(raft.Connectivity)
	delegate := new(raft.RPCDelegate)

	// publish RPC service
	err = connectivity.ListenAndServe("Raft", delegate, "")
	check(err)

	// retrieve the actual RPC address
	rpcAddr := fmt.Sprintf("%s:%d", GetLocalIP(), connectivity.Port())
	log.Println("[RPC]", rpcAddr)

	// call registry to record this RPC node
	u := &url.URL{
		Scheme: "http",
		Host:   c.Get("registry"),
	}
	querystring := u.Query()
	querystring.Add("id", id)
	querystring.Add("addr", rpcAddr)
	u.RawQuery = querystring.Encode()

	//log.Println("GET", u.String())
	req, err := http.NewRequest(http.MethodPut, u.String(), nil)
	check(err)
	_, err = http.DefaultClient.Do(req)
	check(err)

	// poll registry to ensure that at least 5 raft nodes are ready
	// retrieve the addresses of other nodes
	var connectInfo raft.ConnectInfo
	for {
		time.Sleep(time.Second)
		//log.Println("PUT", u.String())
		req, err = http.NewRequest(http.MethodGet, u.String(), nil)
		check(err)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			continue
		}
		b, err := ioutil.ReadAll(resp.Body)
		check(err)
		err = json.Unmarshal(b, &connectInfo)
		check(err)
		if resp.StatusCode == http.StatusServiceUnavailable {
			log.Println("[Registry] Total nodes count:", len(connectInfo.Addrs))
			continue
		}
		break
	}
	log.Printf("[Registry] RECV %+v\n", connectInfo)
	log.Printf("[Registry] Local: [%d] %s\n", connectInfo.Index, connectInfo.Addrs[connectInfo.Index])

	// connect others nodes
	peers := make([]raft.Peer, len(connectInfo.Addrs))
	for i, addr := range connectInfo.Addrs {
		if i == connectInfo.Index {
			continue
		}
		peers[i], err = rpc.DialHTTP("tcp", addr)
		check(err)
	}
	log.Println("[RPC] All peers connected")

	// create kv service as a state machine for application logic
	log.Println("[Server] Create Raft state machine")
	store := new(raft.Archive)
	kv := raft.NewKV(peers, connectInfo.Index, store, 1<<31)

	delegate.Raft = kv.Raft
	// todo: rafactor this timeout by calling registry to block
	// wait RPC delegates on other nodes have been set
	time.Sleep(time.Second)
	kv.Run()

	// create server and listen
	s := newServer(kv)
	addr := c.Get("service")
	log.Println("[Server] Listen and serve on:", addr)
	check(s.Run(addr))
}
