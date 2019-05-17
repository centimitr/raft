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
		Action(":registry :service", kv).
		Run()
}

func kv(c *scli.Context) {
	var err error

	// random ID for registry
	id := uuid.New().String()
	log.Println("id:", id)

	// new connectivity and RPC delegate
	connectivity := new(raft.Connectivity)
	delegate := new(raft.RPCDelegate)

	// publish RPC service
	err = connectivity.ListenAndServe("Raft", delegate, "")
	check(err)
	rpcAddr := fmt.Sprintf("localhost:%d", connectivity.Port())
	log.Println("rpcAddr:", rpcAddr)

	// register to registry
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

	// check registry for ConnectInfo
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

	// get peers
	peers := make([]raft.Peer, len(connectInfo.Addrs))
	for i, addr := range connectInfo.Addrs {
		if i == connectInfo.Index {
			continue
		}
		peers[i], err = rpc.DialHTTP("tcp", addr)
		check(err)
	}

	//kv
	//store := raft.MakePersister()
	store := new(raft.Archive)
	kv := raft.NewKV(peers, connectInfo.Index, store, 1<<31)
	delegate.Raft = kv.Raft

	time.Sleep(3 * time.Second)
	kv.Run()
	time.Sleep(3 * time.Second)

	//var v string

	//v, err = kv.Get("a")
	//log.Println(err)
	//log.Println("v:", v)

	for {
		log.Println("IsLeader:", kv.Raft.IsLeader())

		//err = kv.Set("a", "X")
		//log.Println(err)
		//
		time.Sleep(5 * time.Second)

		//v, err = kv.Get("a")
		//log.Println(err)
		//log.Println("v:", v)
	}

	// service
	//r := gin.New()
	//
	//addr := c.Get("service")
	//log.Println("running:", addr)
	//check(r.Run(addr))
}
