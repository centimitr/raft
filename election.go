package raft

import (
	"sync"
	"time"
)

var (
	ElectionTimeout = 2 * time.Second
	VotingTimeout   = 2 * time.Second
)

type Voting struct {
	Done    chan struct{}
	Timeout chan struct{}
	Cancel  chan struct{}
}

func (v *Voting) Request() {
	v.Done = make(chan struct{})
	v.Timeout = make(chan struct{})
	v.Cancel = make(chan struct{})
	time.AfterFunc(VotingTimeout, func() {
		close(v.Timeout)
	})

	var wg sync.WaitGroup
	// request votes
	wg.Wait()
	close(v.Done)
}

func (v *Voting) Win() bool {
	return false
}

type Election struct {
	state      *State
	Timer      *time.Timer
	Processing bool
	voting     *Voting
}

func NewElection(state *State) *Election {
	return &Election{
		state: state,
		Timer: time.NewTimer(ElectionTimeout),
	}
}

func (e *Election) ResetTimer() {
	e.Timer.Reset(ElectionTimeout)
}

func (e *Election) Abandon() {
	close(e.voting.Cancel)
}

func (e *Election) requestVotes() {

}

func (e *Election) announceBeingLeader() {
	var wg sync.WaitGroup
	// notify all
	wg.Wait()
}

func randomElectionInterval() time.Duration {
	// 150 - 300
	return 250 * time.Millisecond
}

func (e *Election) Start() {
	e.Processing = true
	log("elect!")
	e.state.CurrentTerm++
	e.state.Role = Candidate

	v := new(Voting)
	e.voting = v
	go v.Request()
	select {
	case <-v.Done:
		if v.Win() {
			e.state.Role = Leader
			go e.announceBeingLeader()
		}
	case <-v.Timeout:
		time.Sleep(randomElectionInterval())
		e.Start()
	case <-v.Cancel:
		e.state.Role = Follower
	}
	e.Processing = false
}
