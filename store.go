package raft

import (
	"bytes"
	"encoding/gob"
)

func (r *Raft) Store() {
	buf := new(bytes.Buffer)
	e := gob.NewEncoder(buf)
	_ = e.Encode(r.CurrentTerm)
	_ = e.Encode(r.VotedFor)
	_ = e.Encode(r.Log.Logs)
	_ = e.Encode(r.Log.Snapshot.LastIncludedIndex)
	_ = e.Encode(r.Log.Snapshot.LastIncludedTerm)
	_ = r.store.StoreStatus(buf.Bytes())
}

func (r *Raft) Restore() {
	data, _ := r.store.LoadStatus()
	if data == nil || len(data) == 0 {
		return
	}
	buf := bytes.NewBuffer(data)
	d := gob.NewDecoder(buf)
	_ = d.Decode(&r.CurrentTerm)
	_ = d.Decode(&r.VotedFor)
	_ = d.Decode(&r.Log.Logs)
	_ = d.Decode(&r.Log.Snapshot.LastIncludedIndex)
	_ = d.Decode(&r.Log.Snapshot.LastIncludedTerm)
}
