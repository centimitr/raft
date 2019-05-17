package raft

import (
	"encoding/gob"
	"sync"
)

type KV struct {
	mu sync.Mutex

	currentIndex NodeIndex
	raft         *Raft
	applyCh      chan ApplyMsg

	snapshotThreshold int // snapshot if log grows this big

	store         StableStore
	m             map[string]string
	snapshotIndex int

	notify   map[int]chan struct{}
	shutdown chan struct{}
}

func (kv *KV) Init() {
	gob.Register(cmd{})
	kv.applyCh = make(chan ApplyMsg)
	kv.m = make(map[string]string)
	kv.notify = make(map[int]chan struct{})
	kv.shutdown = make(chan struct{})
}

func (kv *KV) Shutdown() {
	close(kv.shutdown)
	kv.raft.Shutdown()
}

func StartKV(peers []Peer, currentIndex NodeIndex, store StableStore, snapshotThreshold int) *KV {
	kv := &KV{
		currentIndex:      currentIndex,
		snapshotThreshold: snapshotThreshold,
		store:             store,
	}
	kv.Init()
	kv.Load()
	kv.raft = NewRaft(peers, currentIndex, store, kv.applyCh)
	go kv.loop()
	kv.raft.Run()
	return kv
}
