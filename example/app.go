package main

import (
	"net/http"
	"raft"
)

func main() {
	go raft.New().Run()
	_ = http.ListenAndServe(":3000", nil)
}
