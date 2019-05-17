package main

import (
	"github.com/devbycm/raft"
	"github.com/gin-gonic/gin"
	"net/http"
)

func newServer(kv *raft.KV) *gin.Engine {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		if !kv.Raft.IsLeader() {
			c.String(http.StatusForbidden, "raft: current node not a leader")
		}
	})
	r.GET("doc", func(c *gin.Context) {
		println("get doc")
		c.JSON()
	})
	r.PUT("doc", func(c *gin.Context) {
		println("put doc")
	})
	return r
}
