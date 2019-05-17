package main

import (
	"fmt"
	"github.com/devbycm/raft"
	"github.com/gin-gonic/gin"
	"net/http"
)

const KEY = "doc"

func checkLeader(kv *raft.KV) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !kv.Raft.IsLeader() {
			c.String(http.StatusForbidden, fmt.Sprint(raft.ErrNotLeader))
			c.Abort()
		}
	}
}

func newServer(kv *raft.KV) *gin.Engine {
	r := gin.New()

	// reject requests that make not on a leader
	r.Use(checkLeader(kv))

	// returns latest doc to client
	r.GET("/doc", func(c *gin.Context) {
		s, _ := kv.Get(KEY)
		c.String(http.StatusOK, s)
	})

	// merge clients' docs to the kv's then send back
	r.PUT("/doc", func(c *gin.Context) {
		// read doc from kv
		s, _ := kv.Get(KEY)
		doc := NewDoc(s)

		// read user doc
		var userDoc Doc
		_ = c.BindJSON(&userDoc)

		// merge
		doc.Merge(userDoc)

		// update doc from kv
		_ = kv.Set(KEY, doc.String())

		c.JSON(http.StatusOK, doc)
	})
	return r
}
