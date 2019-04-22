package raft

import (
	"sync"
	"time"
)

var (
	ElectionTimeout = 2 * time.Second
)

type Election struct {
	State        *State
	Timer        *time.Timer
	Processing   bool
	votingDone   chan struct{}
	votingCancel chan struct{}
}

func NewElection(state *State) *Election {
	return &Election{
		State: state,
		Timer: time.NewTimer(ElectionTimeout),
	}
}

func (e *Election) ResetTimer() {
	e.Timer.Reset(ElectionTimeout)
}

func (e *Election) Abandon() {
	close(e.votingCancel)
}

func (e *Election) requestVotes() {
	var wg sync.WaitGroup
	// request votes
	wg.Wait()
	close(e.votingDone)
}

func (e *Election) announceBeingLeader() {
	var wg sync.WaitGroup
	// notify all
	wg.Wait()
}

func (e *Election) Start() {
	e.votingDone = make(chan struct{})
	e.votingCancel = make(chan struct{})

	e.Processing = true
	log("elect!")
	e.State.CurrentTerm++
	e.State.Role = Candidate
	go e.requestVotes()
	select {
	case <-e.votingDone:
		if true {
			e.State.Role = Leader
			go e.announceBeingLeader()
		}
	case <-e.votingCancel:
		e.State.Role = Follower
	}
	e.Processing = false
}
