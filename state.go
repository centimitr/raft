package raft

import (
	"sync"
	"time"
)

type NodeIndex = int
type Role = int
type Term = int
type LogEntryIndex = int

const (
	Follower Role = iota
	Candidate
	Leader
)

type Peer interface {
	Call(method string, arg interface{}, reply interface{}) error
}

// Config
type Config struct {
	ElectionTimeout  time.Duration
	HeartbeatTimeout time.Duration
}

// Raft
type Raft struct {
	Config

	peers      []Peer
	peersCount int
	StableStore

	Id          NodeIndex
	Role        Role
	CurrentTerm Term

	apply chan ApplyMsg

	Log
	progress
	election

	shutdown chan struct{}
	mu       sync.Mutex
}

func (r *Raft) init() {
	r.Role = Follower
	r.VotedFor = -1

	r.Log.init(&r.mu)
	r.progress.init(r.peersCount)
	r.election.init(r.Config.ElectionTimeout)

	r.shutdown = make(chan struct{})
}

// forEachPeer does not include the node itself
func (r *Raft) forEachPeer(fn func(peerIndex NodeIndex)) {
	for i := 0; i < r.peersCount; i++ {
		if i != r.Id {
			fn(i)
		}
	}
}

func (r *Raft) IsLeader() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.Role == Leader
}

func (r *Raft) CheckLeadership(term Term) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.Role == Leader && r.CurrentTerm == term
}

func (r *Raft) becomeFollower() {
	r.Role = Follower
	r.VotedFor = -1
}
