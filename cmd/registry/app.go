package main

import (
	"github.com/devbycm/raft"
	"github.com/devbycm/scli"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"sync"
)

const (
	//CLUSTER_SIZE = 5
	CLUSTER_SIZE = 3
	//CLUSTER_SIZE = 1
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
		resp.Addrs = make([]string, size)
		i := 0
		for mid, addr := range m {
			resp.Addrs[i] = addr
			if mid == id {
				resp.Index = i
			}
			i++
		}
		mu.RUnlock()
		if resp.Index == -1 || size < CLUSTER_SIZE {
			c.JSON(http.StatusServiceUnavailable, resp)
			return
		}
		log.Printf("return: %s\n", id)
		c.JSON(http.StatusOK, resp)
	})

	addr := c.Get("addr")
	log.Println("local:", GetLocalIP())
	log.Println("running:", addr)
	log.Fatalln(r.Run(addr))
}
