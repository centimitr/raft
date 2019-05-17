package raft

import (
	"sync"
	"time"
)

// type aliases help understand variables
type NodeIndex = int
type Role = int
type Term = int
type LogEntryIndex = int

const (
	Follower Role = iota
	Candidate
	Leader
)

// Peer is a general interface for RPC client endpoints
// It can be used to replace the original RPC with a testable one
type Peer interface {
	Call(method string, arg interface{}, reply interface{}) error
}

// Config stores config for raft
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

func (r *Raft) becomeFollower() {
	r.Role = Follower
	r.VotedFor = -1
}

func (r *Raft) init() {
	r.becomeFollower()

	r.Log.init(&r.mu)
	r.progress.init(r.peersCount)
	r.election.init(r.Config.ElectionTimeout)

	r.shutdown = make(chan struct{})
}

// forEachPeer iterate over peer nodes.
func (r *Raft) forEachPeer(fn func(peerIndex NodeIndex)) {
	for i := 0; i < r.peersCount; i++ {
		if i != r.Id {
			fn(i)
		}
	}
}

// IsLeader is concurrent safe to user by outer service.
func (r *Raft) IsLeader() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.Role == Leader
}

// CheckLeadership help outer service's async operations check leadership.
func (r *Raft) CheckLeadership(term Term) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.Role == Leader && r.CurrentTerm == term
}
