package raft

import (
	"encoding/gob"
	"math"
	"sync"
)

type KV struct {
	currentIndex NodeIndex
	Raft         *Raft
	applyCh      chan ApplyMsg

	m  map[string]string
	mu sync.Mutex

	store             StableStore
	snapshotThreshold int
	snapshotIndex     int

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

func (kv *KV) Run() {
	if kv.Raft == nil {
		panic("kv: no raft binding")
	}
	go kv.loop()
	kv.Raft.Run()
}

func (kv *KV) Shutdown() {
	close(kv.shutdown)
	kv.Raft.Shutdown()
}

func NewKV(peers []Peer, currentIndex NodeIndex, store StableStore, snapshotThreshold int) *KV {
	kv := &KV{
		currentIndex:      currentIndex,
		snapshotThreshold: int(math.MaxInt64),
		store:             store,
	}
	kv.Init()
	kv.Load()
	kv.Raft = NewRaft(peers, currentIndex, store, kv.applyCh)
	return kv
}
