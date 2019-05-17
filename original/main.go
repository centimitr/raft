package main

import (
	"fmt"
	"github.com/devbycm/raft"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
)

func tryKV() {
	kv := new(raft.KV)
	fmt.Println(kv.Get("users"))
	fmt.Println(kv.Set("users", []string{"1", "2", "3"}))
	fmt.Println(kv.Get("users"))
	fmt.Println(kv.Remove("users"))
	fmt.Println(kv.Get("users"))
	fmt.Println(kv.MustGet("users"))
	fmt.Println(kv.GetDefault("users", []string{"1", "2", "3"}))
	fmt.Println(kv.Get("users"))
}

var kv = new(raft.KV)

type ClientState struct {
	Meta struct {
		Title     string
		Arguments string
	}
}

type StateSyncPatch struct {
}

func test() {
	_ = kv
	upgrader := websocket.Upgrader{}
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}

	var conns = make(map[string]*websocket.Conn)

	r := gin.New()
	r.GET("/:user", func(c *gin.Context) {
		user := c.Param("user")
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		conns[user] = conn
		fmt.Println(user)
		go func(conn *websocket.Conn) {
			fmt.Println("start")
			for {
				var s ClientState
				err := conn.ReadJSON(&s)
				if err != nil {
					fmt.Println(err)
					continue
				}
				fmt.Println(s)
				var p StateSyncPatch
				_ = conn.WriteJSON(p)
			}
		}(conn)
	})
	_ = r.Run(":3000")
}

func main() {
	_ = app()
}
