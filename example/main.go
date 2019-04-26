package main

import (
	"fmt"
	"github.com/devbycm/raft"
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

func main() {
	app()
}
