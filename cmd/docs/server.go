package main

import (
	"encoding/json"
	"fmt"
	"github.com/devbycm/raft"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

// newServer creates the server serve frontend clients
func newServer(kv *raft.KV) *gin.Engine {
	r := gin.New()
	r.Use(cors.Default())
	r.Use(checkLeader(kv))
	r.GET("/status", func(c *gin.Context) {
		c.Status(http.StatusOK)
		return
	})
	r.GET("/doc", getDoc(kv))
	r.PUT("/doc", putDoc(kv))
	return r
}

// checkLeader is a middleware rejects requests that make not on a leader
func checkLeader(kv *raft.KV) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !kv.Raft.IsLeader() {
			c.String(http.StatusForbidden, fmt.Sprint(raft.ErrNotLeader))
			c.Abort()
			return
		}
	}
}

const KEY = "doc"

// returns latest doc to client
func getDoc(kv *raft.KV) gin.HandlerFunc {
	return func(c *gin.Context) {
		s, _ := kv.Get(KEY)
		var doc Doc
		_ = json.Unmarshal([]byte(s), &doc)
		c.JSON(http.StatusOK, doc)
		return
	}
}

// merge clients' docs to the kv's then send back
func putDoc(kv *raft.KV) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("put")
		// read doc from kv
		s, _ := kv.Get(KEY)
		doc := NewDoc(s)

		// read user doc
		var userDoc Doc
		err := c.BindJSON(&userDoc)

		if err == nil {
			_ = kv.Set(KEY, userDoc.String())
			c.JSON(http.StatusOK, userDoc)
		} else {
			c.JSON(http.StatusOK, doc)
		}
		return
	}
}
