package raft

import (
	"math/rand"
	"sync"
	"time"
)

var (
	ElectionTimeout = 2 * time.Second
	VotingTimeout   = 2 * time.Second
)

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

// randomElectionInterval returns a time duration between 150~300ms
func randomElectionInterval() time.Duration {
	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)
	return time.Duration(150+r.Intn(150)) * time.Millisecond
}

func (e *Election) Start() {
	log("elect!")
	e.Processing = true
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
