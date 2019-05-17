package raft

import (
	"bytes"
	"encoding/gob"
)

func (kv *KV) needSnapshot() bool {
	if kv.snapshotThreshold < 0 {
		return false
	}
	if kv.snapshotThreshold < kv.store.StateSize() {
		return true
	}
	return false
}

func (kv *KV) Store(index int) {
	kv.snapshotIndex = index
	buf := new(bytes.Buffer)
	e := gob.NewEncoder(buf)
	_ = e.Encode(kv.m)
	_ = e.Encode(kv.snapshotIndex)
	_ = kv.store.StoreSnapshot(buf.Bytes())
}

func (kv *KV) LoadFromData(data []byte) {
	if data == nil || len(data) == 0 {
		return
	}
	d := gob.NewDecoder(bytes.NewBuffer(data))

	kv.m = make(map[string]string)

	_ = d.Decode(&kv.m)
	_ = d.Decode(&kv.snapshotIndex)
}

func (kv *KV) Load() {
	data, _ := kv.store.LoadSnapshot()
	kv.LoadFromData(data)
}
