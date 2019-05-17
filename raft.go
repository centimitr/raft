package raft

import (
	"errors"
	"math/rand"
	"time"
)

var (
	ErrNotLeader = errors.New("kv: current node not a leader")
)

func NewRaft(peers []Peer, me NodeIndex, store StableStore, applyCh chan ApplyMsg) *Raft {
	r := &Raft{
		Config: Config{
			ElectionTimeout:  time.Millisecond * time.Duration(500+rand.Intn(100)*5),
			HeartbeatTimeout: 50 * time.Millisecond,
		},
		peers:       peers,
		peersCount:  len(peers),
		StableStore: store,
		Id:          me,
		apply:       applyCh,
	}
	r.init()

	r.Restore()
	r.Log.LastApplied = r.Snapshot.LastIncludedIndex
	r.Log.CommitIndex = r.Snapshot.LastIncludedIndex

	r.log("up: term: %d", r.CurrentTerm)
	return r
}

func (r *Raft) Run() {
	go r.electionLoop()
	go r.applyLoop()
}

func (r *Raft) Shutdown() {
	close(r.shutdown)
	r.commitCond.Broadcast()
}

type ApplyReceipt struct {
	Term  Term
	Index LogEntryIndex
	Err   error
}

func (r *Raft) Apply(command interface{}) (receipt ApplyReceipt) {
	receipt.Index = -1
	select {
	case <-r.shutdown:
		return
	default:
		r.mu.Lock()
		defer r.mu.Unlock()

		if r.Role != Leader {
			receipt.Err = ErrNotLeader
			return
		}

		r.Log.append(r.CurrentTerm, command)

		lastIndex := r.Log.lastIndex()
		r.log("append: {%d: %v}", lastIndex, command)

		receipt.Index = lastIndex
		receipt.Term = r.CurrentTerm

		r.progress.update(r.Id, lastIndex)
		r.Store()
	}
	return
}
