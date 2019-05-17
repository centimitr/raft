package raft

import (
	"sync"
)

// simulated stable StableStore
type Archive struct {
	// should replace with IO
	data sync.Map
	mu   sync.Mutex
}

func (a *Archive) load(key string) ([]byte, error) {
	v, _ := a.data.LoadOrStore(key, []byte{})
	return v.([]byte), nil
}

func (a *Archive) store(key string, data []byte) error {
	a.data.Store(key, data)
	return nil
}

func (a *Archive) LoadStatus() ([]byte, error) {
	return a.load("status")
}

func (a *Archive) LoadSnapshot() ([]byte, error) {
	return a.load("snapshot")
}

func (a *Archive) StoreStatus(data []byte) error {
	return a.store("status", data)
}

func (a *Archive) StoreSnapshot(data []byte) error {
	return a.store("snapshot", data)
}

func (a *Archive) StateSize() int {
	state, _ := a.LoadStatus()
	return len(state)
}

//type StableStore interface {
//	LoadStatus() ([]byte, error)
//	LoadSnapshot() ([]byte, error)
//	StoreStatus([]byte) error
//	StoreSnapshot([]byte) error
//	StateSize() int
//}
