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
	log.Println("id:", id)

	// Connectivity wraps a RPC listener
	// RPC delegate is published to RPC registry
	connectivity := new(raft.Connectivity)
	delegate := new(raft.RPCDelegate)

	// publish RPC service
	err = connectivity.ListenAndServe("Raft", delegate, "")
	check(err)

	// retrieve the actual RPC address
	rpcAddr := fmt.Sprintf("localhost:%d", connectivity.Port())
	log.Println("rpcAddr:", rpcAddr)

	// call registry to record this RPC node
	u := &url.URL{
		Scheme: "http",
		Host:   "localhost" + c.Get("registry"),
	}
	querystring := u.Query()
	querystring.Add("id", id)
	querystring.Add("addr", rpcAddr)
	u.RawQuery = querystring.Encode()

	log.Println("GET", u.String())
	req, err := http.NewRequest(http.MethodPut, u.String(), nil)
	check(err)
	_, err = http.DefaultClient.Do(req)
	check(err)

	// poll registry to ensure that at least 5 raft nodes are ready
	// retrieve the addresses of other nodes
	var connectInfo raft.ConnectInfo
	for {
		time.Sleep(time.Second)
		log.Println("PUT", u.String())
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
			log.Println("waiting:", len(connectInfo.Addrs))
			continue
		}
		break
	}
	log.Printf("RECV %+v\n", connectInfo)
	log.Println("current:", connectInfo.Addrs[connectInfo.Index])

	// connect others nodes
	peers := make([]raft.Peer, len(connectInfo.Addrs))
	for i, addr := range connectInfo.Addrs {
		if i == connectInfo.Index {
			continue
		}
		peers[i], err = rpc.DialHTTP("tcp", addr)
		check(err)
	}

	// create kv service as a state machine for application logic
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
	log.Println("running:", addr)
	check(s.Run(addr))
}
