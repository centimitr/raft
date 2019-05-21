package main

import (
	"github.com/devbycm/raft"
	"github.com/devbycm/scli"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
)

const (
	ClusterSize = 5
	//ClusterSize = 3
)

func main() {
	new(scli.App).
		Action(":addr", registry).
		Run()
}

func registry(c *scli.Context) {
	m := make(map[string]string)
	mu := sync.RWMutex{}

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	r.PUT("", func(c *gin.Context) {
		id := c.Query("id")
		addr := c.Query("addr")
		mu.Lock()
		m[id] = addr
		mu.Unlock()
		log.Printf("new: %s -> %s\n", id, addr)
		c.String(http.StatusOK, "")
	})

	r.GET("", func(c *gin.Context) {
		id := c.Query("id")
		resp := raft.ConnectInfo{
			Index: -1,
		}
		mu.RLock()
		size := len(m)
		currentAddr := m[id]
		//i := 0
		var addrs []string
		for _, addr := range m {
			addrs = append(addrs, addr)
		}
		sort.Strings(addrs)
		for i, addr := range addrs {
			if currentAddr == addr {
				resp.Index = i
			}
		}
		resp.Addrs = addrs

		mu.RUnlock()
		if resp.Index == -1 || size < ClusterSize {
			c.JSON(http.StatusServiceUnavailable, resp)
			return
		}
		log.Printf("return: %s\n", id)
		log.Printf(strings.Join(resp.Addrs, ", "))
		c.JSON(http.StatusOK, resp)
	})

	addr := c.Get("addr")
	log.Println("local:", GetLocalIP())
	log.Println("running:", addr)
	log.Fatalln(r.Run(addr))
}
